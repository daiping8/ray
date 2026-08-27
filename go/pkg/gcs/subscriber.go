// Package gcs provides the Go client for Ray Global Control Store (GCS).
package gcs

import (
	"context"
	"io"
)

// ErrorSubscriber 错误订阅者接口
// 对应 Python GcsErrorSubscriber
type ErrorSubscriber interface {
	io.Closer
	// Subscribe 订阅错误流
	Subscribe() error
	// Poll 轮询下一个错误
	Poll(ctx context.Context) (errorID []byte, errorData *ErrorData, err error)
}

// LogSubscriber 日志订阅者接口
// 对应 Python GcsLogSubscriber
type LogSubscriber interface {
	io.Closer
	// Subscribe 订阅日志流
	Subscribe() error
	// Poll 轮询下一条日志
	Poll(ctx context.Context) (*LogData, error)
}
