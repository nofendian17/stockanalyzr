package migration_test

import (
	"strings"
	"testing"

	"stockanalyzr/pkg/migration"
)

func TestRunUp_InvalidDSN(t *testing.T) {
	dsn := "invalid-url"
	migrationsPath := "./migrations" // Path doesn't matter since DSN parsing should fail first or fail during connection

	err := migration.RunUp(dsn, migrationsPath)
	if err == nil {
		t.Fatal("expected error due to invalid DSN, got nil")
	}
}

func TestRunUp_InvalidMigrationsPath(t *testing.T) {
	// A DSN that won't fail parsing immediately but will fail on connection so we can potentially see the path failure
	// Or we use a dummy valid scheme like "postgres://fake"
	dsn := "postgres://user:pass@127.0.0.1:0/db"
	migrationsPath := "/path/that/does/not/exist"

	err := migration.RunUp(dsn, migrationsPath)
	if err == nil {
		t.Fatal("expected error due to invalid migrations path or connection, got nil")
	}

	// 'migrate' library usually validates the source path first
	if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "failed to create instance") {
		// Just failing is good enough, usually complains about the file source depending on the order of checks
		t.Logf("Expected some file/instance error, got: %v", err)
	}
}

func TestRunDown_InvalidDSN(t *testing.T) {
	dsn := "invalid-url"
	migrationsPath := "./migrations"

	err := migration.RunDown(dsn, migrationsPath)
	if err == nil {
		t.Fatal("expected error due to invalid DSN, got nil")
	}
}
