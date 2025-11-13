package migrate

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite" // registers the sqlite driver (pure go)
	_ "github.com/golang-migrate/migrate/v4/source/file"     // registers the file source
)

// newMigrate constructs a *migrate.Migrate for the given SQLite URL and migrations dir.
func newMigrate(databaseURL, migrationsDir string) (*migrate.Migrate, error) {
	// prepend file:// to point at the local folder of .sql files
	sourceURL := "file://" + migrationsDir
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("initializing migrate: %w", err)
	}
	return m, nil
}

// Up applies all up-migrations. ErrNoChange is treated as success.
func Up(databaseURL, migrationsDir string) error {
	m, err := newMigrate(databaseURL, migrationsDir)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("applying up migrations: %w", err)
	}
	return nil
}

// Down rolls back the most recent migration. ErrNoChange is treated as success.
func Down(databaseURL, migrationsDir string) error {
	m, err := newMigrate(databaseURL, migrationsDir)
	if err != nil {
		return err
	}
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("applying down migrations: %w", err)
	}
	return nil
}

// Version returns the current migration version and dirty state.
// If no migration has been applied, version==0 and dirty==false.
func Version(databaseURL, migrationsDir string) (version uint, dirty bool, err error) {
	m, err := newMigrate(databaseURL, migrationsDir)
	if err != nil {
		return 0, false, err
	}
	v, d, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("getting migration version: %w", err)
	}
	return v, d, nil
}
