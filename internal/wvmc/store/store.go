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

// Migrate создает таблицы в БД, если их еще не существует
func Migrate(db *sql.DB) error {
	createUsersTable := `CREATE TABLE IF NOT EXISTS users (
		id int unsigned NOT NULL AUTO_INCREMENT,
		name varchar(255) NOT NULL,
		email varchar(255) NOT NULL,
		company varchar(255) NOT NULL,
		role int NOT NULL,
		password text NOT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY email (email) USING BTREE
	  ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`

	createServersTable := `CREATE TABLE IF NOT EXISTS servers (
		id varchar(255) NOT NULL,
		title varchar(255) NOT NULL,
		hv varchar(255) NOT NULL,
		ip4 varchar(255) NOT NULL,
		user_name varchar(255) NOT NULL,
		user_password varchar(255) NOT NULL,
		PRIMARY KEY (id)
	  ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`

	createUsersServersTable := `CREATE TABLE IF NOT EXISTS users_servers (
		user_id int NOT NULL,
		server_id varchar(255) NOT NULL,
		KEY user_id (user_id),
		KEY server_id (server_id) USING BTREE
	) ENGINE = InnoDB DEFAULT CHARSET = utf8;`

	createRefreshTokkensTable := `CREATE TABLE IF NOT EXISTS refresh_tokens (
		user_id int NOT NULL,
		token text NOT NULL,
		PRIMARY KEY (user_id),
		UNIQUE KEY user_id (user_id)
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8;`

	_, err := db.Exec(fmt.Sprintf("%s", createUsersTable))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("%s", createServersTable))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("%s", createUsersServersTable))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("%s", createRefreshTokkensTable))
	if err != nil {
		return err
	}

	return nil
}
