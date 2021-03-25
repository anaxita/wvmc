package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/anaxita/wvmc/internal/wvmc/hasher"
	"github.com/anaxita/wvmc/internal/wvmc/model"
	"github.com/gorilla/mux"
)

// GetUsers возвращает список всех пользователей
func (s *Server) GetUsers() http.HandlerFunc {
	type response struct {
		User []model.User `json:"users"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.store.User(r.Context()).All()
		if err != nil {
			if err == sql.ErrNoRows {
				SendOK(w, http.StatusOK, response{make([]model.User, 0)})
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
		UserID int `json:"id,string"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := model.User{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		store := s.store.User(r.Context())

		_, err := store.Find("email", req.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				encPassword, err := hasher.Hash(req.Password)
				if err != nil {
					SendErr(w, http.StatusInternalServerError, err, "Невозможно создать хеш")
					return
				}

				req.EncPassword = string(encPassword)

				createdID, err := store.Create(req)
				if err != nil {
					SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
					return
				}

				SendOK(w, http.StatusCreated, response{createdID})
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendErr(w, http.StatusBadRequest, errors.New("user is exists"), "Пользователь уже существует")
	}
}

// EditUser обновляет данные пользователя
func (s *Server) EditUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := model.User{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		store := s.store.User(r.Context())

		_, err := store.Find("id", req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "Пользователь не найден")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		err = store.Edit(req)
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
		req := model.User{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		store := s.store.User(r.Context())

		_, err := store.Find("id", req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "Пользователь не найден")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		err = store.Delete(req.ID)
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
		UserID  string         `json:"user_id"`
		Servers []model.Server `json:"servers"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		err := s.store.User(r.Context()).AddServer(req.UserID, req.Servers)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendOK(w, http.StatusOK, "Added")
	}
}

// GetUserServers получает сервера пользователя
func (s *Server) GetUserServers() http.HandlerFunc {
	type response struct {
		Servers []model.Server `json:"servers"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		userID, ok := vars["user_id"]
		if !ok {
			SendErr(w, http.StatusBadRequest, errors.New("user id is undefiend"), "Неверный данные в запросе")
			return
		}

		servers, err := s.store.Server(r.Context()).FindByUser(userID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendOK(w, http.StatusOK, response{make([]model.Server, 0)})
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendOK(w, http.StatusOK, response{servers})
	}
}
