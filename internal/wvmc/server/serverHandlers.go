package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/control"
	"github.com/anaxita/wvmc/internal/wvmc/model"
	"github.com/gorilla/mux"
)

// GetServers возвращает список серверов
func (s *Server) GetServers() http.HandlerFunc {
	type response struct {
		Servers []model.Server `json:"servers"`
	}

	var adminRole = 1
	var userRole = 0

	return func(w http.ResponseWriter, r *http.Request) {

		user := r.Context().Value(CtxString("user")).(model.User)

		if user.Role == adminRole {

			vms, err := s.controlService.GetServersDataForAdmins()
			if err != nil {
				SendErr(w, http.StatusOK, err, "Ошибка получения статусов")
				logit.Log("PS", err)
				return
			}
			SendOK(w, http.StatusOK, response{vms})
			return
		}

		if user.Role == userRole {
			servers, err := s.store.Server(r.Context()).FindByUser(user.ID)
			if err != nil {
				if err == sql.ErrNoRows {
					SendOK(w, http.StatusOK, response{make([]model.Server, 0)})
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

		loop:
			for k, v := range vms {
				for _, srv := range servers {
					if srv.ID == v.ID {
						vms[k].Company = srv.Company
						vms[k].Description = srv.Description
						vms[k].OutAddr = srv.OutAddr
						vms[k].IP = srv.IP

						continue loop
					}
				}
			}

			SendOK(w, http.StatusOK, response{vms})
		}

	}
}

// GetServer получает информацию об 1 сервере по его хв и имени
func (s *Server) GetServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hv := vars["hv"]
		name := vars["name"]

		store := s.store.Server(r.Context())

		server, err := store.Find("title", name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusOK, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		vmInfo, err := control.NewServerService(&control.Command{}).GetServerData(server, hv, name)
		if err != nil {
			SendErr(w, http.StatusOK, err, "can't to get vm info")
			return
		}

		SendOK(w, http.StatusOK, vmInfo)

	}
}

// CreateServer создает сервер
func (s *Server) CreateServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := model.Server{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		store := s.store.Server(r.Context())

		_, err := store.Find("id", req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err := store.Create(req)
				if err != nil {
					SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
					return
				}

				SendOK(w, http.StatusOK, "added")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendErr(w, http.StatusOK, errors.New("server is already exists"), "Сервер уже существует")
	}
}

// EditServer обновляет данные сервера
func (s *Server) EditServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := model.Server{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		store := s.store.Server(r.Context())

		_, err := store.Find("id", req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusOK, err, "Сервер не найден")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		err = store.Edit(req)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendOK(w, http.StatusOK, "Updated")
	}
}

// DeleteServer удаляет пользователя
func (s *Server) DeleteServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := model.Server{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, err, "Неверный данные в запросе")
			return
		}

		store := s.store.Server(r.Context())

		_, err := store.Find("id", req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "Сервер не найден")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		err = store.Delete(req.ID)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		SendOK(w, http.StatusOK, "Deleted")
	}
}

// ControlServer выполняет команды на сервере
func (s *Server) ControlServer() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		server := r.Context().Value(CtxString("server")).(model.Server)
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
			SendErr(w, http.StatusBadRequest, errors.New("undefiend command"), "Неизвестная команда")
			return
		}

		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка выполнения команды")
			return
		}

		SendOK(w, http.StatusOK, "Команда выполнена успешно")
	}
}

// UpdateAllServersInfo обновляет данные в БД по серверам
func (s *Server) UpdateAllServersInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := s.controlService.UpdateAllServersInfo()
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка powershell")
			return
		}
		duplicates := make(map[string]int)
		duplicatesServers := make([]model.Server, 0)
		for _, server := range servers {
			if duplicates[server.ID] > 0 {
				logit.Log("ДУБЛЬ", server.Name, server.ID)
				duplicatesServers = append(duplicatesServers, server)
			}
			duplicates[server.ID] += 1

			_, err := s.store.Server(r.Context()).Create(server)
			if err != nil {
				logit.Log("Невозможно добавить сервер", server.Name, err)
				SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
				return
			}
		}

		logit.Log("ДУБЛИ: ", duplicatesServers)

		SendOK(w, http.StatusOK, "Updated")
	}
}

// Get services
func (s *Server) GetServerServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hv := vars["hv"]
		name := vars["name"]

		srv, err := s.store.Server(r.Context()).FindByHVandName(hv, name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusOK, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		services, err := control.NewServerService(&control.Command{}).GetServerServices(srv.IP, srv.User, srv.Password)
		if err != nil {
			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
			return
		}

		SendOK(w, http.StatusOK, services)

	}
}

// Get processes
func (s *Server) GetServerManager() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hv := vars["hv"]
		name := vars["name"]

		srv, err := s.store.Server(r.Context()).FindByHVandName(hv, name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusOK, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		processes, err := control.NewServerService(&control.Command{}).GetProcesses(srv.IP, srv.User, srv.Password)
		if err != nil {
			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
			return
		}

		SendOK(w, http.StatusOK, processes)

	}
}

// ControlServerManager control processes and user rdp sessions
func (s *Server) ControlServerManager() http.HandlerFunc {

	type req struct {
		ServerHV   string `json:"server_hv"`
		ServerName string `json:"server_name"`
		EnityID    int    `json:"enity_id"`
		Command    string `json:"command"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var task req

		err = json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			SendErr(w, http.StatusBadRequest, err, "невалидный json")
			return
		}

		server, err := s.store.Server(r.Context()).FindByHVandName(task.ServerHV, task.ServerName)
		if err != nil {
			SendErr(w, http.StatusNotFound, err, "Server is not found")
			return
		}

		switch task.Command {
		case "disconnect":
			_, err = s.controlService.StoptWinProcess(server.IP, server.User, server.Password, task.EnityID)
		case "stop":
			_, err = s.controlService.DisconnectRDPUser(server.IP, server.User, server.Password, task.EnityID)
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
func (s *Server) GetServerDisks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hv := vars["hv"]
		name := vars["name"]

		srv, err := s.store.Server(r.Context()).FindByHVandName(hv, name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusOK, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}
		disksInfo, err := control.NewServerService(&control.Command{}).GetDiskFreeSpace(srv.IP, srv.User, srv.Password)
		if err != nil {
			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
			return
		}

		SendOK(w, http.StatusOK, disksInfo)
	}
}

// ControlServerServices управляет службами сервера
func (s *Server) ControlServerServices() http.HandlerFunc {

	type req struct {
		ServerHV    string `json:"server_hv"`
		ServerName  string `json:"server_name"`
		ServiceName string `json:"service_name"`
		Command     string `json:"command"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
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

		server, err := s.store.Server(r.Context()).FindByHVandName(task.ServerHV, task.ServerName)
		if err != nil {
			SendErr(w, http.StatusNotFound, err, "Server is not found")
			return
		}

		switch task.Command {
		case "start":
			_, err = s.controlService.StartWinService(server.IP, server.User, server.Password, task.ServiceName)
		case "stop":
			_, err = s.controlService.StopWinService(server.IP, server.User, server.Password, task.ServiceName)
		case "restart":
			_, err = s.controlService.RestartWinService(server.IP, server.User, server.Password, task.ServiceName)
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
