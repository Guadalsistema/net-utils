package middleware_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"

	"github.com/Guadalsistema/net-utils/middleware"
	"github.com/Guadalsistema/net-utils/trace"
)

func TestLoggingMetaMiddlewareDebug(t *testing.T) {
	/* ---------- capture slog output ---------- */
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelDebug, // capture both Info and Debug lines
	}))
	orig := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(orig) })

	/* ---------- build a next handler that asserts context ---------- */
	var gotTxId string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txId, ok := trace.TraceIdFrom(r.Context())
		if !ok {
			t.Fatalf("meta missing from context")
		}
		if txId == "" {
			t.Fatalf("txIdis empty")
		}
		gotTxId = txId

		// Read the body after middleware processing to ensure it's readable
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		r.Body.Close() // Close the body after reading
		if string(body) != `{foo:bar}` {
			t.Fatalf("expected body to be `{foo:bar}`, got %s", string(body))
		}

		// echo body to ensure response logger sees something
		w.Write([]byte("pong"))
	})

	/* ---------- fire a request through the middleware -------------- */
	reqBody := `{foo:bar}`
	req := httptest.NewRequest(http.MethodPost, "/ping", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	middleware.TraceMiddleware(logger)(next).ServeHTTP(rr, req)

	/* ---------- basic HTTP assertions ------------------------------ */
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	/* ---------- log assertions ------------------------------------- */
	logs := logBuf.String()

	cases := []struct {
		substr string
		desc   string
	}{
		{gotTxId, "tx-id in logs"},
		{"Request", "request/response log line"},
		{reqBody, "request body logged"},
		{"pong", "response body logged"},
	}

	for _, c := range cases {
		if !strings.Contains(logs, c.substr) {
			t.Errorf("missing %s (%q) in logs", c.desc, c.substr)
		}
	}
}

func TestLoggingMetaMiddlewareInfo(t *testing.T) {
	/* ---------- capture slog output ---------- */
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelInfo, // capture both Info and Debug lines
	}))
	orig := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(orig) })

	/* ---------- build a next handler that asserts context ---------- */
	var gotTxId string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txId, ok := trace.TraceIdFrom(r.Context())
		if !ok {
			t.Fatalf("meta missing from context")
		}
		if txId == "" {
			t.Fatalf("txIdis empty")
		}
		gotTxId = txId

		// Read the body after middleware processing to ensure it's readable
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		r.Body.Close() // Close the body after reading
		if string(body) != `{foo:bar}` {
			t.Fatalf("expected body to be `{foo:bar}`, got %s", string(body))
		}

		// echo body to ensure response logger sees something
		w.Write([]byte("pong"))
	})

	/* ---------- fire a request through the middleware -------------- */
	reqBody := `{foo:bar}`
	req := httptest.NewRequest(http.MethodPost, "/ping", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	middleware.TraceMiddleware(logger)(next).ServeHTTP(rr, req)

	/* ---------- basic HTTP assertions ------------------------------ */
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	/* ---------- log assertions ------------------------------------- */
	logs := logBuf.String()

	cases := []struct {
		substr string
		desc   string
	}{
		{gotTxId, "tx-id in logs"},
		{"Request", "request/response log line"},
	}

	for _, c := range cases {
		if !strings.Contains(logs, c.substr) {
			t.Errorf("missing %s (%q) in logs", c.desc, c.substr)
		}
	}
}

func TestMetaMiddleware_ContextHandling(t *testing.T) {
	// ─── setup ────────────────────────────────────────────────────────────────────
	middle := middleware.TraceMiddleware(slog.Default())

	var capturedCtx context.Context // will hold the ctx inside the handler

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()

		txId, ok := trace.TraceIdFrom(capturedCtx)
		if !ok {
			t.Fatal("no meta found in context")
		}
		if txId == "" {
			t.Error("expected TxId to be set")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	initialCtx := req.Context() // for later comparison
	rr := httptest.NewRecorder()

	// ─── execute ─────────────────────────────────────────────────────────────────
	middle(nextHandler).ServeHTTP(rr, req)

	// ─── assertions outside the handler ─────────────────────────────────────────
	// 1. original request remained unchanged
	if req.Context() != initialCtx {
		t.Error("original request context must remain unchanged")
	}

	// 2. handler really received a context with meta
	if capturedCtx == nil {
		t.Fatal("handler context was not captured")
	}
	if txId, ok := trace.TraceIdFrom(capturedCtx); !ok || txId == "" {
		t.Fatal("handler context missing Meta or TxId")
	}

	// 3. response status
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}
