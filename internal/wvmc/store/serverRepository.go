package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/anaxita/wvmc/internal/wvmc/model"
	"github.com/jmoiron/sqlx"
)

// ServerRepository - содержит методы работы с пользовательскими моделями.
type ServerRepository struct {
	db  *sqlx.DB
	ctx context.Context
}

// Find ищет первое совпадение сервер с заданным ключом и значением, возвращает модель либо ошибку.
func (r *ServerRepository) Find(key, value interface{}) (model.Server, error) {
	var s model.Server

	query := fmt.Sprintf(
		"SELECT id, vmid, title, ip4, hv, company, out_addr, description, user_name, user_password FROM servers WHERE %s = ?",
		key)

	if err := r.db.QueryRowContext(r.ctx, query, value).Scan(
		&s.ID,
		&s.VMID,
		&s.Name,
		&s.IP,
		&s.HV,
		&s.OutAddr,
		&s.Company,
		&s.Description,
		&s.User,
		&s.Password,
	); err != nil {
		return s, errors.New(err.Error())
	}

	return s, nil
}

// FindByHvAndName ищет сервер по местоположению и имени, возвращает модель либо ошибку.
func (r *ServerRepository) FindByHvAndName(hv, name string) (model.Server, error) {
	var s model.Server

	query := "SELECT ip4, user_name, user_password FROM servers WHERE hv = ? AND title = ?"
	if err := r.db.QueryRowContext(r.ctx, query, hv, name).Scan(
		&s.IP,
		&s.User,
		&s.Password,
	); err != nil {
		return s, err
	}

	return s, nil
}

// Create создает сервер и возвращает его ID, либо ошибку.
func (r *ServerRepository) Create(s model.Server) (int, error) {
	query := `INSERT INTO servers (vmid, title, ip4, hv, company, user_name, user_password) 
    VALUES (?, ?, ?, ?, ?, ?, ?) 
    ON CONFLICT (title, hv) 
        DO UPDATE SET title = ?, ip4 = ?, hv = ?`

	result, err := r.db.ExecContext(
		r.ctx, query, s.VMID, s.Name, s.IP, s.HV, s.Company, s.User, s.Password, s.Name, s.IP, s.HV)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// DeleteByUser удаляет доступ к серверам у определенного пользователя, возвращает ошибку в случае неудачи.
func (r *ServerRepository) DeleteByUser(userID string) error {
	_, err := r.db.ExecContext(r.ctx, "DELETE FROM users_servers WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	return err
}

// All возвращает массив из серверов БД или ошибку.
func (r *ServerRepository) All() ([]model.Server, error) {
	var servers []model.Server

	rows, err := r.db.QueryContext(r.ctx,
		"SELECT id, vmid, title, ip4, hv, company, user_name, user_password FROM servers")
	if err != nil {
		return servers, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	for rows.Next() {
		var s model.Server
		err := rows.Scan(&s.ID, &s.VMID, &s.Name, &s.IP, &s.HV, &s.Company, &s.User, &s.Password)
		if err != nil {
			return servers, err
		}
		servers = append(servers, s)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return servers, nil
}

// FindByUser возвращает массив серверов пользователя по его ID.
func (r *ServerRepository) FindByUser(userID string) ([]model.Server, error) {
	var servers []model.Server

	rows, err := r.db.QueryContext(r.ctx,
		"SELECT s.id, s.vmid, s.title, s.ip4, s.hv, s.company, s.description, s.out_addr, s.user_name, s.user_password FROM servers AS s INNER JOIN users_servers AS us ON (s.id = us.server_id) WHERE us.user_id = ?",
		userID)
	if err != nil {
		return servers, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	for rows.Next() {
		var s model.Server
		err := rows.Scan(&s.ID, &s.VMID, &s.Name, &s.IP, &s.HV, &s.Company, &s.Description,
			&s.OutAddr, &s.User, &s.Password)

		if err != nil {
			return servers, err
		}

		servers = append(servers, s)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return servers, err
}
