package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/anaxita/logit"
	_ "github.com/go-sql-driver/mysql" // ...
)

// Store содержит в себе подключение к базе данных и репозитории
type Store struct {
	db *sql.DB
}

// Connect создает подключение к БД
func Connect(dbtype, user, password, addr, dbname string) (*sql.DB, error) {
	logit.Info("Соединяемся с БД ...")

	db, err := sql.Open(dbtype, fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, addr, dbname))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	logit.Info("Успешно соединились с БД", dbname)

	return db, nil
}

// New создает новый стор с подключением к БД
func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// User возвращает указатель на UserRepository
func (s *Store) User(c context.Context) *UserRepository {
	return &UserRepository{
		db:  s.db,
		ctx: c,
	}
}
