package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/dgrijalva/jwt-go"
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

// createToken создает новый access токен и записывает в него модель пользователя
func createToken(t string, user entity.User) string {

	// Создаем данные токена с временем жизни 15 минут и моделью пользователя
	var claims customClaims

	if t == "access" {
		claims = customClaims{
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 3600).Unix(),
			},
			user,
			t,
		}
	}
	if t == "refresh" {
		claims = customClaims{
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * (24 * 30)).Unix(),
			},
			user,
			t,
		}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, _ := token.SignedString([]byte(os.Getenv("TOKEN")))
	return tokenString
}
