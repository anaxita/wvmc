package hasher

import (
	"golang.org/x/crypto/bcrypt"
)

// Hash возвращает хешированную строку
func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
}

// Compare сравнивает хеш со строкой
func Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
