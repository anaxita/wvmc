package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/internal/service"
	"github.com/anaxita/wvmc/pkg/hasher"

	"github.com/dgrijalva/jwt-go"
)

// CtxString является ключем контекста для http запросов
type CtxString string

type customClaims struct {
	jwt.StandardClaims
	User entity.User
	Type string
}

type AuthHandler struct {
	userService *service.User
	authService *service.Auth
}

// NewAuthHandler возвращает новый AuthHandler
func NewAuthHandler(us *service.User, as *service.Auth) *AuthHandler {
	return &AuthHandler{userService: us, authService: as}
}

// SignIn выполняет аутентификацию пользователей и возвращает в ответе токен и роль пользователя
func (s *AuthHandler) SignIn() http.HandlerFunc {

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

		if user.Role == entity.UserRoleAdmin && user.Email != "admin" {
			addr := strings.Split(r.RemoteAddr, ":")
			ip := net.ParseIP(addr[0])
			if !ip.IsPrivate() {
				SendErr(
					w, http.StatusBadRequest,
					entity.ErrAccessDenied,
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

// RefreshToken выполняет переиздание токена
func (s *AuthHandler) RefreshToken() http.Handler {
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
