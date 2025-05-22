package utils

import "context"

type ctxKey string

const correlationIDKey ctxKey = "correlationID"

// WithCorrelationID attaches the correlation ID to the context.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// GetCorrelationID extracts the correlation ID from the context if present.
func GetCorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(correlationIDKey).(string); ok {
		return v
	}
	return ""
}
