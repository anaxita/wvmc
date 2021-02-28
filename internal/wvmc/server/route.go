package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) configureRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/users", showUsers("hi")).Methods("OPTIONS, GET")

	s.router = r
}

func showUsers(word string) http.HandlerFunc {
	type response struct {
		Message string `json:"message"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if word == "error" {
			json.NewEncoder(w).Encode(response{"САМ ТЫ ОШИБКА"})
			return
		}
		json.NewEncoder(w).Encode(response{word})

	}
}
