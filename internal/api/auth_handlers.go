package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/anaxita/wvmc/internal/api/requests"
	"github.com/anaxita/wvmc/internal/api/responses"
	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/internal/service"
	"github.com/anaxita/wvmc/pkg/hasher"
	"go.uber.org/zap"

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
	*helperHandler
	userService *service.User
	authService *service.Auth
}

// NewAuthHandler возвращает новый AuthHandler
func NewAuthHandler(l *zap.SugaredLogger, us *service.User, as *service.Auth) *AuthHandler {
	return &AuthHandler{
		helperHandler: newHelperHandler(l),
		userService:   us,
		authService:   as,
	}
}

// SignIn выполняет аутентификацию пользователей и возвращает в ответе токен и роль пользователя
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var resp responses.SignIn

	err := func() error {
		var req requests.SignIn
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return err
		}

		// TODO add validateion

		user, err := h.userService.FindByEmail(r.Context(), req.Email)
		if err != nil {
			return err
		}

		err = hasher.Compare(user.Password, req.Password)
		if err != nil {
			return err
		}

		if user.Role == entity.UserRoleAdmin && user.Email != "admin" {
			addr := strings.Split(r.RemoteAddr, ":")
			ip := net.ParseIP(addr[0])
			if !ip.IsPrivate() {
				return fmt.Errorf("%w: Доступ разрешен только с локального IP, ваш айпи %v", entity.ErrForbidden, ip)
			}
		}

		accessToken := createToken("access", user)
		refreshToken := createToken("refresh", user)

		err = h.authService.CreateRefreshToken(r.Context(), user.ID, refreshToken)
		if err != nil {
			return err
		}

		resp = responses.NewSignIn(accessToken, refreshToken)

		return nil
	}()
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJson(w, resp)
}

// RefreshToken выполняет переиздание токена
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var resp responses.Refresh

	err := func() error {
		var req requests.Refresh
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return err
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
			return fmt.Errorf("%w: %v", entity.ErrUnauthorized, err)
		}

		// Проверяем корректность Claims (Данных внутри токена) и проверяем Валидность (не совсем понимаю что это значит) ключа подписи
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			tokenType, ok := claims["Type"]
			if ok && tokenType == "refresh" {

				u := entity.User{}
				userjson, _ := json.Marshal(claims["User"])
				json.Unmarshal(userjson, &u)

				err = h.authService.RefreshToken(r.Context(), req.RefreshToken)

				if err != nil {
					return fmt.Errorf("%w: %v", entity.ErrUnauthorized, err)
				}

				refreshToken := createToken("refresh", u)

				err = h.authService.CreateRefreshToken(r.Context(), u.ID, refreshToken)
				if err != nil {
					return err
				}

				resp = responses.NewRefresh(createToken("access", u), refreshToken)

				return nil
			}

			return fmt.Errorf("%w: token is not refresh", entity.ErrUnauthorized)
		}

		return fmt.Errorf("%w: invalid token", entity.ErrUnauthorized)
	}()

	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJson(w, resp)
}
