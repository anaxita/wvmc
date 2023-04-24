package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/anaxita/wvmc/internal/wvmc/model"
	"github.com/jmoiron/sqlx"
)

// UserRepository - содержит методы работы с пользовательскими моделями
type UserRepository struct {
	db  *sqlx.DB
	ctx context.Context
}

// Find ищет первое совпадение пользователя с заданным ключом и значением, возвращает модель либо ошибку
func (r *UserRepository) Find(key, value string) (model.User, error) {
	u := model.User{}

	query := fmt.Sprintf("SELECT id, name, email, password, company, role FROM users WHERE %s = ?",
		key)
	if err := r.db.QueryRowContext(r.ctx, query, value).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.EncPassword,
		&u.Company,
		&u.Role,
	); err != nil {
		return u, err
	}

	return u, nil
}

// Create создает пользователя и возвращает его ID, либо ошибку
func (r *UserRepository) Create(u model.User) (int, error) {
	query := "INSERT INTO users (name, email, company, password, role) VALUES (?, ?, ?, ?, ?)"

	result, err := r.db.ExecContext(r.ctx, query, u.Name, u.Email, u.Company, u.EncPassword, u.Role)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Edit обновляет данные пользователя u с паролем или без withPass, возвращает ошибку в случае неудачи
func (r *UserRepository) Edit(u model.User, withPass bool) error {
	var query string
	var err error

	if withPass {
		query = "UPDATE users SET name = ?, company = ?, role = ?, password = ? WHERE id = ? "
		_, err = r.db.ExecContext(r.ctx, query, u.Name, u.Company, u.Role, u.EncPassword, u.ID)
	} else {
		query = "UPDATE users SET name = ?, company = ?, role = ? WHERE id = ? "
		_, err = r.db.ExecContext(r.ctx, query, u.Name, u.Company, u.Role, u.ID)
	}
	if err != nil {
		return err
	}

	return nil
}

// Delete удаляет пользователя, возвращает ошибку в случае неудачи
func (r *UserRepository) Delete(id string) error {
	if id == "129" {
		return errors.New("нельзя удалить главного админа")
	}
	_, err := r.db.ExecContext(r.ctx, "DELETE FROM users WHERE id = ? ", id)
	if err != nil {
		return err
	}

	return nil
}

// All возвращает массив из пользователей БД или ошибку
func (r *UserRepository) All() ([]model.User, error) {
	var users []model.User

	rows, err := r.db.QueryContext(r.ctx, "SELECT id, name, email, company, role FROM users")
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Company, &user.Role)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return users, nil
}

// CreateRefreshToken добавляет  запись о токене или обновляет , если запись уже есть
func (r *UserRepository) CreateRefreshToken(userID, refreshToken string) error {
	query := "INSERT INTO refresh_tokens (user_id, token) VALUES(?, ?) ON CONFLICT(user_id) DO UPDATE SET user_id = user_id, token = ? "
	_, err := r.db.ExecContext(r.ctx, query, userID, refreshToken, refreshToken)
	if err != nil {
		return err
	}

	return nil
}

// GetRefreshToken проверяет есть ли в базе токен
func (r *UserRepository) GetRefreshToken(token string) error {
	var t string

	query := "SELECT user_id FROM refresh_tokens WHERE token = ?"
	err := r.db.QueryRowContext(r.ctx, query, token).Scan(&t)
	if err != nil {
		return err
	}

	return nil
}

// AddServer добавляет сервера пользователю по его айди
func (r *UserRepository) AddServer(userID string, servers []model.Server) error {
	query := "INSERT INTO users_servers (user_id, server_id) VALUES(?, ?)"

	stmt, err := r.db.PrepareContext(r.ctx, query)
	if err != nil {
		return err
	}

	for _, v := range servers {
		_, err := stmt.ExecContext(r.ctx, userID, v.ID)
		if err != nil {
			return err
		}
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return nil
}
