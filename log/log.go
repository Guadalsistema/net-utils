package log

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/Guadalsistema/net-utils/trace"
)

func logContext(l *slog.Logger, ctx context.Context, level slog.Level, msg string, args ...any) error {
	if !l.Enabled(ctx, level) {
		return nil
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])

	trace, ok := trace.TraceIdFrom(ctx)
	if !ok {
		return fmt.Errorf("context does not contain trace ID")
	}

	r.Add("trace", trace)

	r.Add(args...)
	err := l.Handler().Handle(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to log message: %w", err)
	}
	return nil
}
func ContextInfo(l *slog.Logger, ctx context.Context, msg string, args ...any) error {
	return logContext(l, ctx, slog.LevelInfo, msg, args...)
}

func ContextWarning(l *slog.Logger, ctx context.Context, msg string, args ...any) error {
	return logContext(l, ctx, slog.LevelWarn, msg, args...)
}

func ContextError(l *slog.Logger, ctx context.Context, msg string, args ...any) error {
	return logContext(l, ctx, slog.LevelError, msg, args...)
}

func ContextDebug(l *slog.Logger, ctx context.Context, msg string, args ...any) error {
	return logContext(l, ctx, slog.LevelDebug, msg, args...)
}
