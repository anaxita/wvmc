package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
	"os"
	"strings"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/model"
	"github.com/dgrijalva/jwt-go"
)

// Auth выполняет проверку токена
func (s *Server) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		tokenString := strings.Split(authHeader, " ")

		if len(tokenString) != 2 || tokenString[0] != "Bearer" {
			logit.Log("Нет токена в заголовке", authHeader)
			SendErr(w, http.StatusUnauthorized, errors.New("no token in headers"), "Нет токена в заголовке")
			return
		}

		token, err := jwt.ParseWithClaims(tokenString[1], &customClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("TOKEN")), nil
		})

		if err != nil {
			SendErr(w, http.StatusUnauthorized, err, "Токен истёк")
			return
		}

		if claims, ok := token.Claims.(*customClaims); ok && token.Valid {
			if claims.Type == "access" {
				ctxUser := CtxString("user")
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUser, claims.User)))
				return
			}

			SendErr(w, http.StatusUnauthorized, errors.New("token it not 'access'"), "Неверный тип токен")
			return
		}
		SendErr(w, http.StatusUnauthorized, err, "Токен не валиден")
	})
}

// Cors устанавливает cors заголовки
func (s *Server) Cors(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		logit.Info("REQUEST", r.Method, r.RemoteAddr, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// RefreshToken выполняет переиздание токена
func (s *Server) RefreshToken() http.Handler {
	type respTokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	type reqTokens struct {
		RefreshToken string `json:"refresh_token"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := reqTokens{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный формат запроса")
			return
		}

		// Парсим токен
		token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			// Проверяем, что метод авторизации верный
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Если все ок - возвращаем ключ подписи
			return []byte(os.Getenv("TOKEN")), nil
		})

		if err != nil {
			SendErr(w, http.StatusUnauthorized, err, "Ошибка проверки сигнатуры")
			return
		}

		// Проверяем корректность Claims (Данных внутри токена) и проверяем Валидность (не совсем понимаю что это значит) ключа подписи
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			tokenType, ok := claims["Type"]
			if ok && tokenType == "refresh" {

				u := model.User{}
				userjson, _ := json.Marshal(claims["User"])
				json.Unmarshal(userjson, &u)

				store := s.store.User(r.Context())
				err = store.GetRefreshToken(req.RefreshToken)

				if err != nil {
					SendErr(w, http.StatusUnauthorized, err, "Токен уже использовался")
					return
				}

				refreshToken := createToken("refresh", u)

				err = store.CreateRefreshToken(u.ID, refreshToken)
				if err != nil {
					SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
					return
				}

				tokens := respTokens{
					AccessToken:  createToken("access", u),
					RefreshToken: refreshToken,
				}

				SendOK(w, http.StatusOK, tokens)
				return
			}
			SendErr(w, http.StatusBadRequest, errors.New("token type is not refresh"), "Неверный тип токена")
			return
		}
		SendErr(w, http.StatusUnauthorized, errors.New("token is invalid"), "Токен невалидный")
	})
}

// CheckIsAdmin проверяет является ли пользователь админом
func (s *Server) CheckIsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adminRole := 1
		ctxUser := r.Context().Value(CtxString("user")).(model.User)

		logit.Info("Проверяем права пользователя", ctxUser.Email)

		if ctxUser.Role != adminRole {
			SendErr(w, http.StatusForbidden, errors.New("user is not admin"), "Пользователь не администратор")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CheckControlPermissions проверяет права на сервер у пользователя
func (s *Server) CheckControlPermissions(next http.Handler) http.Handler {
	type controlRequest struct {
		ServerID string `json:"server_id"`
		Command  string `json:"command"`
	}

	var adminRole = 1

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := controlRequest{}
		var err error
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, errors.New("server_id and command fields cannot be empty"), "server_id и command не могут быть пустыми")
			return
		}

		if req.ServerID == "" || req.Command == "" {
			SendErr(w, http.StatusBadRequest, errors.New("fields cannot be empty"), "Все поля должны быть заполнены")
			return
		}

		ctxUser := r.Context().Value(CtxString("user")).(model.User)

		server, err := s.store.Server(r.Context()).Find("id", req.ServerID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusBadRequest, err, "Сервер не найден")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		logit.Info("Проверяем права пользователя", ctxUser.Email)

		if ctxUser.Role != adminRole {
			serversByUser, err := s.store.Server(r.Context()).FindByUser(ctxUser.ID)
			if err != nil {
				if err == sql.ErrNoRows {
					SendErr(w, http.StatusBadRequest, err, "Сервер не найден")
					return
				}

				SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
				return
			}

			for _, srv := range serversByUser {
				if srv.ID == req.ServerID {
					ctxServer := CtxString("server")
					ctxCommand := CtxString("command")
					newctx := context.WithValue(r.Context(), ctxServer, srv)
					newctx = context.WithValue(newctx, ctxCommand, req.Command)
					next.ServeHTTP(w, r.WithContext(newctx))
					return
				}
			}

			SendErr(w, http.StatusForbidden, err, "Доступ запрещен")
			return
		}

		ctxServer := CtxString("server")
		ctxCommand := CtxString("command")
		newctx := context.WithValue(r.Context(), ctxServer, server)
		newctx = context.WithValue(newctx, ctxCommand, req.Command)
		next.ServeHTTP(w, r.WithContext(newctx))
	})
}
