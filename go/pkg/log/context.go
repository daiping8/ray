package log

import (
	"context"

	"github.com/go-logr/logr"
)

// loggerKey is the key type for context storage.
type loggerKey struct{}

// FromContext retrieves a Logger from context.
// If no Logger is found in context, returns the global Log.
func FromContext(ctx context.Context) logr.Logger {
	if l, ok := ctx.Value(loggerKey{}).(logr.Logger); ok {
		return l
	}
	return Log
}

// IntoContext stores a Logger into context.
func IntoContext(ctx context.Context, l logr.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromContextOrDiscard retrieves a Logger from context.
// If no Logger is found, returns a Discard Logger (no output).
func FromContextOrDiscard(ctx context.Context) logr.Logger {
	if l, ok := ctx.Value(loggerKey{}).(logr.Logger); ok {
		return l
	}
	return logr.Discard()
}