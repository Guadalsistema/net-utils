package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Guadalsistema/net-utils/log"
	"github.com/Guadalsistema/net-utils/trace"
	"github.com/Guadalsistema/net-utils/utils"
)

/* -------------------------------------------------------------------------- */

// decompressGzip function to handle gzip decompression
func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// maxBodyLog limits how much of the body we copy for logging.
// You can override it before you register the middleware.
var maxBodyLog int64 = 1 << 20 // 1 MiB
// ctxKeyRawBody is an unexported context key.

// TraceMiddleware returns a fully-formed http.Handler middleware.
// It captures the request body and response body, and logs them.
// It also creates a unique transaction Idfor the request available in the request context.
func TraceMiddleware(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := l
			start := time.Now()
			/* ---------- advance work: Tx-Id+ capture body ---------- */
			txId := utils.RandomKey(8)

			// Create new request with modified context
			newReq := r.WithContext(trace.WithTraceId(r.Context(), txId))
			resp := &utils.ResponseRecorder{ResponseWriter: w, Status: http.StatusOK}
			ctx := newReq.Context()

			/* ----------- capture request body via TeeReader ----------- */
			var logbuf bytes.Buffer
			tee := io.TeeReader(newReq.Body, &utils.CappedWriter{Writer: &logbuf, Remain: maxBodyLog})
			newReq.Body = io.NopCloser(tee)

			if err := log.ContextDebug(logger, ctx, "Request", "Url", newReq.URL.String(), "method", newReq.Method); err != nil {
				slog.ErrorContext(ctx, "Failed to log request", "error", err)
			}

			if logger.Enabled(ctx, slog.LevelDebug) {
				payload, err := io.ReadAll(newReq.Body)
				if err != nil {
					log.ContextError(logger, ctx, "Failed to read request body", "error", err)
					http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				// Read request for log
				log.ContextDebug(logger, ctx, "Request body", "method", newReq.Method, "size", logbuf.Len(), log.LogHeaders(newReq.Header), "body", utils.TruncateString(logbuf.String(), maxBodyLog))
				newReq.Body = io.NopCloser(bytes.NewReader(payload)) // volver a ponerlo
			}

			if next == nil {
				log.ContextError(logger, ctx, "Server error", "error", "next HTTP handler is nil.")
				http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if newReq.Body == nil {
				log.ContextError(logger, ctx, "Server error", "error", "Request body is nil.")
				http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			w.Header().Set("X-Tx-Id", txId) // Set transaction ID in response header

			// Use the new request with modified context
			next.ServeHTTP(resp, newReq)

			/* ---------- log outgoing response ---------- */
			elapsed := time.Since(start)
			log.ContextInfo(logger, newReq.Context(), "Response", "Url", newReq.URL.String(), "method", newReq.Method, "status", resp.Status, "size", resp.Buf.Len(), "elapsed", elapsed)
			if l.Enabled(ctx, slog.LevelDebug) {
				if resp.Header().Get("Content-Encoding") == "gzip" {
					if decompressedBody, err := decompressGzip(resp.Buf.Bytes()); err == nil {
						log.ContextDebug(logger, newReq.Context(), "Response body", "size", len(decompressedBody), log.LogHeaders(resp.Header()), "body", utils.TruncateString(string(decompressedBody), maxBodyLog))
					} else {
						log.ContextError(logger, ctx, "Failed to decompress response body", "error", err)
					}
				} else {
					// Normal logging for non-gzipped responses
					log.ContextDebug(logger, newReq.Context(), "Response body", "size", resp.Buf.Len(), log.LogHeaders(resp.Header()), "body", utils.TruncateString(resp.Buf.String(), maxBodyLog))
				}
			}
		})
	}
}
