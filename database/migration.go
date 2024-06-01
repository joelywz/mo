package database

import (
	"context"

	mysqlerrnum "github.com/bombsimon/mysql-error-numbers/v2"
	"github.com/uptrace/bun/migrate"
)

var (
	DefaultDirectory = "internal/database/migration"
)

func CreateMigration(name string) error {
	migrations := migrate.NewMigrations(migrate.WithMigrationsDirectory(DefaultDirectory))

	migrator := migrate.NewMigrator(nil, migrations)

	_, err := migrator.CreateGoMigration(context.Background(), name)

	if err != nil {
		return err
	}

	return nil
}

func MigrateUp(manager SqlDatabaseManager, migrations *migrate.Migrations) error {

	db, err := manager.Bun()

	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(db, migrations)

	if err := migrator.Init(context.Background()); err != nil {
		if mysqlerrnum.FromError(err) != mysqlerrnum.ErrBadDbError {
			return err
		}

		if err := manager.Create(); err != nil {
			return err
		}

		if err := migrator.Init(context.Background()); err != nil {
			return err
		}
	}

	_, err = migrator.Migrate(context.Background())

	return err
}
