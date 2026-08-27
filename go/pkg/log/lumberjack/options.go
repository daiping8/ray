// Package lumberjack provides log file rotation Writer wrapper.
// Based on gopkg.in/natefinch/lumberjack.v2.
package lumberjack

import "gopkg.in/natefinch/lumberjack.v2"

// Options holds log rotation configuration.
type Options struct {
	// MaxSize is the maximum size per file in MB.
	// Default is 100 MB.
	MaxSize int

	// MaxBackups is the maximum number of old files to retain.
	// Default is 3.
	MaxBackups int

	// MaxAge is the maximum number of days to retain logs.
	// Default is 7 days (0 means no age-based cleanup).
	MaxAge int

	// Compress determines whether to compress old files.
	// Default is true.
	Compress bool

	// LocalTime determines whether to use local time for file naming.
	// Default is true.
	LocalTime bool
}

// DefaultOptions returns default rotation configuration.
func DefaultOptions() *Options {
	return &Options{
		MaxSize:    100, // 100MB
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
		LocalTime:  true,
	}
}

// ApplyTo applies configuration to a lumberjack.Logger.
func (o *Options) ApplyTo(l *lumberjack.Logger) {
	if o == nil || l == nil {
		return
	}
	l.MaxSize = o.MaxSize
	l.MaxBackups = o.MaxBackups
	l.MaxAge = o.MaxAge
	l.Compress = o.Compress
	l.LocalTime = o.LocalTime
}