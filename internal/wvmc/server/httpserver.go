package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/control"
	"github.com/anaxita/wvmc/internal/wvmc/store"
	"github.com/gorilla/mux"
)

// Server - структура http сервера
type Server struct {
	store         *store.Store
	router        *mux.Router
	serverService *control.ServerService
}

// New - создает новый сервер
func New(storage *store.Store) *Server {
	return &Server{
		store:         storage,
		router:        mux.NewRouter(),
		serverService: control.NewServerService(&control.Command{}),
	}
}

func (s *Server) configureRouter() {
	r := s.router
	r.Use(s.Cors)
	r.Handle("/refresh", s.RefreshToken()).Methods("POST", "OPTIONS")
	r.Handle("/signin", s.SignIn()).Methods("POST", "OPTIONS")
	r.Handle("/update", s.UpdateAllServersInfo()).Methods("GET", "OPTIONS") // TODO: удалить когда уйдет в продакшен (аналог /servers/update)
	r.Handle("/log", s.Showlog()).Methods("GET", "OPTIONS")

	users := r.NewRoute().Subrouter()
	users.Use(s.Auth, s.CheckIsAdmin)
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
	servers.Use(s.Auth, s.CheckIsAdmin)
	// servers.Handle("/servers", s.CreateServer()).Methods("POST", "OPTIONS") // disabled because working auto
	servers.Handle("/servers", s.EditServer()).Methods("OPTIONS", "PATCH")
	servers.Handle("/servers/{hv}/{name}", s.GetServer()).Methods("OPTIONS", "GET")
	servers.Handle("/servers/{hv}/{name}/services", s.GetServerServices()).Methods("OPTIONS", "GET")
	servers.Handle("/servers/update", s.UpdateAllServersInfo()).Methods("POST", "OPTIONS")
}

// Start - запускает сервер
func (s *Server) Start() error {
	s.configureRouter()

	cer, err := ioutil.ReadFile("C:\\Apache24\\conf\\ssl\\kmsys.ru.cer")

	if err != nil {
		logit.Fatal("Ошибка открытия kmsys.ru.cer:", err)
	}

	ca, err := ioutil.ReadFile("C:\\Apache24\\conf\\ssl\\ca.cer")
	if err != nil {
		logit.Fatal("Ошибка открытия ca.cer :", err)
	}

	c := fmt.Sprintf("%v \n %v", string(cer), string(ca))

	goCer, err := os.Create("C:\\Apache24\\conf\\ssl\\anaxita.cer")

	if err != nil {
		logit.Fatal("Ошибка создания anaxita.cer :", err)
	}

	goCer.WriteString(c)
	defer goCer.Close()

	logit.Info("Сервер запущен на : ", os.Getenv("PORT_HTTPS"))

	go http.ListenAndServe(os.Getenv("PORT_HTTP"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
	}))

	return http.ListenAndServeTLS(os.Getenv("PORT_HTTPS"), goCer.Name(), "C:\\Apache24\\conf\\ssl\\kmsys.ru.key", s.router)

	// return http.ListenAndServe(os.Getenv("PORT"), s.router)
}
