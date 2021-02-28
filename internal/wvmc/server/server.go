package server

import (
	"net/http"

	"github.com/anaxita/wvmc/internal/wvmc/store"
	"github.com/gorilla/mux"
)

// Server - структура http сервера
type Server struct {
	store  *store.Store
	router *mux.Router
}

// New - создает новый сервер
func New(storage *store.Store) *Server {
	return &Server{store: storage}
}

// Start - запускает сервер
func (s *Server) Start() error {
	s.configureRouter()
	return http.ListenAndServe(":8080", s.router)
}
