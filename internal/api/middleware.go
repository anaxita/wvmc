package api

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Middleware struct {
	*helperHandler
	userService   *service.User
	serverService *service.Server
	authService   *service.Auth
}

func NewMiddleware(l *zap.SugaredLogger, us *service.User, ss *service.Server) *Middleware {
	return &Middleware{
		helperHandler: newHelperHandler(l),
		userService:   us,
		serverService: ss,
	}
}

// Auth выполняет проверку токена
func (s *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		tokenString := strings.Split(authHeader, " ")

		if len(tokenString) != 2 || tokenString[0] != "Bearer" {
			err := fmt.Errorf("%w: no bearer token in header", entity.ErrUnauthorized)
			s.sendError(w, err)
			return
		}

		user, err := s.authService.Auth(ctx, tokenString[1])
		if err != nil {
			s.sendError(w, err)
			return
		}

		ctx = context.WithValue(ctx, entity.CtxUserKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Cors устанавливает cors заголовки
func (s *Middleware) Cors(next http.Handler) http.Handler {

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

// RoleMiddleware проверяет, является ли пользователь админом
func (s *Middleware) RoleMiddleware(roles ...int) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := entity.CtxUser(r.Context())
			if err != nil {
				s.sendError(w, err)
				return
			}

			for _, v := range roles {
				if user.Role == v {
					next.ServeHTTP(w, r)
					return
				}
			}

			s.sendError(w, fmt.Errorf("%w: to do that, you must have one of these roles: %v", entity.ErrForbidden, roles))
		})
	}
}

// Recover middleware recovers from panics and logs the error.
func (s *Middleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.l.Errorw("panic", "error", err, "error_stack", string(debug.Stack()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
