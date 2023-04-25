package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

// respOK единая структура ответа
type respOK struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
}

// respErr единая структура для описания ошибки на техническом и обычном языке
type respErr struct {
	Error interface{} `json:"err"`
	Meta  string      `json:"meta"`
}

// SendOK отправляет http ответ в формате JSON
func SendOK(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	fullResponse := respOK{
		Status:  "ok",
		Message: data,
	}

	if err := json.NewEncoder(w).Encode(fullResponse); err != nil {
		return
	}
}

// SendErr отправляет http ответ в формате JSON
func SendErr(w http.ResponseWriter, code int, meta error, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if meta == nil {
		meta = errors.New("undefined error")
	}

	fullResponse := respOK{
		Status: "err",
		Message: respErr{
			Error: err,
			Meta:  meta.Error(),
		},
	}

	if err := json.NewEncoder(w).Encode(fullResponse); err != nil {
		return
	}
}
