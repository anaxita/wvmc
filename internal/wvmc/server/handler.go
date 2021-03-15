package server

import (
	"encoding/json"
	"net/http"

	"github.com/anaxita/logit"
)

// respOK единая структура ответа без ошибки
type respOK struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
}

// respErr единая структура для описания ошибки на техническом и обычном языке
type respErr struct {
	Error string `json:"err"`
	Meta  string `json:"meta"`
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
		logit.Log("Ошибка отправки ответа в JSON", err)
		return
	}
	logit.Info(code, "Response: ", fullResponse)
}

// SendErr отправляет http ответ в формате JSON
func SendErr(w http.ResponseWriter, code int, meta error, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	fullResponse := respOK{
		Status: "err",
		Message: respErr{
			Error: err,
			Meta:  meta.Error(),
		},
	}

	if err := json.NewEncoder(w).Encode(fullResponse); err != nil {
		logit.Log("Ошибка отправки ответа в JSON", err)
		return
	}
	logit.Info(code, "Response: ", fullResponse)
}
