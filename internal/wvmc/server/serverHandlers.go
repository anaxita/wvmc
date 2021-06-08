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

			vms, err := s.serverService.GetServersDataForAdmins()
			if err != nil {
				SendErr(w, http.StatusInternalServerError, err, "Ошибка получения статусов")
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

			vms, err := s.serverService.GetServersDataForUsers(servers)
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

		_, err := store.Find("title", name)
		if err != nil {
			if err == sql.ErrNoRows {
				SendErr(w, http.StatusNotFound, err, "server is not found")
				return
			}

			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
			return
		}

		vmInfo, err := control.NewServerService(&control.Command{}).GetServerData(hv, name)
		if err != nil {
			SendErr(w, http.StatusNotFound, err, "can't to get vm info")
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

		SendErr(w, http.StatusBadRequest, errors.New("server is already exists"), "Сервер уже существует")
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
				SendErr(w, http.StatusNotFound, err, "Сервер не найден")
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
			_, err = s.serverService.StartServer(server)
		case "stop_power":
			_, err = s.serverService.StopServer(server)

		case "stop_power_force":
			_, err = s.serverService.StopServerForce(server)

		case "start_network":
			_, err = s.serverService.StartServerNetwork(server)

		case "stop_network":
			_, err = s.serverService.StopServerNetwork(server)
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
		servers, err := s.serverService.UpdateAllServersInfo()
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

func (s *Server) GetServerServices() http.HandlerFunc {

	type response struct {
		Services []control.WinServices `json:"services"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// vars := mux.Vars(r)
		// hv := vars["hv"]
		// name := vars["name"]

		// store := s.store.Server(r.Context())

		// s, err := store.Find("title", name)
		// if err != nil {
		// 	if err == sql.ErrNoRows {
		// 		SendErr(w, http.StatusNotFound, err, "server is not found")
		// 		return
		// 	}

		// 	SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
		// 	return
		// }

		vmInfo, err := control.NewServerService(&control.Command{}).GetServerServices("", "", "")
		if err != nil {
			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
			return
		}

		SendOK(w, http.StatusOK, response{vmInfo})

	}
}
