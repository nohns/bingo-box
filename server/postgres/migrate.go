package postgres

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (db *DB) MigrateUp() error {
	// Migrate up
	m, err := migrate.NewWithDatabaseInstance(
		"file://postgres/migration",
		"pgx", db.mgDriver)
	if err != nil {
		return err
	}

	err = m.Up()
	if !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (db *DB) MigrateDown() error {
	// Migrate up
	m, err := migrate.NewWithDatabaseInstance(
		"file://postgres/migration",
		"pgx", db.mgDriver)
	if err != nil {
		return err
	}

	err = m.Down()
	if !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
