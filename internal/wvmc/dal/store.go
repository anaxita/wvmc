package dal

import (
	"context"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

// Store содержит в себе подключение к базе данных и репозитории
type Store struct {
	db *sqlx.DB
}

// New создает новый стор с подключением к БД
func New(db *sqlx.DB) *Store {
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

// Server возвращает указатель на ServerRepository
func (s *Store) Server(c context.Context) *ServerRepository {
	return &ServerRepository{
		db:  s.db,
		ctx: c,
	}
}
