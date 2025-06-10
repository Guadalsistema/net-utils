package database

import (
	"testing"
)

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name           string
		dsn            string
		expectedDriver string
		expectedDSN    string
	}{
		{
			name:           "DSN with scheme",
			dsn:            "sqlite://file.db",
			expectedDriver: "sqlite",
			expectedDSN:    "file.db",
		},
		{
			name:           "DSN without scheme",
			dsn:            "file.db",
			expectedDriver: "sqlite",
			expectedDSN:    "file.db",
		},
		{
			name:           "In-memory DSN with scheme",
			dsn:            "sqlite://:memory:",
			expectedDriver: "sqlite",
			expectedDSN:    ":memory:",
		},
		{
			name:           "In-memory DSN without scheme",
			dsn:            ":memory:",
			expectedDriver: "sqlite",
			expectedDSN:    ":memory:",
		},
		{
			name:           "Complex DSN with scheme",
			dsn:            "sqlite://file::memory:?cache=shared",
			expectedDriver: "sqlite",
			expectedDSN:    "file::memory:?cache=shared",
		},
		{
			name:           "Complex DSN without scheme",
			dsn:            "file::memory:?cache=shared",
			expectedDriver: "sqlite",
			expectedDSN:    "file::memory:?cache=shared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, cleanDSN := ParseDSN(tt.dsn)
			if driver != tt.expectedDriver {
				t.Errorf("ParseDSN() driver = %v, want %v", driver, tt.expectedDriver)
			}
			if cleanDSN != tt.expectedDSN {
				t.Errorf("ParseDSN() cleanDSN = %v, want %v", cleanDSN, tt.expectedDSN)
			}
		})
	}
}

func TestFormatDSN(t *testing.T) {
	tests := []struct {
		name           string
		driver         string
		dsn            string
		expectedResult string
	}{
		{
			name:           "Simple DSN",
			driver:         "sqlite",
			dsn:            "file.db",
			expectedResult: "sqlite://file.db",
		},
		{
			name:           "In-memory DSN",
			driver:         "sqlite",
			dsn:            ":memory:",
			expectedResult: "sqlite://:memory:",
		},
		{
			name:           "Complex DSN",
			driver:         "sqlite",
			dsn:            "file::memory:?cache=shared",
			expectedResult: "sqlite://file::memory:?cache=shared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDSN(tt.driver, tt.dsn)
			if result != tt.expectedResult {
				t.Errorf("FormatDSN() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestDSNRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "Simple DSN",
			dsn:  "sqlite://file.db",
		},
		{
			name: "In-memory DSN",
			dsn:  "sqlite://:memory:",
		},
		{
			name: "Complex DSN",
			dsn:  "sqlite://file::memory:?cache=shared",
		},
		{
			name: "DSN without scheme",
			dsn:  "file.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the DSN
			driver, cleanDSN := ParseDSN(tt.dsn)

			// Format it back
			formattedDSN := FormatDSN(driver, cleanDSN)

			// Parse it again
			finalDriver, finalDSN := ParseDSN(formattedDSN)

			// Verify the round trip
			if finalDriver != driver {
				t.Errorf("Round trip driver mismatch: got %v, want %v", finalDriver, driver)
			}
			if finalDSN != cleanDSN {
				t.Errorf("Round trip DSN mismatch: got %v, want %v", finalDSN, cleanDSN)
			}
		})
	}
}
