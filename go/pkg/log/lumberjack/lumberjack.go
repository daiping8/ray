package lumberjack

import (
	"io"
	"os"

	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

func buildWriters(paths []string, opts *Options) []io.Writer {
	writers := make([]io.Writer, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		writers = append(writers, &lumberjack.Logger{
			Filename:   path,
			MaxSize:    opts.MaxSize,
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAge,
			Compress:   opts.Compress,
			LocalTime:  opts.LocalTime,
		})
	}
	return writers
}

func buildWriteSyncers(paths []string, opts *Options) []zapcore.WriteSyncer {
	syncers := make([]zapcore.WriteSyncer, 0, len(paths))
	for _, writer := range buildWriters(paths, opts) {
		syncers = append(syncers, zapcore.Lock(zapcore.AddSync(writer)))
	}
	return syncers
}

// NewWriter creates an io.Writer with rotation support.
// If paths is empty, returns the specified fallback Writer.
// If fallback is nil, defaults to os.Stdout.
func NewWriter(paths []string, opts *Options, fallback io.Writer) io.Writer {
	if len(paths) == 0 || paths[0] == "" {
		if fallback == nil {
			return os.Stdout
		}
		return fallback
	}

	if opts == nil {
		opts = DefaultOptions()
	}

	writers := buildWriters(paths, opts)
	if len(writers) == 0 {
		if fallback == nil {
			return os.Stdout
		}
		return fallback
	}
	return io.MultiWriter(writers...)
}

// NewWriteSyncer creates a zapcore.WriteSyncer.
// If paths is empty, returns the specified fallback WriteSyncer.
// If fallback is nil, defaults to os.Stdout wrapped as WriteSyncer.
func NewWriteSyncer(paths []string, opts *Options, fallback zapcore.WriteSyncer) zapcore.WriteSyncer {
	if len(paths) == 0 || paths[0] == "" {
		if fallback == nil {
			return zapcore.Lock(zapcore.AddSync(os.Stdout))
		}
		return fallback
	}

	if opts == nil {
		opts = DefaultOptions()
	}

	syncers := buildWriteSyncers(paths, opts)
	if len(syncers) == 0 {
		if fallback == nil {
			return zapcore.Lock(zapcore.AddSync(os.Stdout))
		}
		return fallback
	}
	return zapcore.NewMultiWriteSyncer(syncers...)
}
