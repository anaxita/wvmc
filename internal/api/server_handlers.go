package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/anaxita/wvmc/internal/api/requests"
	"github.com/anaxita/wvmc/internal/entity"
	"github.com/anaxita/wvmc/internal/notice"
	"github.com/anaxita/wvmc/internal/service"
	"go.uber.org/zap"
)

type ServerHandler struct {
	*helperHandler
	serverService  *service.Server
	controlService *service.Control
	notifier       *notice.KMSBOT
}

func NewServerHandler(l *zap.SugaredLogger, ss *service.Server, cs *service.Control, notifier *notice.KMSBOT) *ServerHandler {
	return &ServerHandler{
		helperHandler:  newHelperHandler(l),
		serverService:  ss,
		controlService: cs,
		notifier:       notifier,
	}
}

// GetServers возвращает список серверов
func (h *ServerHandler) GetServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.serverService.Servers(r.Context())
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJson(w, servers)
}

// GetServer получает информацию об 1 сервере по его хв и имени
// func (h *ServerHandler) GetServer(w http.ResponseWriter, r *http.Request) {
// 	var server entity.Server
//
// 	err := func() error {
// 		serverID := mux.Vars(r)["server_id"]
// 		id, err := strconv.ParseInt(serverID, 10, 64)
// 		if err != nil {
// 			return fmt.Errorf("%w: %s", entity.ErrValidate, err)
// 		}
//
// 		server, err = h.serverService.FindByID(r.Context(), id)
// 		if err != nil {
// 			return err
// 		}
//
// 		return nil
// 	}()
// 	if err != nil {
// 		h.sendError(w, err)
// 		return
// 	}
//
// 	h.sendJson(w, server)
// }

// ControlServer выполняет команды на сервере
func (h *ServerHandler) ControlServer(w http.ResponseWriter, r *http.Request) {
	var req requests.ControlServer
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.sendError(w, fmt.Errorf("%w: %s", entity.ErrValidate, err))
		return
	}

	err = h.serverService.Control(r.Context(), req.ServerID, req.Command)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendEmpty(w)
}

// UpdateAllServersInfo обновляет данные в БД по серверам
func (h *ServerHandler) UpdateAllServersInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := os.Getenv("SERVER_USER_NAME")
		password := os.Getenv("SERVER_USER_PASSWORD")

		servers, err := h.controlService.GetServersDataForAdmins()
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка powershell")
			return
		}

		for _, server := range servers {
			server.User = user
			server.Password = password
			_, err := h.serverService.Create(r.Context(), server)
			if err != nil {
				SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
				return
			}
		}
	}
}

