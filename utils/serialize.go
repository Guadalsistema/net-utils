package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// OrderedObject is an ordered sequence of name/value members.
type OrderedObject []ObjectMember

// ObjectMember is one member of a JSON object.
// TODO conside implement it as generic type ObjectMember[K comparable, V any].
// This would allow us to use any type for the key, not just string.
type ObjectMember struct {
	Key   string
	Value any
}

// getValue returns the value associated with key in an OrderedObject.
func (oo OrderedObject) Get(key string) (any, bool) {
	for _, pair := range oo {
		if pair.Key == key {
			return pair.Value, true
		}
	}
	return nil, false
}

// UnmarshalJSON implements json.Unmarshaler, decoding into an OrderedObject.
func (o *OrderedObject) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))

	// 1) Consume '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected JSON object start, got %v", t)
	}

	var out OrderedObject
	for dec.More() {
		// 2) Read the key
		tk, err := dec.Token()
		if err != nil {
			return err
		}
		key := tk.(string)

		// 3) Grab the raw bytes of the next value
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return err
		}

		// 4) Dispatch based on the first non-WS byte and handle various types
		val, err := determineValueType(raw)
		if err != nil {
			return err
		}

		out = append(out, ObjectMember{key, val})
	}

	// 5) Consume '}'
	if _, err := dec.Token(); err != nil {
		return err
	}

	*o = out
	return nil
}

// determineValueType determines the appropriate Go type for the given raw JSON value.
func determineValueType(raw json.RawMessage) (any, error) {
	// Handling the raw bytes as a string to check for the explicit null value quickly
	if string(raw) == "null" {
		return nil, nil // Handle JSON null directly
	}

	switch firstNonWS(raw) {
	case '{':
		var nested OrderedObject
		if err := json.Unmarshal(raw, &nested); err != nil {
			return nil, err
		}
		return nested, nil

	case '[':
		arr, err := unmarshalArray(raw)
		if err != nil {
			return nil, err
		}
		return arr, nil

	default:
		// Primitive types: identify and unmarshal them
		if result, err := unmarshalPrimitive[int64](raw); err == nil {
			return result, nil
		}
		if result, err := unmarshalPrimitive[float64](raw); err == nil {
			return result, nil
		}
		if result, err := unmarshalPrimitive[bool](raw); err == nil {
			return result, nil
		}
		if result, err := unmarshalPrimitive[string](raw); err == nil {
			return result, nil
		}

		return nil, fmt.Errorf("cannot unmarshal value: %s", raw)
	}
}

// unmarshalPrimitive attempts to decode raw into the type parameter T.
func unmarshalPrimitive[T any](raw json.RawMessage) (T, error) {
	var result T
	err := json.Unmarshal(raw, &result)
	return result, err
}

// Helper: Returns the first non-whitespace byte in a byte slice
func firstNonWS(data []byte) byte {
	for _, b := range data {
		switch b {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return b
		}
	}
	return 0
}

// MarshalJSON implements json.Marshaler, emitting members in insertion order.
func (o OrderedObject) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte('{')
	for i, m := range o {
		if i > 0 {
			buf.WriteByte(',')
		}
		// write the key
		keyB, _ := json.Marshal(m.Key)
		buf.Write(keyB)
		buf.WriteByte(':')

		// write the value (recurse if it's another OrderedObject)
		switch v := m.Value.(type) {
		case OrderedObject:
			b, _ := json.Marshal(v)
			buf.Write(b)
		default:
			b, _ := json.Marshal(v)
			buf.Write(b)
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// unmarshalArray decodes a JSON array into []any, recursing on nested structures.
func unmarshalArray(data []byte) ([]any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))

	// consume '['
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected JSON array start, got %v", t)
	}

	var arr []any
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}

		switch firstNonWS(raw) {
		case '{':
			var nested OrderedObject
			if err := json.Unmarshal(raw, &nested); err != nil {
				return nil, err
			}
			arr = append(arr, nested)

		case '[':
			n2, err := unmarshalArray(raw)
			if err != nil {
				return nil, err
			}
			arr = append(arr, n2)

		default:
			var prim any
			if err := json.Unmarshal(raw, &prim); err != nil {
				return nil, err
			}
			arr = append(arr, prim)
		}
	}

	// consume ']'
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return arr, nil
}
