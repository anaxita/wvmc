package server

import (
	"context"
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
			SendErr(w, http.StatusUnauthorized, errors.New("No token in headers"), "Нет токена в заголовке")
			return
		}

		token, err := jwt.ParseWithClaims(tokenString[1], &customClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
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
			SendErr(w, http.StatusUnauthorized, errors.New("Token it not 'access'"), "Неверный тип токен")
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
		AccessToken  string `json:"token"`
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
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok != true {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
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
			SendErr(w, http.StatusBadRequest, errors.New("Token type is not refresh"), "Неверный тип токена")
			return
		}
		SendErr(w, http.StatusUnauthorized, errors.New("Token is invalid"), "Токен невалидный")
	})
}

// CheckIsAdmin проверяет является ли пользователь админом
func (s *Server) CheckIsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var admin = 1
		ctxUser := r.Context().Value(CtxString("user")).(model.User)
		if ctxUser.Role != admin {
			SendErr(w, http.StatusForbidden, errors.New("User is not admin"), "Неверный формат запроса")
			return
		}
		next.ServeHTTP(w, r)
	})
}
