package server

import (
	"net/http"

	"github.com/anaxita/wvmc/internal/wvmc/store"
	"github.com/gorilla/mux"
)

// Server - структура http сервера
type Server struct {
	store  store.Storager
	router *mux.Router
}

// New - создает новый сервер
func New(store store.Storager) *Server {
	return &Server{store: store}
}

// Start - запускает сервер
func (s *Server) Start() error {
	s.configureRouter()
	return http.ListenAndServe(":8080", s.router)
}
