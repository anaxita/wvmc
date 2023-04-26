package api

import (
	"os"
	"time"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/dgrijalva/jwt-go"
)

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
