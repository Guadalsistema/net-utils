package utils

import (
	"bytes"
	"encoding/json"
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
