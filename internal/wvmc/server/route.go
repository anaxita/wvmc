package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) configureRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/", hello())

	s.router = r
}

func hello() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	}
}
