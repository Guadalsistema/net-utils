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
