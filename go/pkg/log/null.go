package log

import "github.com/go-logr/logr"

// NullLogSink is a no-op LogSink implementation that discards all logs.
// Used for initialization and as the default empty log implementation.
type NullLogSink struct{}

// Init does nothing.
func (n *NullLogSink) Init(info logr.RuntimeInfo) {}

// Enabled always returns false, disabling all log levels.
func (n *NullLogSink) Enabled(level int) bool { return false }

// Info outputs nothing.
func (n *NullLogSink) Info(level int, msg string, keysAndValues ...interface{}) {}

// Error outputs nothing.
func (n *NullLogSink) Error(err error, msg string, keysAndValues ...interface{}) {}

// WithName returns itself, behavior unchanged.
func (n *NullLogSink) WithName(name string) logr.LogSink { return n }

// WithValues returns itself, behavior unchanged.
func (n *NullLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink { return n }