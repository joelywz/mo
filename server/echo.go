package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

func New() *echo.Echo {
	return echo.New()
}

func Start(lc fx.Lifecycle, e *echo.Echo, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {

			slog.Info("starting", "host", cfg.Host, "port", cfg.Port)

			go e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
			return nil
		},
		OnStop: func(context.Context) error {
			return e.Close()
		},
	})
}
