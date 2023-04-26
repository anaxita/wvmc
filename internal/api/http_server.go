package api

import (
	"net/http"
	"time"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/gorilla/mux"
)

// NewServer - создает новый сервер
func NewServer(
	port string,
	user *UserHandler,
	auth *AuthHandler,
	server *ServerHandler,
	mw *Middleware,
) *http.Server {
	r := mux.NewRouter()
	r.Use(mw.Cors)

	setAuthRoutes(r, auth)
	setUserRoutes(r, user, mw)
	setServerRoutes(r, server, mw)

	return &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    time.Second * 5,
		WriteTimeout:   time.Second * 5,
		IdleTimeout:    time.Second * 5,
		MaxHeaderBytes: 1 << 20,
	}
}

func setAuthRoutes(r *mux.Router, authHandler *AuthHandler) {
	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/signin", authHandler.SignIn).Methods(http.MethodPost, http.MethodOptions)
	auth.HandleFunc("/refresh", authHandler.RefreshToken).Methods(http.MethodPost, http.MethodOptions)
}

func setUserRoutes(r *mux.Router, userHandler *UserHandler, mw *Middleware) {
	users := r.PathPrefix("/users").Subrouter()
	users.Use(mw.Auth, mw.RoleMiddleware(entity.UserRoleAdmin))
	users.Handle("", userHandler.GetUsers()).Methods(http.MethodGet, http.MethodOptions)
	users.Handle("", userHandler.CreateUser()).Methods(http.MethodPost, http.MethodOptions)
	users.Handle("", userHandler.EditUser()).Methods(http.MethodPatch, http.MethodOptions)
	users.Handle("", userHandler.DeleteUser()).Methods(http.MethodDelete, http.MethodOptions)
	users.Handle("/servers", userHandler.AddServersToUser()).Methods(http.MethodPost, http.MethodOptions)
	users.Handle("/{user_id}/servers", userHandler.GetUserServers()).Methods(http.MethodGet, http.MethodOptions)
}

func setServerRoutes(r *mux.Router, serverHandler *ServerHandler, mw *Middleware) {
	servers := r.PathPrefix("/servers").Subrouter()
	servers.Use(mw.Auth)
	servers.Handle("", serverHandler.GetServers()).Methods(http.MethodGet, http.MethodOptions)
	servers.Handle("/{hv}/{name}", serverHandler.GetServer()).Methods(http.MethodGet, http.MethodOptions)
	servers.Handle("/{hv}/{name}/disks", serverHandler.GetServerDisks()).Methods(http.MethodGet, http.MethodOptions)
	servers.Handle("/{hv}/{name}/services", serverHandler.GetServerServices()).Methods(http.MethodGet, http.MethodOptions)
	servers.Handle("/{hv}/{name}/services", serverHandler.ControlServerServices()).Methods(http.MethodPost, http.MethodOptions)
	servers.Handle("/{hv}/{name}/manager", serverHandler.GetServerManager()).Methods(http.MethodGet, http.MethodOptions)
	servers.Handle("/{hv}/{name}/manager", serverHandler.ControlServerManager()).Methods(http.MethodPost, http.MethodOptions)
	servers.Handle("/update", serverHandler.UpdateAllServersInfo()).Methods(http.MethodPost, http.MethodOptions)

	serversControl := servers.PathPrefix("/control").Subrouter()
	serversControl.Use(mw.CheckControlPermissions)
	serversControl.Handle("", serverHandler.ControlServer()).Methods(http.MethodPost, http.MethodOptions)
}
