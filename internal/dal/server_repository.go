package dal

import (
	"context"
	"database/sql"
	"errors"

	entity2 "github.com/anaxita/wvmc/internal/entity"
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

func (r *ServerRepository) FindByTitle(ctx context.Context, title string) (s entity2.Server, err error) {
	query := "SELECT id, vmid, title, ip4, hv, company, out_addr, description, user_name, user_password FROM servers WHERE title = ?"

	err = r.db.GetContext(ctx, &s, query, title)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s, entity2.ErrNotFound
		}

		return s, err
	}

	return s, nil
}

func (r *ServerRepository) FindByID(ctx context.Context, id int64) (s entity2.Server, err error) {
	q := "SELECT id, vmid, title, ip4, hv, company, out_addr, description, user_name, user_password FROM servers WHERE id = ?"

	err = r.db.GetContext(ctx, &s, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s, entity2.ErrNotFound
		}

		return s, err
	}

	return s, nil
}

// FindByHvAndTitle ищет сервер по местоположению и имени, возвращает модель либо ошибку.
func (r *ServerRepository) FindByHvAndTitle(ctx context.Context, hv, name string) (s entity2.Server, err error) {
	query := "SELECT id, vmid, title, ip4, hv, company, out_addr, description, user_name, user_password FROM servers WHERE hv = ? AND title = ?"
	err = r.db.GetContext(ctx, &s, query, hv, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s, entity2.ErrNotFound
		}

		return s, err
	}

	return s, nil
}

// Upsert создает сервер и возвращает его ID, либо ошибку.
func (r *ServerRepository) Upsert(ctx context.Context, s entity2.Server) (int64, error) {
	query := `INSERT INTO servers (vmid, title, ip4, hv, company, user_name, user_password) 
    VALUES (?, ?, ?, ?, ?, ?, ?) 
    ON CONFLICT (title, hv) 
        DO UPDATE SET title = ?, ip4 = ?, hv = ?`

	result, err := r.db.ExecContext(
		ctx, query, s.VMID, s.Name, s.IP, s.HV, s.Company, s.User, s.Password, s.Name, s.IP, s.HV)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
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
func (r *ServerRepository) Servers(ctx context.Context) (s []entity2.Server, err error) {
	q := "SELECT id, vmid, title, ip4, hv, company, user_name, user_password FROM servers"

	err = r.db.SelectContext(ctx, &s, q)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// FindByUser возвращает массив серверов пользователя по его ID.
func (r *ServerRepository) FindByUser(ctx context.Context, userID int64) (s []entity2.Server, err error) {
	q := `
	SELECT 
		s.id, 
		s.vmid,
		s.title,
		s.ip4, 
		s.hv, 
		s.company, 
		s.description, 
		s.out_addr,
		s.user_name,
		s.user_password 
	FROM servers AS s 
	INNER JOIN users_servers AS us 
		ON (s.id = us.server_id) 
	WHERE us.user_id = ?`

	err = r.db.SelectContext(ctx, &s, q, userID)
	if err != nil {
		return s, err
	}

	return s, err
}

// AddServersToUser добавляет сервера пользователю по его айди
func (r *ServerRepository) AddServersToUser(ctx context.Context, userID int64, serversIDs []int64) error {
	// TODO remove old servers in tx

	query := "INSERT INTO users_servers (user_id, server_id) VALUES(?, ?)"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	for _, id := range serversIDs {
		_, err := stmt.ExecContext(ctx, userID, id)
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
