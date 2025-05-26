package trace

import (
	"context"
)

type traceIdKey struct{} // unexported unique type

func WithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, traceIdKey{}, traceId)
}

func TraceIdFrom(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	m, ok := ctx.Value(traceIdKey{}).(string)
	if !ok || m == "" {
		return "", false
	}
	return m, true
}
