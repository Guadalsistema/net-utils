package utils_test

import (
	"testing"

	"github.com/Guadalsistema/net-utils/utils"
)

// Example struct for testing
type MyStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestDecodeStrict(t *testing.T) {
	// ✅ Case: valid input
	raw := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	got, err := utils.DecodeStrict[MyStruct](raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Alice" || got.Age != 30 {
		t.Fatalf("unexpected result: %+v", got)
	}

	// ❌ Case: unknown field should fail
	rawBad := map[string]any{
		"name": "Bob",
		"age":  40,
		"foo":  "bar", // not in struct
	}

	_, err = utils.DecodeStrict[MyStruct](rawBad)
	if err == nil {
		t.Fatalf("expected error for unknown field, got nil")
	}
}
