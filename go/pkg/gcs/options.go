// Package gcs provides the Go client for Ray Global Control Store (GCS).
package gcs

import (
	"github.com/ray-project/ray/go/pkg/ids"
)

// ClientOptions GCS 客户端连接选项
type ClientOptions struct {
	// Address GCS 服务器地址，格式为 "host:port"
	Address string
	// ClusterID 集群 ID，为空时自动获取
	ClusterID ids.ClusterID
	// TimeoutMs 连接超时（毫秒）
	TimeoutMs int64
}

// GlobalStateOptions GlobalStateAccessor 连接选项
type GlobalStateOptions struct {
	// Address GCS 服务器地址，格式为 "host:port"
	Address string
	// ClusterID 集群 ID，为空时自动获取
	ClusterID ids.ClusterID
	// TimeoutMs 连接超时（毫秒）
	TimeoutMs int64
}
