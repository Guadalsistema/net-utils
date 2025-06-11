package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// OrderedObject is an ordered sequence of name/value members.
type OrderedObject []ObjectMember

// ObjectMember is one member of a JSON object.
type ObjectMember struct {
	Name  string
	Value any
}

// UnmarshalJSON implements json.Unmarshaler, decoding into an OrderedObject.
func (o *OrderedObject) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))

	// 1) consume '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected JSON object start, got %v", t)
	}

	var out OrderedObject
	for dec.More() {
		// 2) read the key
		tk, err := dec.Token()
		if err != nil {
			return err
		}
		key := tk.(string)

		// 3) grab the raw bytes of the next value
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return err
		}

		// 4) dispatch based on the first non-WS byte
		var val any
		switch firstNonWS(raw) {
		case '{':
			var nested OrderedObject
			if err := json.Unmarshal(raw, &nested); err != nil {
				return err
			}
			val = nested

		case '[':
			arr, err := unmarshalArray(raw)
			if err != nil {
				return err
			}
			val = arr

		default:
			// primitive: string, number, bool, or null
			if err := json.Unmarshal(raw, &val); err != nil {
				return err
			}
		}

		out = append(out, ObjectMember{key, val})
	}

	// 5) consume '}'
	if _, err := dec.Token(); err != nil {
		return err
	}

	*o = out
	return nil
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
		keyB, _ := json.Marshal(m.Name)
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

// firstNonWS returns the first non-whitespace byte in data (or 0).
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
