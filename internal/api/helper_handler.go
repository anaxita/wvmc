package api

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type helperHandler struct {
	l *zap.SugaredLogger
}

func newHelperHandler(l *zap.SugaredLogger) *helperHandler {
	return &helperHandler{l: l}
}

func (h *helperHandler) sendJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
	h.l.Info("response ok")
}

func (h *helperHandler) sendEmpty(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
	h.l.Info("response ok")
}

func (h *helperHandler) sendError(w http.ResponseWriter, err error) {
	if err == nil {
		h.sendEmpty(w)
		h.l.Warn("sendError called with nil error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})

	h.l.Errorw("response", zap.Error(err))
}
