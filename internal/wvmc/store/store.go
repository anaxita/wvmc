package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/hasher"

	// _ "github.com/go-sql-driver/mysql" // ...
	_ "github.com/mattn/go-sqlite3"
)

// Store содержит в себе подключение к базе данных и репозитории
type Store struct {
	db *sql.DB
}

// Connect создает подключение к БД
func Connect(dbtype, user, password, addr, dbname string) (*sql.DB, error) {
	logit.Info("Соединяемся с БД ...")

	// mysql
	// db, err := sql.Open(dbtype, fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, addr, dbname))

	// sqllite3
	db, err := sql.Open(dbtype, fmt.Sprintf("file:%s?_auth&_auth_user=%s&_auth_pass=%s&_auth_crypt=sha512", dbname, user, password))
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

// Server возвращает указатель на ServerRepository
func (s *Store) Server(c context.Context) *ServerRepository {
	return &ServerRepository{
		db:  s.db,
		ctx: c,
	}
}

// Migrate создает таблицы в БД, если их еще не существует
func Migrate(db *sql.DB) error {
	logit.Info("Выполняем миграции ...")

	createUsersTable, _ := os.Read("./sql/users.sql")
	createServersTable, _ := os.ReadFile("./sql/servers.sql")
	createUsersServersTable, _ := os.ReadFile("./sql/users_servers.sql")
	createRefreshTokkensTable, _ := os.ReadFile("./sql/refresh_tokens.sql")
	createHypervsTable, _ := os.ReadFile("./sql/hypervs.sql")

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

	logit.Info("Миграции выполнены успешно")
	return nil
}
