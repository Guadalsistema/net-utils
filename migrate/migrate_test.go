package migrate_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/Guadalsistema/net-utils/migrate"
)

// TestUpDown tests the Up and Down functions using an in‑memory SQLite database.
func TestUpDown(t *testing.T) {
	// Use shared in-memory SQLite database.
	dsn := "file::memory:?cache=shared"
	databaseURL := "sqlite://" + dsn
	// For direct db/sql usage, we use the DSN without the scheme.

	// Create a temporary directory to simulate migrations directory.
	tempMigrationsDir, err := os.MkdirTemp("", "migrations")
	if err != nil {
		t.Fatalf("failed to create temp migrations directory: %v", err)
	}
	defer os.RemoveAll(tempMigrationsDir)

	err = os.WriteFile(filepath.Join(tempMigrationsDir, "001_create_seller_table.up.sql"), []byte("CREATE TABLE seller (id INTEGER PRIMARY KEY);"), 0644)
	if err != nil {
		t.Fatalf("failed to create migration file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempMigrationsDir, "001_create_seller_table.down.sql"), []byte("DROP TABLE seller;"), 0644)
	if err != nil {
		t.Fatalf("failed to create migration file: %v", err)
	}

	// Apply Up migrations.
	if err := migrate.Up(databaseURL, tempMigrationsDir); err != nil {
		t.Fatalf("Up() failed: %v", err)
	}

	// Verify that the test table exists.
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='seller'").Scan(&name)
	if err != nil {
		t.Fatalf("expected table 'seller' to exist after Up, but got error: %v", err)
	}

	// Rollback the migration.
	if err := migrate.Down(databaseURL, tempMigrationsDir); err != nil {
		t.Fatalf("Down() failed: %v", err)
	}

	// Verify that the test table no longer exists.
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='seller'").Scan(&name)
	if err == nil {
		t.Fatalf("expected table 'seller' to be dropped after Down")
	}
}

// TestNewMigrateRelativePath ensures that passing a relative migrations directory
// to the public Up/Down functions works because newMigrate converts it to an
// absolute path internally.
func TestNewMigrateRelativePath(t *testing.T) {
	// Use shared in-memory SQLite database.
	dsn := "file::memory:?cache=shared"
	databaseURL := "sqlite://" + dsn

	// Create a temporary root directory and a nested migrations directory.
	rootDir, err := os.MkdirTemp("", "rootdir")
	if err != nil {
		t.Fatalf("failed to create temp root dir: %v", err)
	}
	defer os.RemoveAll(rootDir)

	migrationsDir := filepath.Join(rootDir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("failed to create migrations dir: %v", err)
	}

	// Write a simple up/down pair into the migrations directory.
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_create_seller_table.up.sql"), []byte("CREATE TABLE seller (id INTEGER PRIMARY KEY);"), 0644); err != nil {
		t.Fatalf("failed to create migration file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_create_seller_table.down.sql"), []byte("DROP TABLE seller;"), 0644); err != nil {
		t.Fatalf("failed to create migration file: %v", err)
	}

	// Change working directory to rootDir so that a relative path "migrations"
	// resolves against it. This simulates a caller passing a relative path.
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(rootDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Use the relative path "migrations" when calling Up/Down.
	relPath := "migrations"

	if err := migrate.Up(databaseURL, relPath); err != nil {
		t.Fatalf("Up() with relative path failed: %v", err)
	}

	// Open DB directly and verify the table exists.
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='seller'").Scan(&name)
	if err != nil {
		t.Fatalf("expected table 'seller' to exist after Up with relative path, but got error: %v", err)
	}

	// Now rollback using the same relative path.
	if err := migrate.Down(databaseURL, relPath); err != nil {
		t.Fatalf("Down() with relative path failed: %v", err)
	}

	// Verify the table is gone.
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='seller'").Scan(&name)
	if err == nil {
		t.Fatalf("expected table 'seller' to be dropped after Down with relative path")
	}
}
