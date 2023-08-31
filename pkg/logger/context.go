package logger

import (
	"context"
	"log/slog"
)

var loggerKey = ctxKey{}

// ctxKey is used to store the logger in the ctx.
// Using a new type avoids collisions.
type ctxKey struct{}

// WithContext sets the logger as the logger for the context
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns the logger from the context
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}

	return new()
}
