package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/control"
	"github.com/anaxita/wvmc/internal/wvmc/model"
)

// GetServers возвращает список серверов
func (s *Server) GetServers() http.HandlerFunc {
	type response struct {
		Servers []control.VM `json:"servers"`
	}

	var userRole = 0
	var adminRole = 1

	return func(w http.ResponseWriter, r *http.Request) {

		user := r.Context().Value(CtxString("user")).(model.User)

		if user.Role == adminRole {
			hvList := strings.Split(os.Getenv("HV_LIST"), ",")
			var vms []control.VM
			var wg sync.WaitGroup
			var mu sync.Mutex

			wg.Add(len(hvList))
			for _, hv := range hvList {
				go func(hv string, vms *[]control.VM, wg *sync.WaitGroup, mu *sync.Mutex) {
					defer wg.Done()

					servers, err := s.serverService.GetServerDataForAdmins(hv)
					if err != nil {
						// SendErr(w, http.StatusInternalServerError, err, "Ошибка получения статусов")
						// return
						logit.Log("PS", err)
					}

					mu.Lock()
					defer mu.Unlock()
					*vms = append(*vms, servers...)

				}(hv, &vms, &wg, &mu)
			}

			wg.Wait()
			// vms, err := s.serverService.GetServersDataForAdmins()
			// if err != nil {
			// 	SendErr(w, http.StatusInternalServerError, err, "Ошибка получения статусов")
			// 	logit.Log("PS", err)
			// 	return
			// }
			// SendOK(w, http.StatusOK, response{vms})

		}

		if user.Role == userRole {
			servers, err := s.store.Server(r.Context()).FindByUser(user.ID)
			if err != nil {
				if err == sql.ErrNoRows {
					SendOK(w, http.StatusOK, response{make([]control.VM, 0)})
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
			SendOK(w, http.StatusOK, response{vms})
		}

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
	type controlRequest struct {
		ServerID string `json:"server_id"`
		Command  string `json:"command"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := controlRequest{}
		var err error
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendErr(w, http.StatusBadRequest, errors.New("server_id and command fields cannot be empty"), "server_id и command не могут быть пустыми")
			return
		}

		if req.ServerID == "" || req.Command == "" {
			SendErr(w, http.StatusOK, err, "Ошибка выполнения команды")
			return
		}
		s := control.NewServerService(&control.Command{})

		switch req.Command {
		case "start-power":
			_, err = s.StartServer(req.ServerID)
		case "stop-power":
			_, err = s.StopServer(req.ServerID)

		case "start-power-force":
			_, err = s.StopServerForce(req.ServerID)

		case "start-network":
			_, err = s.StartServerNetwork(req.ServerID)

		case "stop-network":
			_, err = s.StopServerNetwork(req.ServerID)
		default:
			SendErr(w, http.StatusBadRequest, errors.New("undefiend command"), "Неизвестная команда")
			return
		}

		if err != nil {
			SendErr(w, http.StatusOK, err, "Ошибка выполнения команды")
			return
		}

		SendOK(w, http.StatusOK, "Команда выполнена успешно")
	}
}

// UpdateAllServersInfo обновляет данные в БД по серверам
func (s *Server) UpdateAllServersInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var servers []model.Server

		out, err := control.NewServerService(&control.Command{}).UpdateAllServersInfo()
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка powershell")
			return
		}

		err = json.Unmarshal(out, &servers)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка декодирования")
			return
		}

		duplicates := make(map[string]int)

		for _, server := range servers {
			if duplicates[server.ID] > 0 {
				logit.Log("ДУБЛЬ", server.Name, server.ID)
			}
			duplicates[server.ID] += 1

			_, _ = s.store.Server(r.Context()).Create(server)
			if err != nil {
				// SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
				logit.Log("Невозможно добавить сервер", server.Name, err)
				// return
			}
		}
		logit.Info("Дубликаты:", duplicates)

		SendOK(w, http.StatusOK, "Updated")
	}
}
