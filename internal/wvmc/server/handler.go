package server

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/anaxita/logit"
)

// respOK единая структура ответа без ошибки
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
		logit.Log("Ошибка отправки ответа в JSON", err)
		return
	}
	logit.Info("RESPONSE: ", code, fullResponse.Status)
}

// SendErr отправляет http ответ в формате JSON
func SendErr(w http.ResponseWriter, code int, meta error, err interface{}) {
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
	logit.Info("RESPONSE: ", code, fullResponse)
}

// Showlog показывает лог
func (s *Server) Showlog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filepath := os.Getenv("LOG")
		var bytesToRead int64 = 4096

		file, err := os.Open(filepath)
		if err != nil {
			SendErr(w, http.StatusInternalServerError, err, "Ошибка открытия файла лога")
		}
		defer file.Close()

		buf := make([]byte, bytesToRead)

		stat, err := os.Stat(filepath)
		if err != nil {
			logit.Log(err, buf)
		}

		start := stat.Size() - bytesToRead

		_, err = file.ReadAt(buf, start)
		if err != nil {
			logit.Log(err, buf)
		}

		w.Write(buf)
	}
}
