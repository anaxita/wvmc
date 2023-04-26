package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/internal/notice"
	"github.com/anaxita/wvmc/internal/service"
	"github.com/gorilla/mux"
)

type ServerHandler struct {
	serverService  *service.Server
	controlService *service.Control
	notifier       *notice.KMSBOT
}

func NewServerHandler(ss *service.Server, cs *service.Control, notifier *notice.KMSBOT) *ServerHandler {
	return &ServerHandler{serverService: ss, controlService: cs, notifier: notifier}
}

// GetServers возвращает список серверов
func (s *ServerHandler) GetServers() http.HandlerFunc {
	type response struct {
		Servers []entity.Server `json:"servers"`
	}

	var adminRole = 1
	var userRole = 0

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user := r.Context().Value(CtxString("user")).(entity.User)

		{
			ip4 := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
			if !ip4.IsPrivate() && !ip4.IsUnspecified() {
				defer s.notifier.AddIPToWL(user.Name, ip4.String(), "vmcontrol")
			}
		}

		if user.Role == adminRole {
			vms, err := s.controlService.GetServersDataForAdmins()
			if err != nil {
				SendErr(w, http.StatusOK, err, "Ошибка получения статусов")
				return
			}

			servers, err := s.serverService.Servers(ctx)
			if err != nil {
				SendErr(w, http.StatusOK, err, "Ошибка получения списка серверов")
				return
			}

			for k, v := range vms {
				for _, srv := range servers {
					if srv.VMID == v.VMID && srv.HV == v.HV {
						vms[k].ID = srv.ID
						vms[k].Company = srv.Company
						vms[k].Description = srv.Description
						vms[k].OutAddr = srv.OutAddr
						vms[k].IP = srv.IP

						break
					}
				}
			}
			SendOK(w, http.StatusOK, response{vms})
			return
		}

		if user.Role == userRole {
			servers, err := s.serverService.FindByUserID(ctx, user.ID)
			if err != nil {
				if err == sql.ErrNoRows {
					SendOK(w, http.StatusOK, response{make([]entity.Server, 0)})
					return
				}

				SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
				return
			}

			vms, err := s.controlService.GetServersDataForUsers(servers)
			if err != nil {
				SendErr(w, http.StatusInternalServerError, err, "Ошибка получения статусов")
				return
			}

			for k, v := range vms {
				for _, srv := range servers {
					if srv.VMID == v.VMID && srv.HV == v.HV {
						vms[k].ID = srv.ID
						vms[k].Company = srv.Company
						vms[k].Description = srv.Description
						vms[k].OutAddr = srv.OutAddr
						vms[k].IP = srv.IP

						break
					}
				}
			}

			SendOK(w, http.StatusOK, response{vms})
		}

	}
}

// GetServer получает информацию об 1 сервере по его хв и имени
func (s *ServerHandler) GetServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		hv := vars["hv"]
		name := vars["name"]

		server, err := s.serverService.FindByTitle(ctx, name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusOK, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		vmInfo, err := s.controlService.GetServerData(server, hv, name)
		if err != nil {
			SendErr(w, http.StatusOK, err, "can't to get vm info")
			return
		}

		SendOK(w, http.StatusOK, vmInfo)

	}
}

// ControlServer выполняет команды на сервере
func (s *ServerHandler) ControlServer() http.HandlerFunc {

	const notice = `
User: %s %s %s
Server: %s
HV: %s
Action: %s
`

	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(CtxString("user")).(entity.User)

		var err error

		server := r.Context().Value(CtxString("server")).(entity.Server)
		command := r.Context().Value(CtxString("command")).(string)

		switch command {
		case "start_power":
			_, err = s.controlService.StartServer(server)
		case "stop_power":
			_, err = s.controlService.StopServer(server)

		case "stop_power_force":
			_, err = s.controlService.StopServerForce(server)

		case "start_network":
			_, err = s.controlService.StartServerNetwork(server)

		case "stop_network":
			_, err = s.controlService.StopServerNetwork(server)
		default:
			SendErr(w, http.StatusBadRequest, errors.New("incorrect command"),
				"Неизвестная команда")
			return
		}

		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка выполнения команды")
			return
		}

		err = s.notifier.Notify(fmt.Sprintf(notice, user.Email, user.Name, user.Company, server.Name,
			server.HV, command))
		if err != nil {
			log.Println("Не удалось отправить уведомление", err)
		}

		SendOK(w, http.StatusOK, "Команда выполнена успешно")
	}
}

