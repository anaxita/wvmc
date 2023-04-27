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
	r.Use(mw.Recover, mw.Cors)

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

func setUserRoutes(r *mux.Router, user *UserHandler, mw *Middleware) {
	users := r.PathPrefix("/users").Subrouter()
	users.Use(mw.Auth, mw.RoleMiddleware(entity.RoleAdmin))
	users.HandleFunc("", user.GetUsers).Methods(http.MethodGet, http.MethodOptions)
	users.HandleFunc("", user.CreateUser).Methods(http.MethodPost, http.MethodOptions)
	users.HandleFunc("", user.EditUser).Methods(http.MethodPatch, http.MethodOptions)
	users.HandleFunc("", user.DeleteUser).Methods(http.MethodDelete, http.MethodOptions)
	users.HandleFunc("/{id}/servers", user.AddServers).Methods(http.MethodPost, http.MethodOptions)
	users.HandleFunc("/{id}/servers", user.GetUserServers).Methods(http.MethodGet, http.MethodOptions)
}

func setServerRoutes(r *mux.Router, serverHandler *ServerHandler, mw *Middleware) {
	servers := r.PathPrefix("/servers").Subrouter()
	servers.Use(mw.Auth)
	servers.HandleFunc("", serverHandler.GetServers).Methods(http.MethodGet, http.MethodOptions)
	// servers.Handle("/{hv}/{name}", serverHandler.GetServer()).Methods(http.MethodGet, http.MethodOptions)
	// servers.Handle("/{hv}/{name}/disks", serverHandler.GetServerDisks()).Methods(http.MethodGet, http.MethodOptions)
	// servers.Handle("/{hv}/{name}/services", serverHandler.getServerServices()).Methods(http.MethodGet, http.MethodOptions)
	// servers.Handle("/{hv}/{name}/services", serverHandler.ControlServerServices()).Methods(http.MethodPost, http.MethodOptions)
	// servers.Handle("/{hv}/{name}/manager", serverHandler.GetServerManager()).Methods(http.MethodGet, http.MethodOptions)
	// servers.Handle("/{hv}/{name}/manager", serverHandler.ControlServerManager()).Methods(http.MethodPost, http.MethodOptions)
	servers.HandleFunc("/update", serverHandler.UpdateAllServersInfo).Methods(http.MethodPost, http.MethodOptions)
	servers.HandleFunc("/control", serverHandler.ControlServer).Methods(http.MethodPost, http.MethodOptions)
}
