package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	entity2 "github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/pkg/hasher"

	"github.com/dgrijalva/jwt-go"
)

// CtxString является ключем контекста для http запросов
type CtxString string

type customClaims struct {
	jwt.StandardClaims
	User entity2.User
	Type string
}

// createToken создает новый access токен и записывает в него модель пользователя
func createToken(t string, user entity2.User) string {

	// Создаем данные токена с временем жизни 15 минут и моделью пользователя
	var claims customClaims

	if t == "access" {
		claims = customClaims{
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 3600).Unix(),
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
		AccessToken  string `json:"access_token"`
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

		if req.Email == "" || req.Password == "" {
			SendErr(w, http.StatusBadRequest, errors.New("email or password cannot be empty"),
				"Поля email или password не могут быть пустыми")
			return
		}

		user, err := s.userService.FindByEmail(r.Context(), req.Email)
		if err != nil {
			SendErr(w, http.StatusOK, err, "Неверный логин или пароль")
			return
		}

		err = hasher.Compare(user.Password, req.Password)
		if err != nil {
			SendErr(w, http.StatusOK, err, "Неверный логин или пароль")
			return
		}

		if user.Role == entity2.UserRoleAdmin && user.Email != "admin" {
			addr := strings.Split(r.RemoteAddr, ":")
			ip := net.ParseIP(addr[0])
			if !ip.IsPrivate() {
				SendErr(
					w, http.StatusBadRequest,
					entity2.ErrAccessDenied,
					fmt.Sprintf("Доступ разрешен только с локального IP, ваш айпи %v",
						r.RemoteAddr),
				)

				return
			}
		}

		accessToken := createToken("access", user)
		refreshToken := createToken("refresh", user)

		err = s.authService.CreateRefreshToken(r.Context(), user.ID, refreshToken)
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
