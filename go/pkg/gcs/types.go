// Package gcs provides the Go client for Ray Global Control Store (GCS).
package gcs

import (
	"fmt"
	"time"

	"github.com/ray-project/ray/go/pkg/ids"
)

// ErrorData 错误数据结构
// 对应 Python GcsErrorSubscriber 中的错误数据格式
type ErrorData struct {
	// JobID 关联的作业 ID
	JobID ids.JobID
	// Type 错误类型
	Type string
	// ErrorMessage 详细错误描述
	ErrorMessage string
	// Timestamp Unix 时间戳（毫秒）
	Timestamp int64
}

// String 实现 fmt.Stringer 接口
func (e ErrorData) String() string {
	return fmt.Sprintf("ErrorData{JobID: %s, Type: %s, Message: %s, Time: %s}",
		e.JobID.Hex(), e.Type, e.ErrorMessage, time.UnixMilli(e.Timestamp))
}

// LogData 日志数据结构
// 对应 Python GcsLogSubscriber 中的日志数据格式
type LogData struct {
	// IP 来源 IP 地址
	IP string
	// PID 进程 ID
	PID uint32
	// JobID 关联的作业 ID
	JobID ids.JobID
	// IsError 是否为错误日志
	IsError bool
	// ActorName Actor 名称（如适用）
	ActorName string
	// TaskName 任务名称（如适用）
	TaskName string
	// Lines 日志文本行
	Lines []string
}

// LogBatchPayload is the publish-side payload for a log batch.
// It mirrors the Python GcsClient.publish_logs input schema.
type LogBatchPayload struct {
	IP        string
	PID       string
	JobID     string
	IsError   bool
	Lines     []string
	ActorName string
	TaskName  string
}

// String 实现 fmt.Stringer 接口
func (l LogData) String() string {
	return fmt.Sprintf("LogData{IP: %s, PID: %d, JobID: %s, IsError: %v, Lines: %d}",
		l.IP, l.PID, l.JobID.Hex(), l.IsError, len(l.Lines))
}
