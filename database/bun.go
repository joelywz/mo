package database

import (
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
)

type SqlDatabaseManager interface {
	Create() error
	Drop() error
	Db() (*sql.DB, error)
	Bun() (*bun.DB, error)
}

func NewManager(cfg *Config) (SqlDatabaseManager, error) {

	manager, err := getSqlManager(cfg)

	if err != nil {
		return nil, err
	}

	return manager, nil
}

func getSqlManager(cfg *Config) (SqlDatabaseManager, error) {
	switch cfg.Dialect {
	case "mysql":
		return NewMySQLManager(cfg)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", cfg.Dialect)
	}
}
