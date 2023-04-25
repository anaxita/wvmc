package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	entity2 "github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/pkg/hasher"
	"github.com/gorilla/mux"
)

// GetUsers возвращает список всех пользователей
func (s *Server) GetUsers() http.HandlerFunc {
	type response struct {
		User []entity2.User `json:"users"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.userService.Users(r.Context())
		if err != nil {
			if err == sql.ErrNoRows {
				SendOK(w, http.StatusOK, response{make([]entity2.User, 0)})
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendOK(w, http.StatusOK, response{users})
	}
}

// CreateUser создает пользователя
func (s *Server) CreateUser() http.HandlerFunc {
	type response struct {
		UserID int64 `json:"id,string"` // TODO why string?
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req entity2.User

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		req.Password = strings.TrimSpace(req.Password)
		req.Name = strings.TrimSpace(req.Name)
		req.Company = strings.TrimSpace(req.Company)

		if req.Email == "" || req.Password == "" || req.Name == "" {
			SendErr(w, http.StatusBadRequest, errors.New("email password or name cannot be empty"),
				"Поля email, password или name не могут быть пустыми")
			return
		}

		match, err := regexp.Match(`^([a-zA-Z0-9_]){3,15}$`, []byte(req.Email))
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "incorrect email")
			return
		}

		if !match {
			SendErr(w, http.StatusBadRequest,
				errors.New("email должен быть 3-15 символов, начинаться с буквы и содержать только английские буквы, цифры и знак подчеркивания"),
				"incorrect regexp")
			return
		}

		_, err = s.userService.FindByEmail(r.Context(), req.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				encPassword, err := hasher.Hash(req.Password)
				if err != nil {
					SendErr(w, http.StatusInternalServerError, err, "Невозможно создать хеш")
					return
				}

				req.Password = string(encPassword)

				user, err := s.userService.Create(ctx, req)
				if err != nil {
					SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
					return
				}

				SendOK(w, http.StatusCreated, response{user.ID})
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendErr(w, http.StatusBadRequest, errors.New("user is exists"),
			"Пользователь уже существует")
	}
}

// EditUser обновляет данные пользователя
func (s *Server) EditUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := entity2.User{}
		var err error

		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		_, err = s.userService.FindByID(r.Context(), req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "Пользователь не найден")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		// edit user data with/without password
		if req.Password != "" {
			encPassword, _err := hasher.Hash(req.Password)
			if err != nil {
				SendErr(w, http.StatusInternalServerError, _err, "Невозможно создать хеш пароля")
				return
			}

			req.Password = string(encPassword)

			err = s.userService.Edit(r.Context(), req, true)

		} else {
			err = s.userService.Edit(r.Context(), req, false)
		}

		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendOK(w, http.StatusOK, "Updated")
	}
}

// DeleteUser удаляет пользователя
func (s *Server) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := entity2.User{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		_, err := s.userService.FindByID(r.Context(), req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "Пользователь не найден")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		err = s.userService.Delete(r.Context(), req.ID)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendOK(w, http.StatusOK, "Deleted")
	}
}

// AddServersToUser добавляет пользователю сервер
func (s *Server) AddServersToUser() http.HandlerFunc {
	type request struct {
		UserID  int64            `json:"user_id"`
		Servers []entity2.Server `json:"servers"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		_, err := s.userService.FindByID(r.Context(), req.UserID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "User not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		allServers, err := s.serverService.Servers(r.Context())
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "User not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		serversToAdd := make([]entity2.Server, 0)
	loop:
		for _, server := range allServers {
			for _, reqServer := range req.Servers {
				if server.ID == reqServer.ID {
					serversToAdd = append(serversToAdd, reqServer)
					continue loop
				}
			}
		}

		err = s.serverService.AddServersToUser(r.Context(), req.UserID, serversToAdd)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendOK(w, http.StatusOK, "Added")
	}
}

// GetUserServers возвращат список серверов где доступные пользователю помечены полем added = true
func (s *Server) GetUserServers() http.HandlerFunc {
	type addedServers struct {
		ID      int64  `json:"id"`
		VMID    string `json:"vmid"`
		Name    string `json:"name"`
		HV      string `json:"hv"`
		IP      string `json:"ip"`
		Company string `json:"company"`
		Added   bool   `json:"is_added"`
	}

	type response struct {
		Servers []addedServers `json:"servers"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		vars := mux.Vars(r)
		userID, ok := vars["user_id"]
		if !ok {
			SendErr(w, http.StatusBadRequest, errors.New("user id is undefined"),
				"Неверный данные в запросе")
			return
		}

		uID, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			SendErr(w, http.StatusBadRequest, err, "user id is not number")
			return
		}

		allServers, err := s.serverService.Servers(ctx)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		userServers, err := s.serverService.FindByUserID(ctx, uID)
		if err != nil {
			if err == sql.ErrNoRows {
				userServers = make([]entity2.Server, 0)
			} else {
				SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
				return
			}
		}

		res := make([]addedServers, 0, len(allServers))

		for _, srv := range allServers {
			res = append(res,
				addedServers{
					ID:      srv.ID,
					VMID:    srv.VMID,
					Name:    srv.Name,
					HV:      srv.HV,
					Company: srv.Company,
				})
		}

	loop:
		for k, addedSrv := range res {
			for _, us := range userServers {
				if addedSrv.ID == us.ID {
					res[k].Added = true
					continue loop
				}
			}
		}

		SendOK(w, http.StatusOK, response{res})
	}
}
