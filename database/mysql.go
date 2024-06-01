package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"
)

var _ SqlDatabaseManager = (*MySQLManager)(nil)

type MySQLManager struct {
	cfg *Config
}

func NewMySQLManager(cfg *Config) (*MySQLManager, error) {
	return &MySQLManager{
		cfg: cfg,
	}, nil
}

func (m *MySQLManager) Create() error {
	db, err := m.open(false)

	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec("CREATE DATABASE " + m.cfg.Name)

	return err
}

func (m *MySQLManager) Drop() error {

	db, err := m.open(false)

	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec("DROP DATABASE " + m.cfg.Name)

	return err
}

func (m *MySQLManager) Db() (*sql.DB, error) {
	return m.open(true)
}

func (m *MySQLManager) Bun() (*bun.DB, error) {

	sqldb, err := m.open(true)

	if err != nil {
		return nil, err
	}

	bunDb := bun.NewDB(sqldb, mysqldialect.New())

	if m.cfg.Debug {
		bunDb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	return bunDb, nil
}

func (m *MySQLManager) open(withDb bool) (*sql.DB, error) {

	connString := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		m.cfg.User,
		m.cfg.Pass,
		m.cfg.Host,
		m.cfg.Port,
		m.cfg.Name,
	)

	if !withDb {
		connString = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/",
			m.cfg.User,
			m.cfg.Pass,
			m.cfg.Host,
			m.cfg.Port,
		)
	}

	return sql.Open("mysql", connString)
}
