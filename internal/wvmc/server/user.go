package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/anaxita/wvmc/internal/wvmc/model"
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
		}
		SendOK(w, http.StatusOK, response{users})
	}
}

// CreateUser создает пользователя
func (s *Server) CreateUser() http.HandlerFunc {
	type response struct {
		UserID int `json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := model.User{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
		}

		createdID, err := s.store.User(r.Context()).Create(req)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
		}
		SendOK(w, http.StatusCreated, response{createdID})
	}
}
