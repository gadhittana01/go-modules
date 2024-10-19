package utils

import (
	"database/sql"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func RunMigrationPool(db *sql.DB, config *BaseConfig) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE EXTENSION pgcrypto")
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(config.MigrationURL, config.DBName, driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
