package server

import "github.com/caarlos0/env/v11"

type Config struct {
	Host string `env:"SERVER_HOST" envDefault:"127.0.0.1"`
	Port string `env:"SERVER_PORT" envDefault:"8080"`
}

func ParseConfig() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	return &cfg, err
}
