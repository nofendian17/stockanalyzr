package migration

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunUp applies all pending migrations.
func RunUp(dsn string, migrationsPath string) error {
	m, err := newMigrate(dsn, migrationsPath)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration: up failed: %w", err)
	}

	return nil
}

// RunDown rolls back all migrations.
func RunDown(dsn string, migrationsPath string) error {
	m, err := newMigrate(dsn, migrationsPath)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration: down failed: %w", err)
	}

	return nil
}

func newMigrate(dsn string, migrationsPath string) (*migrate.Migrate, error) {
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return nil, fmt.Errorf("migration: failed to create instance: %w", err)
	}

	return m, nil
}
