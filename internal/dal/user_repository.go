package dal

import (
	"context"
	"database/sql"
	"errors"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/jmoiron/sqlx"
)

// UserRepository - содержит методы работы с пользовательскими моделями
type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (u entity.User, err error) {
	q := "SELECT id, name, email, password, company, role FROM users WHERE email = ? LIMIT 1"

	err = r.db.GetContext(ctx, &u, q, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, entity.ErrNotFound
		}

		return u, err
	}

	return u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (u entity.User, err error) {
	q := "SELECT id, name, email, company, role FROM users WHERE id = ? LIMIT 1"

	err = r.db.GetContext(ctx, &u, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, entity.ErrNotFound
		}

		return u, err
	}

	return u, nil
}

// Create создает пользователя и возвращает его ID, либо ошибку
func (r *UserRepository) Create(ctx context.Context, user entity.User) (int64, error) {
	query := "INSERT INTO users (name, email, company, password, role) VALUES (?, ?, ?, ?, ?)"

	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.Company, user.Password, user.Role)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Edit обновляет данные пользователя u с паролем или без withPass, возвращает ошибку в случае неудачи
func (r *UserRepository) Edit(ctx context.Context, u entity.User, withPass bool) error {
	var query string
	var err error

	if withPass {
		query = "UPDATE users SET name = ?, company = ?, role = ?, password = ? WHERE id = ? "
		_, err = r.db.ExecContext(ctx, query, u.Name, u.Company, u.Role, u.Password, u.ID)
	} else {
		query = "UPDATE users SET name = ?, company = ?, role = ? WHERE id = ? "
		_, err = r.db.ExecContext(ctx, query, u.Name, u.Company, u.Role, u.ID)
	}
	if err != nil {
		return err
	}

	return nil
}

// Delete удаляет пользователя, возвращает ошибку в случае неудачи
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ? ", id)
	return err
}

// Users возвращает массив из пользователей БД или ошибку
func (r *UserRepository) Users(ctx context.Context) (users []entity.User, err error) {
	q := "SELECT id, name, email, company, role FROM users"

	err = r.db.SelectContext(ctx, &users, q)
	if err != nil {
		return users, err
	}

	return users, nil
}

// CreateRefreshToken добавляет запись о токене или обновляет, если запись уже есть
func (r *UserRepository) CreateRefreshToken(ctx context.Context, userID int64, refreshToken string) error {
	query := "INSERT INTO refresh_tokens (user_id, token) VALUES(?, ?) ON CONFLICT(user_id) DO UPDATE SET user_id = user_id, token = ? "

	_, err := r.db.ExecContext(ctx, query, userID, refreshToken, refreshToken)
	if err != nil {
		return err
	}

	return nil
}

// RefreshToken проверяет есть ли в базе токен
func (r *UserRepository) RefreshToken(ctx context.Context, token string) error {
	var t string

	query := "SELECT user_id FROM refresh_tokens WHERE token = ?"
	err := r.db.QueryRowContext(ctx, query, token).Scan(&t)
	if err != nil {
		return err
	}

	return nil
}
