// Package log provides structured logging interface and global Logger.
// Design inspired by controller-runtime logging.
package log

import (
	"github.com/go-logr/logr"
)

// Log is the global Logger, initialized with DelegatingLogSink wrapping NullLogSink.
// After calling SetLogger, all logs will be routed to the actual implementation.
var Log logr.Logger

// SetLogger sets the underlying implementation of the global Logger.
// Passing nil/Discard Logger will reset to NullLogSink.
func SetLogger(l logr.Logger) {
	if l.GetSink() == nil {
		Log = logr.New(&NullLogSink{})
		return
	}
	if delegate, ok := Log.GetSink().(*DelegatingLogSink); ok {
		delegate.Fulfill(l.GetSink())
	} else {
		Log = l
	}
}

func init() {
	Log = logr.New(NewDelegatingLogSink())
}

// WithName returns a Logger with the specified name.
// This is a convenience method for Log.WithName(name).
func WithName(name string) logr.Logger {
	return Log.WithName(name)
}

// WithValues returns a Logger with preset key-value pairs.
// This is a convenience method for Log.WithValues(keysAndValues...).
func WithValues(keysAndValues ...interface{}) logr.Logger {
	return Log.WithValues(keysAndValues...)
}