// UpdateAllServersInfo обновляет данные в БД по серверам
func (s *ServerHandler) UpdateAllServersInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := os.Getenv("SERVER_USER_NAME")
		password := os.Getenv("SERVER_USER_PASSWORD")

		servers, err := s.controlService.GetServersDataForAdmins()
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка powershell")
			return
		}

		for _, server := range servers {
			server.User = user
			server.Password = password
			_, err := s.serverService.Create(r.Context(), server)
			if err != nil {
				SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
				return
			}
		}
	}
}

// GetServerServices ...
func (s *ServerHandler) GetServerServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hv := vars["hv"]
		name := vars["name"]

		srv, err := s.serverService.FindByHvAndTitle(r.Context(), hv, name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		services, err := s.controlService.GetServerServices(srv.IP, srv.User, srv.Password)
		if err != nil {
			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
			return
		}

		SendOK(w, http.StatusOK, services)

	}
}

// GetServerManager ...
func (s *ServerHandler) GetServerManager() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		hv := vars["hv"]
		name := vars["name"]

		srv, err := s.serverService.FindByHvAndTitle(ctx, hv, name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusOK, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		processes, err := s.controlService.GetProcesses(srv.IP, srv.User, srv.Password)
		if err != nil {
			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
			return
		}

		SendOK(w, http.StatusOK, processes)

	}
}

// ControlServerManager control processes and user rdp sessions
func (s *ServerHandler) ControlServerManager() http.HandlerFunc {

	type req struct {
		ServerHV   string `json:"server_hv"`
		ServerName string `json:"server_name"`
		EntityID   int    `json:"entity_id"`
		Command    string `json:"command"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		var err error
		var task req

		err = json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			SendErr(w, http.StatusBadRequest, err, "невалидный json")
			return
		}

		task.ServerHV = vars["hv"]
		task.ServerName = vars["name"]

		server, err := s.serverService.FindByHvAndTitle(ctx, task.ServerHV, task.ServerName)
		if err != nil {
			SendErr(w, http.StatusNotFound, err, "Server is not found")
			return
		}

		switch task.Command {
		case "stop":
			_, err = s.controlService.StoptWinProcess(server.IP, server.User, server.Password,
				task.EntityID)
		case "disconnect":
			_, err = s.controlService.DisconnectRDPUser(server.IP, server.User, server.Password,
				task.EntityID)
		default:
			SendErr(w, http.StatusBadRequest, errors.New("undefind command"), "Неизвестная команда")
			return
		}

		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка выполнения команды")
			return
		}

		SendOK(w, http.StatusOK, "Команда выполнена успешно")
	}
}

// GetServerDisks return info about disks like letter, total and free size
func (s *ServerHandler) GetServerDisks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		hv := vars["hv"]
		name := vars["name"]

		srv, err := s.serverService.FindByHvAndTitle(ctx, hv, name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusOK, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}
		disksInfo, err := s.controlService.GetDiskFreeSpace(srv.IP, srv.User, srv.Password)
		if err != nil {
			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
			return
		}

		SendOK(w, http.StatusOK, disksInfo)
	}
}

// ControlServerServices управляет службами сервера
func (s *ServerHandler) ControlServerServices() http.HandlerFunc {

	type req struct {
		ServerHV    string `json:"server_hv"`
		ServerName  string `json:"server_name"`
		ServiceName string `json:"service_name"`
		Command     string `json:"command"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		var err error
		var task req

		err = json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			SendErr(w, http.StatusBadRequest, err, "невалидный json")
			return
		}

		task.ServerHV = vars["hv"]
		task.ServerName = vars["name"]

		server, err := s.serverService.FindByHvAndTitle(ctx, task.ServerHV, task.ServerName)
		if err != nil {
			SendErr(w, http.StatusNotFound, err, "Server is not found")
			return
		}

		switch task.Command {
		case "start":
			_, err = s.controlService.StartWinService(server.IP, server.User, server.Password,
				task.ServiceName)
		case "stop":
			_, err = s.controlService.StopWinService(server.IP, server.User, server.Password,
				task.ServiceName)
		case "restart":
			_, err = s.controlService.RestartWinService(server.IP, server.User, server.Password,
				task.ServiceName)
		default:
			SendErr(w, http.StatusBadRequest, errors.New("undefined command"), "Неизвестная команда")
			return
		}

		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибока выполнения команды")
			return
		}

		SendOK(w, http.StatusOK, "Команда выполнена успешно")
	}
}
