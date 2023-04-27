package dal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/anaxita/wvmc/internal/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ServerRepository - содержит методы работы с пользовательскими моделями.
type ServerRepository struct {
	db *sqlx.DB
}

func NewServerRepository(db *sqlx.DB) *ServerRepository {
	return &ServerRepository{
		db: db,
	}
}

func (r *ServerRepository) FindByID(ctx context.Context, id int64) (s entity.Server, err error) {
	q := "SELECT id, vmid, title, ip4, hv, company, out_addr, description, user_name, user_password FROM servers WHERE id = ?"

	err = r.db.GetContext(ctx, &s, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s, entity.ErrNotFound
		}

		return s, err
	}

	return s, nil
}

// Upsert создает сервер и возвращает его ID, либо ошибку.
func (r *ServerRepository) Upsert(ctx context.Context, s ...entity.Server) error {
	query := `INSERT INTO servers (id, title, hv, state, status) 
    VALUES (?, ?, ?, ?, ?, ?, ?) 
    ON CONFLICT (id) 
        DO UPDATE SET title = ?, ip4 = ?, hv = ?, state = ?, status = ?`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, s := range s {
		_, err := stmt.ExecContext(ctx, query, s.ID, s.Title, s.HV, s.State, s.Status)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteByUser удаляет доступ к серверам у определенного пользователя, возвращает ошибку в случае неудачи.
func (r *ServerRepository) DeleteByUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users_servers WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	return err
}

// Servers возвращает массив из серверов БД или ошибку.
func (r *ServerRepository) Servers(ctx context.Context) (s []entity.Server, err error) {
	q := "SELECT id, title, hv, state, status FROM servers ORDER BY title"

	err = r.db.SelectContext(ctx, &s, q)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// FindByUser возвращает массив серверов пользователя по его ID.
func (r *ServerRepository) FindByUser(ctx context.Context, userID uuid.UUID) (s []entity.Server, err error) {
	q := "SELECT server_id FROM users_servers WHERE user_id = ?"

	err = r.db.SelectContext(ctx, &s, q, userID)
	if err != nil {
		return s, err
	}

	return s, err
}

// SetUserServers добавляет сервера пользователю по его айди
func (r *ServerRepository) SetUserServers(ctx context.Context, userID uuid.UUID, serversIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "DELETE FROM users_servers WHERE user_id = ?"

	_, err = tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user's servers: %w", err)
	}

	query = "INSERT INTO users_servers (user_id, server_id) VALUES(?, ?)"

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	for _, id := range serversIDs {
		_, err := stmt.ExecContext(ctx, userID, id)
		if err != nil {
			return fmt.Errorf("failed to insert user's server: %w", err)
		}
	}

	return tx.Commit()
}
