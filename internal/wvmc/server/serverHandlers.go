package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/anaxita/wvmc/internal/wvmc/model"
)

// GetServers возвращает список серверов
func (s *Server) GetServers() http.HandlerFunc {
	type response struct {
		Servers []model.Server `json:"servers"`
	}

	userRole := 0
	adminRole := 1
	return func(w http.ResponseWriter, r *http.Request) {
		var servers []model.Server
		var err error

		store := s.store.Server(r.Context())

		user := r.Context().Value(CtxString("user")).(model.User)
		if user.Role == adminRole {
			servers, err = store.All()
		}

		if user.Role == userRole {
			servers, err = store.FindByUser(user.ID)
		}

		if err != nil {
			if err == sql.ErrNoRows {
				SendOK(w, http.StatusOK, response{make([]model.Server, 0)})
				return
			}
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
		}

		SendOK(w, http.StatusOK, response{servers})
	}
}

// CreateServer создает сервер
func (s *Server) CreateServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := model.Server{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
		}

		store := s.store.Server(r.Context())

		_, err := store.Find("id", req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err := store.Create(req)
				if err != nil {
					SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
					return
				}

				SendOK(w, http.StatusCreated, "Created")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendErr(w, http.StatusBadRequest, errors.New("User is exists"), "Сервер уже существует")
	}
}

// EditServer обновляет данные сервера
func (s *Server) EditServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := model.Server{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
		}

		store := s.store.Server(r.Context())

		_, err := store.Find("id", req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "Сервер не найден")
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

// DeleteServer удаляет пользователя
func (s *Server) DeleteServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := model.Server{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
		}

		store := s.store.Server(r.Context())

		_, err := store.Find("id", req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "Сервер не найден")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
		}

		err = store.Delete(req.ID)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
		}

		SendOK(w, http.StatusOK, "Deleted")
	}
}
