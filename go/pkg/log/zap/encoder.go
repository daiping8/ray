package zap

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// newEncoder creates an encoder based on options.
func newEncoder(opts *Options) zapcore.Encoder {
	if opts == nil {
		opts = DefaultOptions()
	}

	if opts.Development || opts.Encoder == ConsoleEncoder {
		return zapcore.NewConsoleEncoder(rayDevEncoderConfig())
	}
	return zapcore.NewJSONEncoder(productionEncoderConfig())
}

// rayDevEncoderConfig returns Ray development mode encoder config.
// Format matches Python: "2026-04-07 11:40:40,880\tINFO agent.py:170 -- message"
func rayDevEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // INFO/WARN/ERROR (no color)
		EncodeTime:     rayTimeEncoder,              // 2026-04-07 11:40:40,880
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // filename:line
		ConsoleSeparator: "\t",                     // tab separator after time
	}
}

// rayTimeEncoder implements Ray-compatible log time format.
// Format: "2026-04-07 11:40:40,880" (comma-separated milliseconds)
func rayTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05,000"))
}

// productionEncoderConfig returns production mode encoder config.
func productionEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochMillisTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}