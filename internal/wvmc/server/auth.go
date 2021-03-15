package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anaxita/wvmc/internal/wvmc/hasher"
	"github.com/anaxita/wvmc/internal/wvmc/model"
	"github.com/dgrijalva/jwt-go"
)

// CtxString является ключем контекста для http запросов
type CtxString string

type customClaims struct {
	jwt.StandardClaims
	User model.User
	Type string
}

// createToken создает новый access токен и записывает в него модель пользователя
func createToken(t string, user model.User) string {

	// Создаем данные токена с временем жизни 15 минут и моделью пользователя
	var claims customClaims

	if t == "access" {
		claims = customClaims{
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			},
			user,
			t,
		}
	}
	if t == "refresh" {
		claims = customClaims{
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * (24 * 30)).Unix(),
			},
			user,
			t,
		}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, _ := token.SignedString([]byte(os.Getenv("TOKEN")))
	return tokenString
}

// SignIn выполняет аутентификацию пользователей и возвращает в ответе токен и роль пользователя
func (s *Server) SignIn() http.HandlerFunc {

	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		AccessToken  string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный формат запроса")
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		req.Password = strings.TrimSpace(req.Password)

		user, err := s.store.User(r.Context()).Find("email", req.Email)
		if err != nil {
			SendErr(w, http.StatusUnprocessableEntity, err, "Неверный логин")
			return
		}

		err = hasher.Compare(user.EncPassword, req.Password)
		if err != nil {
			SendErr(w, http.StatusUnprocessableEntity, err, "Неверный пароль")
			return
		}

		accessToken := createToken("access", user)
		refreshToken := createToken("refresh", user)

		err = s.store.User(r.Context()).CreateRefreshToken(user.ID, refreshToken)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		resp := response{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		SendOK(w, http.StatusOK, resp)
	}
}
