package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anaxita/wvmc/internal/api/requests"
	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type UserHandler struct {
	*helperHandler
	userService   *service.User
	serverService *service.Server
}

func NewUserHandler(l *zap.SugaredLogger, us *service.User, ss *service.Server) *UserHandler {
	return &UserHandler{
		helperHandler: newHelperHandler(l),
		userService:   us,
		serverService: ss,
	}
}

// GetUsers возвращает список всех пользователей
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []entity.User

	err := func() (err error) {
		users, err = h.userService.Users(r.Context())
		return err
	}()
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJson(w, users)
}

// CreateUser создает пользователя
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req entity.UserCreate

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, fmt.Errorf("%w: %s", entity.ErrValidate, err))
		return
	}

	user, err := h.userService.Create(r.Context(), req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJson(w, user)
}

// EditUser обновляет данные пользователя
func (h *UserHandler) EditUser(w http.ResponseWriter, r *http.Request) {
	var req entity.UserEdit

	err := func() (err error) {
		userID := mux.Vars(r)["id"]
		id, err := uuid.Parse(userID)
		if err != nil {
			return fmt.Errorf("%w: %s", entity.ErrValidate, err)
		}

		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			return fmt.Errorf("%w: %s", entity.ErrValidate, err)
		}

		err = h.userService.Edit(r.Context(), id, req)
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendEmpty(w)
}

// DeleteUser удаляет пользователя
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	err := func() (err error) {
		userID := mux.Vars(r)["id"]
		id, err := uuid.Parse(userID)
		if err != nil {
			return fmt.Errorf("%w: %s", entity.ErrValidate, err)
		}

		err = h.userService.Delete(r.Context(), id)
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendEmpty(w)
}

// AddServers добавляет пользователю сервер
func (h *UserHandler) AddServers(w http.ResponseWriter, r *http.Request) {
	err := func() (err error) {
		userID := mux.Vars(r)["id"]
		id, err := uuid.Parse(userID)
		if err != nil {
			return fmt.Errorf("%w: %s", entity.ErrValidate, err)
		}

		var req requests.AddServers
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return fmt.Errorf("%w: %s", entity.ErrValidate, err)
		}

		err = h.serverService.SetUserServers(r.Context(), id, req.ServerIDs)
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendEmpty(w)
}

// GetUserServers возвращат список серверов где доступные пользователю помечены полем added = true
func (h *UserHandler) GetUserServers(w http.ResponseWriter, r *http.Request) {
	var servers []entity.Server

	err := func() (err error) {
		userID := mux.Vars(r)["id"]
		id, err := uuid.Parse(userID)
		if err != nil {
			return fmt.Errorf("%w: %s", entity.ErrValidate, err)
		}

		servers, err = h.serverService.UserServers(r.Context(), id)
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJson(w, servers)
}
