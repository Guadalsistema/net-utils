package database

import (
	"fmt"
	"strings"
)

// ParseDSN takes a DSN string and returns the driver name and the DSN without the scheme.
// If the DSN doesn't have a scheme, it assumes it's for SQLite and returns "sqlite" as the driver.
// Example inputs and outputs:
// - "sqlite://file.db" -> ("sqlite", "file.db")
// - "file.db" -> ("sqlite", "file.db")
// - "sqlite://:memory:" -> ("sqlite", ":memory:")
// - ":memory:" -> ("sqlite", ":memory:")
func ParseDSN(dsn string) (driver string, cleanDSN string) {
	// Check if DSN has a scheme
	if strings.Contains(dsn, "://") {
		parts := strings.SplitN(dsn, "://", 2)
		return parts[0], parts[1]
	}
	// If no scheme, assume SQLite
	return "sqlite", dsn
}

// FormatDSN takes a driver name and DSN and returns a complete DSN with scheme.
// Example inputs and outputs:
// - ("sqlite", "file.db") -> "sqlite://file.db"
// - ("sqlite", ":memory:") -> "sqlite://:memory:"
func FormatDSN(driver string, dsn string) string {
	return fmt.Sprintf("%s://%s", driver, dsn)
}
