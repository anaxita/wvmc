package api

import (
	"net/http"
	"os"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/internal/notice"
	"github.com/anaxita/wvmc/internal/service"
	"github.com/gorilla/mux"
)

// Server - структура http сервера
type Server struct {
	router         *mux.Router
	controlService *service.ControlService
	notify         *notice.KMSBOT
	userService    *service.User
	serverService  *service.Server
	authService    *service.Auth
}

// New - создает новый сервер
func New(
	controlService *service.ControlService,
	notify *notice.KMSBOT,
	userService *service.User,
	serverService *service.Server,
	authService *service.Auth,
) *Server {
	return &Server{
		router:         mux.NewRouter(),
		controlService: controlService,
		notify:         notify,
		userService:    userService,
		serverService:  serverService,
		authService:    authService,
	}
}

func (s *Server) configureRouter() {
	r := s.router
	r.Use(s.Cors)
	r.Handle("/refresh", s.RefreshToken()).Methods("POST", "OPTIONS")
	r.Handle("/signin", s.SignIn()).Methods("POST", "OPTIONS")

	users := r.NewRoute().Subrouter()
	users.Use(s.Auth, s.RoleMiddleware(entity.UserRoleAdmin))

	users.Handle("/users", s.GetUsers()).Methods("OPTIONS", "GET")
	users.Handle("/users", s.CreateUser()).Methods("OPTIONS", "POST")
	users.Handle("/users", s.EditUser()).Methods("OPTIONS", "PATCH")
	users.Handle("/users", s.DeleteUser()).Methods("OPTIONS", "DELETE")
	users.Handle("/users/servers", s.AddServersToUser()).Methods("OPTIONS", "POST")
	users.Handle("/users/{user_id}/servers", s.GetUserServers()).Methods("OPTIONS", "GET")

	serversShow := r.NewRoute().Subrouter()
	serversShow.Use(s.Auth)
	serversShow.Handle("/servers", s.GetServers()).Methods("OPTIONS", "GET")

	serversControl := r.NewRoute().Subrouter()
	serversControl.Use(s.Auth, s.CheckControlPermissions)
	serversControl.Handle("/servers/control", s.ControlServer()).Methods("POST", "OPTIONS")

	servers := r.NewRoute().Subrouter()
	servers.Use(s.Auth, s.RoleMiddleware(entity.UserRoleAdmin))

	servers.Handle("/servers/{hv}/{name}", s.GetServer()).Methods("OPTIONS", "GET")
	servers.Handle("/servers/{hv}/{name}/disks", s.GetServerDisks()).Methods("OPTIONS", "GET")
	servers.Handle("/servers/{hv}/{name}/services", s.GetServerServices()).Methods("OPTIONS", "GET")
	servers.Handle("/servers/{hv}/{name}/services", s.ControlServerServices()).Methods("OPTIONS",
		"POST")
	servers.Handle("/servers/{hv}/{name}/manager", s.GetServerManager()).Methods("OPTIONS", "GET")
	servers.Handle("/servers/{hv}/{name}/manager", s.ControlServerManager()).Methods("OPTIONS",
		"POST")
	servers.Handle("/servers/update", s.UpdateAllServersInfo()).Methods("POST", "OPTIONS")
}

// Start - запускает сервер
func (s *Server) Start() error {
	s.configureRouter()

	return http.ListenAndServe(os.Getenv("HTTP_PORT"), s.router)
}
