package server

import (
	"database/sql"
	"net/http"

	// "sync"

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
