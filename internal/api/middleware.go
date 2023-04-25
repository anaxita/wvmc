package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/gorilla/mux"

	"github.com/dgrijalva/jwt-go"
)

// Auth выполняет проверку токена
func (s *Server) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		tokenString := strings.Split(authHeader, " ")

		if len(tokenString) != 2 || tokenString[0] != "Bearer" {
			SendErr(w, http.StatusUnauthorized, errors.New("no token in headers"),
				"Нет токена в заголовке")
			return
		}

		token, err := jwt.ParseWithClaims(tokenString[1], &customClaims{},
			func(token *jwt.Token) (interface{}, error) {
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

				next.ServeHTTP(w,
					r.WithContext(context.WithValue(r.Context(), ctxUser, claims.User)))
				return
			}

			SendErr(w, http.StatusUnauthorized, errors.New("token it not 'access'"),
				"Неверный тип токен")
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

				u := entity.User{}
				userjson, _ := json.Marshal(claims["User"])
				json.Unmarshal(userjson, &u)

				err = s.authService.RefreshToken(r.Context(), req.RefreshToken)

				if err != nil {
					SendErr(w, http.StatusUnauthorized, err, "Токен уже использовался")
					return
				}

				refreshToken := createToken("refresh", u)

				err = s.authService.CreateRefreshToken(r.Context(), u.ID, refreshToken)
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
			SendErr(w, http.StatusBadRequest, errors.New("token type is not refresh"),
				"Неверный тип токена")
			return
		}
		SendErr(w, http.StatusUnauthorized, errors.New("token is invalid"), "Токен невалидный")
	})
}

// RoleMiddleware проверяет является ли пользователь админом
func (s *Server) RoleMiddleware(roles ...int) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxUser := r.Context().Value(CtxString("user")).(entity.User)

			for _, v := range roles {
				if ctxUser.Role == v {
					next.ServeHTTP(w, r)

					return
				}
			}

			SendErr(w, http.StatusForbidden, errors.New("user has no permissions"),
				"Недостаточно прав")
		})
	}
}

// CheckControlPermissions проверяет права на сервер у пользователя
func (s *Server) CheckControlPermissions(next http.Handler) http.Handler {
	type controlRequest struct {
		ServerID int64  `json:"server_id"`
		Command  string `json:"command"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req controlRequest
		var err error

		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest,
				errors.New("server_id and command fields cannot be empty"),
				"server_id и command не могут быть пустыми")

			return
		}

		if req.ServerID == 0 || req.Command == "" {
			SendErr(w, http.StatusBadRequest, errors.New("fields cannot be empty"),
				"Все поля должны быть заполнены")

			return
		}

		ctxUser := r.Context().Value(CtxString("user")).(entity.User)

		server, err := s.serverService.FindByID(r.Context(), req.ServerID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusBadRequest, err, "Сервер не найден")

				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")

			return
		}

		if ctxUser.Role != entity.UserRoleAdmin {
			serversByUser, err := s.serverService.FindByUserID(r.Context(), ctxUser.ID)
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
					newCtx := context.WithValue(r.Context(), ctxServer, srv)
					newCtx = context.WithValue(newCtx, ctxCommand, req.Command)

					next.ServeHTTP(w, r.WithContext(newCtx))

					return
				}
			}

			SendErr(w, http.StatusForbidden, err, "Доступ запрещен")

			return
		}

		ctxServer := CtxString("server")
		ctxCommand := CtxString("command")
		newCtx := context.WithValue(r.Context(), ctxServer, server)
		newCtx = context.WithValue(newCtx, ctxCommand, req.Command)

		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}
