package auth

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Secret          string        `env:"AUTH_SECRET"`
	AccessDuration  time.Duration `env:"AUTH_ACCESS_DURATION" envDefault:"10m"`
	RefreshDuration time.Duration `env:"AUTH_REFRESH_DURATION" envDefault:"2160h"`
}

func ParseConfig() (*Config, error) {
	conf, err := env.ParseAs[Config]()
	return &conf, err
}
