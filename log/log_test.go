package log_test

import (
	"bytes"
	"context"
	"testing"

	"log/slog"

	"github.com/Guadalsistema/net-utils/log"
	"github.com/Guadalsistema/net-utils/trace"
)

func TestInfoContext(t *testing.T) {
	// Capture slog output
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Capture both Info and Debug lines
	}))
	orig := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(orig) })

	// Create a context with metadata (simulated)
	ctxWithTraceId := context.Background()
	ctxWithTraceId = trace.WithTraceId(ctxWithTraceId, "123-abc") // Simulate middleware context

	// Test logging with metadata
	log.ContextInfo(logger, ctxWithTraceId, "Test message", "key1", "value1")

	// Verify the captured output
	output := logBuf.String()
	if !bytes.Contains(logBuf.Bytes(), []byte("Test message")) {
		t.Errorf("expected log to contain 'Test message', got: %s", output)
	}
	if !bytes.Contains(logBuf.Bytes(), []byte("trace=123-abc")) {
		t.Errorf("expected log to contain 'trace=123-abc', got: %s", output)
	}
	if !bytes.Contains(logBuf.Bytes(), []byte("key1=value1")) {
		t.Errorf("expected log to contain 'key1=value1', got: %s", output)
	}
}
