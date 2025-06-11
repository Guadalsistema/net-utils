package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"testing"
)

// TestOrderedObject_MarshalOrder checks that MarshalJSONTo emits keys
// in exactly the insertion order, even when a nested object is involved.
func TestOrderedObject_MarshalOrder(t *testing.T) {
	tests := []struct {
		name string
		obj  OrderedObject
		want string
	}{
		{
			name: "1,2,3,4",
			obj: OrderedObject{
				{"1", "v1"},
				{"2", OrderedObject{
					{"a", 1},
				}},
				{"3", "v3"},
				{"4", "v4"},
			},
			want: `{"1":"v1","2":{"a":1},"3":"v3","4":"v4"}`,
		},
		{
			name: "1,3,2,4",
			obj: OrderedObject{
				{"1", "v1"},
				{"3", "v3"},
				{"2", OrderedObject{
					{"a", 1},
				}},
				{"4", "v4"},
			},
			want: `{"1":"v1","3":"v3","2":{"a":1},"4":"v4"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal with AllowDuplicateNames so we use our MarshalJSONTo
			b, err := json.Marshal(&tt.obj)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			got := string(bytes.TrimSpace(b))
			if got != tt.want {
				t.Errorf("unexpected JSON order:\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestOrderedObjectUnmarshal(t *testing.T) {
	jsonData := []byte(fmt.Sprintf(`{
		"int64": %d,
		"float64": %g,
		"string": "%s",
		"bool": %t,
		"null": %s
	}`, 42, math.SmallestNonzeroFloat64, "hello", true, "null"))

	var orderedObj OrderedObject
	err := orderedObj.UnmarshalJSON(jsonData)
	if err != nil {
		t.Fatalf("Unmarshaling failed: %v", err)
	}

	testCases := []struct {
		key          string
		expectedType string // expected type as a formatted string
	}{
		{"int64", "int64"},
		{"float64", "float64"}, // Should be unmarshaled as float64
		{"string", "string"},   // Should remain a string
		{"bool", "bool"},       // Should remain a bool
		{"null", "nil"},        // Should be nil
	}

	for _, tc := range testCases {
		val, found := orderedObj.Get(tc.key)
		if !found {
			t.Fatalf("Key %s not found", tc.key)
		}

		if val == nil {
			if tc.expectedType != "nil" {
				t.Errorf("Expected %s for key %s, but got nil", tc.expectedType, tc.key)
			}
			continue
		}

		actualType := fmt.Sprintf("%T", val)

		if actualType != tc.expectedType {
			t.Errorf("For key %s, expected type %s, but got %s", tc.key, tc.expectedType, actualType)
		}
	}
}
