package common

import (
	"fmt"
	"github.com/ray-project/ray/go/pkg/log/lumberjack"
	"github.com/ray-project/ray/go/pkg/log/zap"
	"go.uber.org/zap/zapcore"
	"path/filepath"
)

type LogOption struct {
	// LoggingLevel 日志级别 (可选，默认："info")
	LoggingLevel string

	// LoggingFormat 日志格式 (可选)
	LoggingFormat string

	// LoggingFilename 日志文件名 (可选，默认："monitor.log")
	LoggingFilename string

	// LogsDir 日志目录 (必需)
	LogsDir string

	// LoggingRotateBytes 日志轮转大小 (可选，默认：100MB)
	LoggingRotateBytes int

	// LoggingRotateBackupCount 日志备份数量 (可选，默认：5)
	LoggingRotateBackupCount int

	// StdoutFilepath stdout 输出文件路径 (可选)
	StdoutFilepath string

	// StderrFilepath stderr 输出文件路径 (可选)
	StderrFilepath string
}

// Option 定义了修改 logConfig 的函数类型
type Option func(*LogOption)

// WithLoggingLevel 设置日志级别
func WithLoggingLevel(level string) Option {
	return func(c *LogOption) {
		c.LoggingLevel = level
	}
}

// WithLoggingFormat 设置日志格式
func WithLoggingFormat(format string) Option {
	return func(c *LogOption) {
		c.LoggingFormat = format
	}
}

func WithLogsDir(logsDir string) Option {
	return func(c *LogOption) {
		if logsDir != "" {
			c.LogsDir = logsDir
		}
	}
}

// WithLoggingFilename 设置日志文件名
func WithLoggingFilename(filename string) Option {
	return func(c *LogOption) {
		c.LoggingFilename = filename
	}
}

// WithLoggingRotateBytes 设置日志轮转大小 (单位: Bytes)
func WithLoggingRotateBytes(bytes int) Option {
	return func(c *LogOption) {
		if bytes > 0 {
			c.LoggingRotateBytes = bytes
		}
	}
}

// WithLoggingRotateBackupCount 设置日志备份数量
func WithLoggingRotateBackupCount(count int) Option {
	return func(c *LogOption) {
		if count > 0 {
			c.LoggingRotateBackupCount = count
		}
	}
}

// WithStdoutFilepath 设置 stdout 输出文件路径
func WithStdoutFilepath(path string) Option {
	return func(c *LogOption) {
		c.StdoutFilepath = path
	}
}

// WithStderrFilepath 设置 stderr 输出文件路径
func WithStderrFilepath(path string) Option {
	return func(c *LogOption) {
		c.StderrFilepath = path
	}
}

// NewLogOption 创建并返回一个新的日志配置实例
func NewLogOption(opts ...Option) *LogOption {
	// 默认值
	cfg := &LogOption{
		LoggingLevel:             "info",
		LoggingRotateBytes:       LOGGING_ROTATE_BYTES,
		LoggingRotateBackupCount: LOGGING_ROTATE_BACKUP_COUNT,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// ParseLogLevel 解析日志级别
func ParseLogLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "critical":
		return zapcore.PanicLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unsupported log level: %s", level)
	}
}

// SetupComponentLogger 配置组件日志系统
func (opt *LogOption) SetupComponentLogger() error {
	// 解析日志级别
	loggingZapcoreLevel, err := ParseLogLevel(opt.LoggingLevel)
	if err != nil {
		return err
	}

	// 构建完整日志文件路径
	var logOutputPaths []string
	var logErrorOutputPaths []string

	if opt.LogsDir != "" && opt.LoggingFilename != "" {
		logFilePath := filepath.Join(opt.LogsDir, opt.LoggingFilename)
		logOutputPaths = []string{logFilePath}
		logErrorOutputPaths = []string{logFilePath}
	}

	// 如果配置了 stdout/stderr 文件路径，添加到输出路径
	if opt.StdoutFilepath != "" {
		if opt.LogsDir != "" && !filepath.IsAbs(opt.StdoutFilepath) {
			logOutputPaths = append(logOutputPaths, filepath.Join(opt.LogsDir, opt.StdoutFilepath))
		} else {
			logOutputPaths = append(logOutputPaths, opt.StdoutFilepath)
		}
	}

	if opt.StderrFilepath != "" {
		if opt.LogsDir != "" && !filepath.IsAbs(opt.StderrFilepath) {
			logErrorOutputPaths = append(logErrorOutputPaths, filepath.Join(opt.LogsDir, opt.StderrFilepath))
		} else {
			logErrorOutputPaths = append(logErrorOutputPaths, opt.StderrFilepath)
		}
	}

	// 如果没有配置日志文件路径，使用默认输出
	if len(logOutputPaths) == 0 {
		logOutputPaths = []string{}
	}
	if len(logErrorOutputPaths) == 0 {
		logErrorOutputPaths = []string{}
	}

	// 将字节转换为 MB（lumberjack 使用 MB 为单位）
	maxSizeMB := opt.LoggingRotateBytes / (1024 * 1024)
	if maxSizeMB <= 0 {
		maxSizeMB = 1
	}

	// 根据日志格式确定编码器类型
	var encoderType zap.EncoderType
	if opt.LoggingFormat != "" {
		// 如果配置了自定义格式，使用 Console 编码器
		encoderType = zap.ConsoleEncoder
	} else {
		// 默认使用 JSON 编码器（生产模式）
		encoderType = zap.JSONEncoder
	}

	if err := zap.SetupDefaultLogger(
		zap.WithLevel(loggingZapcoreLevel),
		zap.WithEncoder(encoderType),
		zap.WithDevelopment(DefaultLoggingDevelopment),
		zap.WithOutputPaths(logOutputPaths...),
		zap.WithErrorOutputPaths(logErrorOutputPaths...),
		zap.WithRotation(&lumberjack.Options{
			MaxSize:    maxSizeMB,
			MaxBackups: opt.LoggingRotateBackupCount,
			Compress:   true,
			LocalTime:  true,
		}),
	); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	return nil
}
