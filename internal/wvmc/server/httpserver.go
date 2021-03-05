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

func (s *Server) configureRouter() {
	r := mux.NewRouter()
	r.Use(s.Cors)

	r.Handle("/refresh", s.RefreshToken()).Methods("GET", "OPTIONS")

	signin := r.NewRoute().Subrouter()
	signin.Use(s.Cors)
	signin.HandleFunc("/signing", s.SignIn()).Methods("OPTIONS, POST")

	users := r.NewRoute().Subrouter()
	users.Use(s.Cors, s.Auth)
	users.HandleFunc("/users", s.GetUsers()).Methods("OPTIONS, GET")
	users.HandleFunc("/users", s.CreateUser()).Methods("OPTIONS, POST")
	users.HandleFunc("/users", s.EditUser()).Methods("OPTIONS, PATCH")
	users.HandleFunc("/users", s.DeleteUser()).Methods("OPTIONS, DELETE")
	s.router = r
}

// Start - запускает сервер
func (s *Server) Start() error {
	s.configureRouter()
	return http.ListenAndServe(":8080", s.router)
}
