package app

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT"`
	LogFile  string `env:"LOG_FILE"`
	DB       DBConfig
}
type DBConfig struct {
	Name           string `env:"SQLITE_DB"`
	User           string `env:"SQLITE_USER"`
	Password       string `env:"SQLITE_PASSWORD"`
	MigrationsPath string `env:"SQLITE_MIGRATIONS_PATH"`
}

func NewConfig() (*Config, error) {
	var c Config

	err := godotenv.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	return &c, env.Parse(&c, env.Options{
		RequiredIfNoDef: true,
	})
}