// // getServerServices ...
// func (h *ServerHandler) getServerServices() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		hv := vars["hv"]
// 		name := vars["name"]
//
// 		srv, err := h.serverService.FindByHvAndTitle(r.Context(), hv, name)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				SendErr(w, http.StatusNotFound, err, "server is not found")
// 				return
// 			}
//
// 			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
// 			return
// 		}
//
// 		services, err := h.controlService.getServerServices(srv.IP, srv.User, srv.Password)
// 		if err != nil {
// 			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
// 			return
// 		}
//
// 		SendOK(w, http.StatusOK, services)
//
// 	}
// }
//
// // GetServerManager ...
// func (h *ServerHandler) GetServerManager() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		vars := mux.Vars(r)
// 		hv := vars["hv"]
// 		name := vars["name"]
//
// 		srv, err := h.serverService.FindByHvAndTitle(ctx, hv, name)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				SendErr(w, http.StatusOK, err, "server is not found")
// 				return
// 			}
//
// 			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
// 			return
// 		}
//
// 		processes, err := h.controlService.getProcesses(srv.IP, srv.User, srv.Password)
// 		if err != nil {
// 			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
// 			return
// 		}
//
// 		SendOK(w, http.StatusOK, processes)
//
// 	}
// }
//
// // ControlServerManager control processes and user rdp sessions
// func (h *ServerHandler) ControlServerManager() http.HandlerFunc {
//
// 	type req struct {
// 		ServerHV   string `json:"server_hv"`
// 		ServerName string `json:"server_name"`
// 		EntityID   int    `json:"entity_id"`
// 		Command    string `json:"command"`
// 	}
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		vars := mux.Vars(r)
// 		var err error
// 		var task req
//
// 		err = json.NewDecoder(r.Body).Decode(&task)
// 		if err != nil {
// 			SendErr(w, http.StatusBadRequest, err, "невалидный json")
// 			return
// 		}
//
// 		task.ServerHV = vars["hv"]
// 		task.ServerName = vars["name"]
//
// 		server, err := h.serverService.FindByHvAndTitle(ctx, task.ServerHV, task.ServerName)
// 		if err != nil {
// 			SendErr(w, http.StatusNotFound, err, "Server is not found")
// 			return
// 		}
//
// 		switch task.Command {
// 		case "stop":
// 			_, err = h.controlService.stoptWinProcess(server.IP, server.User, server.Password,
// 				task.EntityID)
// 		case "disconnect":
// 			_, err = h.controlService.disconnectRDPUser(server.IP, server.User, server.Password,
// 				task.EntityID)
// 		default:
// 			SendErr(w, http.StatusBadRequest, errors.New("undefind command"), "Неизвестная команда")
// 			return
// 		}
//
// 		if err != nil {
// 			SendErr(w, http.StatusInternalServerError, err, "Ошибка выполнения команды")
// 			return
// 		}
//
// 		SendOK(w, http.StatusOK, "Команда выполнена успешно")
// 	}
// }
//
// // GetServerDisks return info about disks like letter, total and free size
// func (h *ServerHandler) GetServerDisks() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		vars := mux.Vars(r)
// 		hv := vars["hv"]
// 		name := vars["name"]
//
// 		srv, err := h.serverService.FindByHvAndTitle(ctx, hv, name)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				SendErr(w, http.StatusOK, err, "server is not found")
// 				return
// 			}
//
// 			SendErr(w, http.StatusInternalServerError, err, "Ошибка БД")
// 			return
// 		}
// 		disksInfo, err := h.controlService.getDiskFreeSpace(srv.IP, srv.User, srv.Password)
// 		if err != nil {
// 			SendErr(w, http.StatusOK, err, "Ошибка подключения к серверу")
// 			return
// 		}
//
// 		SendOK(w, http.StatusOK, disksInfo)
// 	}
// }
//
// // ControlServerServices управляет службами сервера
// func (h *ServerHandler) ControlServerServices() http.HandlerFunc {
//
// 	type req struct {
// 		ServerHV    string `json:"server_hv"`
// 		ServerName  string `json:"server_name"`
// 		ServiceName string `json:"service_name"`
// 		Command     string `json:"command"`
// 	}
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		vars := mux.Vars(r)
// 		var err error
// 		var task req
//
// 		err = json.NewDecoder(r.Body).Decode(&task)
// 		if err != nil {
// 			SendErr(w, http.StatusBadRequest, err, "невалидный json")
// 			return
// 		}
//
// 		task.ServerHV = vars["hv"]
// 		task.ServerName = vars["name"]
//
// 		server, err := h.serverService.FindByHvAndTitle(ctx, task.ServerHV, task.ServerName)
// 		if err != nil {
// 			SendErr(w, http.StatusNotFound, err, "Server is not found")
// 			return
// 		}
//
// 		switch task.Command {
// 		case "start":
// 			_, err = h.controlService.startWinService(server.IP, server.User, server.Password,
// 				task.ServiceName)
// 		case "stop":
// 			_, err = h.controlService.stopWinService(server.IP, server.User, server.Password,
// 				task.ServiceName)
// 		case "restart":
// 			_, err = h.controlService.restartWinService(server.IP, server.User, server.Password,
// 				task.ServiceName)
// 		default:
// 			SendErr(w, http.StatusBadRequest, errors.New("undefined command"), "Неизвестная команда")
// 			return
// 		}
//
// 		if err != nil {
// 			SendErr(w, http.StatusInternalServerError, err, "Ошибока выполнения команды")
// 			return
// 		}
//
// 		SendOK(w, http.StatusOK, "Команда выполнена успешно")
// 	}
// }
