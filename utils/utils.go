package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
)

/* -------------------------------------------------------------------------- */
/*  Utilities                                                                 */
/* -------------------------------------------------------------------------- */

func TruncateString(s string, max int64) string {
	if int64(len(s)) > max {
		return s[:max] + "…(truncated)"
	}
	return s
}

// CappedWriter is a wrapper around an io.Writer that limits the amount of data
// that can be written. It discards any data that exceeds the specified cap.
type CappedWriter struct {
	io.Writer
	Remain int64 // the Remaining number of bytes that can be written
}

// Write writes data to the underlying io.Writer but only allows a certain
// number of bytes to be written before discarding the rest. It returns the
// number of bytes that were intended to be written and any error encountered.
func (c *CappedWriter) Write(p []byte) (int, error) {
	if c.Remain <= 0 {
		return len(p), nil // discard once the budget is exhausted
	}
	if int64(len(p)) > c.Remain {
		p = p[:c.Remain]
	}
	n, err := c.Writer.Write(p)
	c.Remain -= int64(n)
	return len(p), err
}

// ResponseRecorder lets us capture body & status that downstream writes.
type ResponseRecorder struct {
	http.ResponseWriter
	Status int
	Buf    bytes.Buffer
}

func (rr *ResponseRecorder) WriteHeader(code int) {
	rr.Status = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *ResponseRecorder) Write(p []byte) (int, error) {
	rr.Buf.Write(p) // copy to our buffer
	return rr.ResponseWriter.Write(p)
}

func RandomKey(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // randomness failure = fatal
	}
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	for i, v := range b {
		b[i] = letters[int(v)%len(letters)]
	}
	return string(b)
}

func DecodeStrict[T any](raw any) (T, error) {
	var zero T
	// Convert the result to AccessToken
	resultJSON, err := json.Marshal(raw)
	if err != nil {
		return zero, err
	}

	decoder := json.NewDecoder(bytes.NewBuffer(resultJSON))
	decoder.DisallowUnknownFields()

	var t T
	if err := decoder.Decode(&t); err != nil {
		return zero, err
	}

	return t, nil
}
