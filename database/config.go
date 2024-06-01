package database

import "github.com/caarlos0/env/v11"

type Config struct {
	Dialect string `env:"DB_DIALECT" envDefault:"mysql"`
	Host    string `env:"DB_HOST" envDefault:"localhost"`
	Port    string `env:"DB_PORT" envDefault:"3306"`
	User    string `env:"DB_USER" envDefault:"root"`
	Pass    string `env:"DB_PASS" envDefault:"root"`
	Name    string `env:"DB_NAME" envDefault:"mo"`
	Debug   bool   `env:"DB_DEBUG" envDefault:"false"`
}

func ParseConfig() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	return &cfg, err
}
