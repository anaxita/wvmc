package app

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func UpMigrations(conn *sql.DB, dbName, migrationsPath string) error {
	driver, err := sqlite3.WithInstance(conn, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		dbName, driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == nil {
		return nil
	}

	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	return err
}

func NewSQLite3Client(c DBConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("file:%s.db?_auth&_auth_user=%s&_auth_pass=%s&_auth_crypt=sha512",
		c.Name, c.User, c.Password)

	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	for i := 0; i < 6; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}

	if err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
