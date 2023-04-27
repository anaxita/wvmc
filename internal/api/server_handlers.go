package api

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	notifier       *notice.Service
}

func NewServerHandler(l *zap.SugaredLogger, ss *service.Server, cs *service.Control, notifier *notice.Service) *ServerHandler {
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
