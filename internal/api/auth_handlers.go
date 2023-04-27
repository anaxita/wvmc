package api

import (
	"encoding/json"
	"net/http"

	"github.com/anaxita/wvmc/internal/api/requests"
	"github.com/anaxita/wvmc/internal/api/responses"
	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/internal/service"
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
	var resp responses.PostSignIn

	err := func() error {
		var req requests.PostSignIn
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return err
		}

		data, err := h.authService.SignIn(r.Context(), req.Email, req.Password)
		if err != nil {
			return err
		}

		// TODO !!! IN MIDDLEWARE !!!
		// TODO add check on local ip (admins cannot login from not local ip)
		// TODO !!! IN MIDDLEWARE !!!

		resp = responses.NewPostSignIn(data)

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
	var resp responses.PostRefresh

	err := func() error {
		var req requests.PostRefresh
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return err
		}

		data, err := h.authService.RefreshTokens(r.Context(), req.RefreshToken)
		if err != nil {
			return err
		}

		resp = responses.NewPostRefresh(data)

		return nil
	}()

	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJson(w, resp)
}
