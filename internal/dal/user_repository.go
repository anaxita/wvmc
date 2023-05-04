package dal

import (
	"context"
	"database/sql"
	"errors"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/google/uuid"
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

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (u entity.User, err error) {
	q := "SELECT id, name, email, company, password, role FROM users WHERE id = ? LIMIT 1"

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
func (r *UserRepository) Create(ctx context.Context, u entity.UserCreate) error {
	query := "INSERT INTO users (id, name, email, company, password, role) VALUES (?, ?, ?, ?, ?, ?)"

	_, err := r.db.ExecContext(ctx, query, u.ID, u.Name, u.Email, u.Company, u.Password, u.Role)
	if err != nil {
		return err
	}

	return nil
}

// Update обновляет данные пользователя u с паролем или без withPass, возвращает ошибку в случае неудачи
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, u entity.UserEdit) error {
	q := "UPDATE users SET name = ?, company = ?, role = ? WHERE id = ? "

	result, err := r.db.ExecContext(ctx, q, u.Name, u.Company, u.Role, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

// Delete удаляет пользователя, возвращает ошибку в случае неудачи
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ? ", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

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
func (r *UserRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) error {
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
