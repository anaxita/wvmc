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
	query := fmt.Sprintf("SELECT id, name, email, password, role FROM users WHERE %s = ?", key)

	if err := r.db.QueryRowContext(r.ctx, query, value).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.EncPassword,
		&u.Role,
	); err != nil {
		return u, err
	}

	logit.Log("УСПЕШНО нашли пользователя:", key, value)

	return u, nil
}

// Create создает пользователя и возвращает его ID, либо ошибку
func (r *UserRepository) Create(u model.User) (int, error) {
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
	logit.Log("Обновляем пользователю следующие поля:", u)

	query := "UPDATE users SET name = ?, email = ?, role = ? WHERE id = ? LIMIT 1"

	_, err := r.db.ExecContext(r.ctx, query, u.Name, u.Email, u.Role, u.ID)

	if err != nil {
		return err
	}

	logit.Log("УСПЕШНО обновили поля")

	return nil
}

// Delete удаляет пользователя, возвращает ошибку в случае неудачи
func (r *UserRepository) Delete(id string) error {
	logit.Log("Удаляем пользователя", id)

	query := "DELETE FROM users WHERE id = ? LIMIT 1"

	_, err := r.db.ExecContext(r.ctx, query, id)

	if err != nil {
		return err
	}

	logit.Log("УСПЕШНО удалили пользователя", id)

	return nil
}

// All возвращает массив из пользователей БД или ошибку
func (r *UserRepository) All() ([]model.User, error) {
	logit.Log("Получаем всех пользователей")

	rows, err := r.db.QueryContext(r.ctx, "SELECT id, name, email, role FROM users")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []model.User

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

	logit.Log("Успешно")
	return users, nil
}
