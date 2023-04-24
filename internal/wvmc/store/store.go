package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/anaxita/wvmc/internal/wvmc/hasher"
	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

// Store содержит в себе подключение к базе данных и репозитории
type Store struct {
	db *sqlx.DB
}

// Connect создает подключение к БД
func Connect(scheme, user, password, dbname string) (*sql.DB, error) {
	db, err := sql.Open(scheme,
		fmt.Sprintf("file:%s?_auth&_auth_user=%s&_auth_pass=%s&_auth_crypt=sha512", dbname, user,
			password))
	if err != nil {
		return nil, err
	}

	for i := 0; i < 3; i++ {
		err = db.Ping()
		if err == nil {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
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

// Migrate создает таблицы в БД, если их еще не существует
func Migrate(db *sql.DB) error {
	createUsersTable, _ := os.ReadFile("./migrations/users.migrations")
	createServersTable, _ := os.ReadFile("./migrations/servers.migrations")
	createUsersServersTable, _ := os.ReadFile("./migrations/users_servers.migrations")
	createRefreshTokkensTable, _ := os.ReadFile("./migrations/refresh_tokens.migrations")
	createHypervsTable, _ := os.ReadFile("./migrations/hypervs.migrations")

	_, err := db.Exec(string(createUsersTable))
	if err != nil {
		return err
	}

	password, err := hasher.Hash(os.Getenv("ADMIN_PASSWORD"))
	if err != nil {
		return err
	}

	query := "INSERT OR IGNORE INTO  users (name, email, password, company,  role) VALUES('Администратор', ?, ?, 'Моя компания', 1)"
	_, err = db.Exec(query, os.Getenv("ADMIN_NAME"), string(password))
	if err != nil {
		return err
	}

	_, err = db.Exec(string(createServersTable))
	if err != nil {
		return err
	}

	_, err = db.Exec(string(createUsersServersTable))
	if err != nil {
		return err
	}

	_, err = db.Exec(string(createRefreshTokkensTable))
	if err != nil {
		return err
	}

	_, err = db.Exec(string(createHypervsTable))
	if err != nil {
		return err
	}

	return nil
}
