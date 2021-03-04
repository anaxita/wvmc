package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/model"
)

// UserRepository - содержит методы работы с пользовательскими моделями
type UserRepository struct {
	db  *sql.DB
	ctx context.Context
}

//Find ищет первое совпадение пользователя с заданным ключом и значением, возвращает модель либо ошибку
func (r *UserRepository) Find(key, value string) (model.User, error) {
	logit.Log("Ищем пользователя:", key, value)

	u := model.User{}

	query := fmt.Sprintf("SELECT id, * FROM users WHERE %s = ?", key)
	if err := r.db.QueryRowContext(r.ctx, query, value).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.EncPassword,
		&u.Role,
	); err != nil {
		return u, err
	}

	logit.Log("Нашли пользователя:", key, value)
	return u, nil
}

// Create создает пользователя и возвращает его ID, либо ошибку
func (r *UserRepository) Create(u model.User) (int, error) {
	logit.Log("Создааем пользователя:", u.Name)
	query := "INSERT INTO users (name, email, password, role) VALUES(?, ?, ?, ?)"
	result, err := r.db.ExecContext(r.ctx, query, u.Name, u.Email, u.EncPassword, u.Role)
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Edit обновляет данные пользователя, возвращает ошибку в случае неудачи
func (r *UserRepository) Edit(u model.User) error {
	logit.Log("Обновляем поля пользователю:", u.Name)

	query := "UPDATE users SET name = ?, email = ?, role = ? WHERE id = ? LIMIT 1"

	_, err := r.db.ExecContext(r.ctx, query, u.Name, u.Email, u.Role, u.ID)
	if err != nil {
		return err
	}

	logit.Log("Успешно обновили поля")
	return nil
}

// Delete удаляет пользователя, возвращает ошибку в случае неудачи
func (r *UserRepository) Delete(id string) error {
	logit.Log("Удаляем пользователя", id)
	_, err := r.db.ExecContext(r.ctx, "DELETE FROM users WHERE id = ? LIMIT 1", id)
	if err != nil {
		return err
	}
	logit.Log("Успешно удалили пользователя", id)
	return nil
}

// All возвращает массив из пользователей БД или ошибку
func (r *UserRepository) All() ([]model.User, error) {
	logit.Log("Получаем всех пользователей")

	var users []model.User

	rows, err := r.db.QueryContext(r.ctx, "SELECT id, name, email, role FROM users")
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	logit.Log("Успешно получили всех пользователей")
	return users, nil
}

// CreateRefreshToken добавляет  запись о токене или обновляет , если запись уже есть
func (r *UserRepository) CreateRefreshToken(userID, refreshToken string) error {
	logit.Log("Записываем в БД рефреш токен пользователя ", userID)

	query := "INSERT INTO refresh_tokens (user_id, token) VALUES(?, ?) ON DUPLICATE KEY UPDATE user_id = VALUES(user_id), token = VALUES(token) "
	_, err := r.db.ExecContext(r.ctx, query, userID, refreshToken)
	if err != nil {
		return err
	}

	logit.Log("Успешно записали рефреш токен")
	return nil
}

// GetRefreshToken проверяет есть ли в базе токен
func (r *UserRepository) GetRefreshToken(token string) error {
	logit.Log("Ищем в БД рефреш токен")
	var t string

	query := "SELECT user_id FROM refresh_tokens WHERE token = ?"
	err := r.db.QueryRowContext(r.ctx, query, token).Scan(&t)
	if err != nil {
		return err
	}

	logit.Log("Рефреш токен найден")
	return nil
}
