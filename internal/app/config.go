package app

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT" envDefault:"80"`
	LogFile  string `env:"LOG_FILE" envDefault:"wvmc_default.log"`
	DB       DBConfig
}
type DBConfig struct {
	Name     string `env:"SQLITE_DB"`
	User     string `env:"SQLITE_USER"`
	Password string `env:"SQLITE_PASSWORD"`
}

func NewConfig(envFiles ...string) (*Config, error) {
	var c Config

	err := godotenv.Load(envFiles...)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	return &c, env.Parse(&c, env.Options{
		RequiredIfNoDef: true,
	})
}
