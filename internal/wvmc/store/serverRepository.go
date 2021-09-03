package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/model"
)

// ServerRepository - содержит методы работы с пользовательскими моделями
type ServerRepository struct {
	db  *sql.DB
	ctx context.Context
}

//Find ищет первое совпадение сервер с заданным ключом и значением, возвращает модель либо ошибку
func (r *ServerRepository) Find(key, value string) (model.Server, error) {
	logit.Info("Ищем сервер:", key, value)

	s := model.Server{}

	query := fmt.Sprintf("SELECT id, title, ip4, hv, company, out_addr, description, user_name, user_password FROM servers WHERE %s = ?", key)
	if err := r.db.QueryRowContext(r.ctx, query, value).Scan(
		&s.ID,
		&s.Name,
		&s.IP,
		&s.HV,
		&s.OutAddr,
		&s.Company,
		&s.Description,
		&s.User,
		&s.Password,
	); err != nil {
		return s, err
	}

	logit.Info("Нашли сервер:", key, value)
	return s, nil
}

//Find ищет сервер по хв + имени, возвращает модель либо ошибку
func (r *ServerRepository) FindByHVandName(hv, name string) (model.Server, error) {
	logit.Info("Ищем сервер:", name, hv)

	s := model.Server{}

	query := "SELECT ip4, user_name, user_password FROM servers WHERE hv = ? AND title = ?"
	if err := r.db.QueryRowContext(r.ctx, query, hv, name).Scan(
		&s.IP,
		&s.User,
		&s.Password,
	); err != nil {
		return s, err
	}

	logit.Info("Нашли сервер:", name, hv)
	return s, nil
}

// Create создает сервер и возвращает его ID, либо ошибку
func (r *ServerRepository) Create(s model.Server) (int, error) {
	logit.Info("Создааем сервер:", s.Name)

	query := "INSERT INTO servers (id, title, ip4, hv, company, user_name, user_password) VALUES (?, ?, ?, ?, ?, ?, ?) on conflict (id) DO UPDATE SET title = ?, ip4 = ?, hv = ?;"

	result, err := r.db.ExecContext(r.ctx, query, s.ID, s.Name, s.IP, s.HV, s.Company, s.User, s.Password, s.Name, s.IP, s.HV)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Edit обновляет данные сервер, возвращает ошибку в случае неудачи
func (r *ServerRepository) Edit(s model.Server) error {
	logit.Info("Обновляем поля серверу:", s.Name)

	var query string
	var err error

	if s.Password != "" {
		query = "UPDATE servers SET company = ?,  description = ?, out_addr = ?, user_name = ?, user_password = ? WHERE id = ? AND hv = ?"
		_, err = r.db.ExecContext(r.ctx, query, s.Company, s.Description, s.OutAddr, s.User, s.Password, s.ID, s.HV)
	} else {
		query = "UPDATE servers SET company = ?,  description = ?, out_addr = ?, user_name = ? WHERE id = ? AND hv = ?"
		_, err = r.db.ExecContext(r.ctx, query, s.Company, s.Description, s.OutAddr, s.User, s.ID, s.HV)
	}

	if err != nil {
		return err
	}

	logit.Info("Успешно обновили поля сервер", s.Name)
	return nil
}

// Delete удаляет сервер, возвращает ошибку в случае неудачи
func (r *ServerRepository) Delete(id string) error {
	logit.Info("Удаляем сервер", id)
	_, err := r.db.ExecContext(r.ctx, "DELETE FROM servers WHERE id = ? LIMIT 1", id)
	if err != nil {
		return err
	}
	logit.Info("Успешно удалили сервер", id)
	return nil
}

// DeleteByUser удаляет доступ к серверам у определенного пользователя, возвращает ошибку в случае неудачи
func (r *ServerRepository) DeleteByUser(userID string) error {
	logit.Info("Удаляем сервера у пользователя", userID)
	_, err := r.db.ExecContext(r.ctx, "DELETE FROM users_servers WHERE user_id = ?", userID)
	if err != nil {
		return err
	}
	logit.Info("Успешно удалили сервера у пользователя", userID)
	return nil
}

// All возвращает массив из серверов БД или ошибку
func (r *ServerRepository) All() ([]model.Server, error) {
	logit.Info("Получаем все сервера")
	var servers []model.Server

	rows, err := r.db.QueryContext(r.ctx, "SELECT id, title, ip4, hv, company, user_name, user_password FROM servers")
	if err != nil {
		return servers, err
	}
	defer rows.Close()

	for rows.Next() {
		var s model.Server
		err := rows.Scan(&s.ID, &s.Name, &s.IP, &s.HV, &s.Company, &s.User, &s.Password)
		if err != nil {
			return servers, err
		}
		servers = append(servers, s)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	logit.Info("Успешно получили все сервера")
	return servers, nil
}

// FindByUser возвращает массив серверов пользователя по его ID
func (r *ServerRepository) FindByUser(userID string) ([]model.Server, error) {
	logit.Info("Получаем все сервера пользователя", userID)
	var servers []model.Server

	rows, err := r.db.QueryContext(r.ctx, "SELECT s.id, s.title, s.ip4, s.hv, s.company, s.description, s.out_addr, s.user_name, s.user_password FROM servers as s INNER JOIN users_servers as us ON (s.id = us.server_ID) WHERE us.user_id = ?", userID)
	if err != nil {
		return servers, err
	}
	defer rows.Close()

	for rows.Next() {
		var s model.Server
		err := rows.Scan(&s.ID, &s.Name, &s.IP, &s.HV, &s.Company, &s.Description, &s.OutAddr, &s.User, &s.Password)
		if err != nil {
			return servers, err
		}
		servers = append(servers, s)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	logit.Info("Успешно получили все сервера пользователя", userID)
	return servers, nil
}
