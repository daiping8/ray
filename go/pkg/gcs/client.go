// Package gcs provides the Go client for Ray Global Control Store (GCS).
package gcs

import (
	"errors"
	"sync"

	"github.com/ray-project/ray/go/pkg/ids"
)

// ErrNotImplemented 表示功能尚未实现
var ErrNotImplemented = errors.New("not implemented")

// ErrKeyNotFound 表示键不存在（区别于键存在但值为空）
var ErrKeyNotFound = errors.New("key not found")

// 全局单例客户端
var (
	clientInstance Client
	clientMu       sync.RWMutex
	clientOnce     sync.Once
)

// Client GCS 客户端主接口
// 通过组合子接口提供所有 GCS 功能
type Client interface {
	InternalKVInterface
	NodeInfoInterface
	NodeResourceInterface
	ActorInfoInterface
	JobInfoInterface
	WorkerInfoInterface
	PlacementGroupInterface
	PublisherInterface
	AutoscalerInterface

	// Address 返回 GCS 服务器地址 (host:port)
	Address() string

	// ClusterID 返回集群 ID
	ClusterID() ids.ClusterID

	// Close 断开与 GCS 的连接
	Close() error

	// IsClosed reports whether the client has been closed.
	// This method is used to check if a cached client is still usable.
	IsClosed() bool

	ReportAutoscalingState(autoscalingState string) error
}

// SetClient 设置全局 GCS 客户端实例
// 由实现包（如 go/internal/gcs/native）在初始化时调用
// 使用 sync.Once 确保只设置一次，避免重复设置
// 此函数允许外部包注入客户端实例，避免循环依赖
func SetClient(client Client) {
	clientOnce.Do(func() {
		clientMu.Lock()
		defer clientMu.Unlock()
		clientInstance = client
	})
}

// GetClient 获取全局 GCS 客户端实例
// 如果尚未初始化，返回 nil 和 ErrNotImplemented 错误
func GetClient() (Client, error) {
	clientMu.RLock()
	defer clientMu.RUnlock()
	if clientInstance == nil {
		return nil, ErrNotImplemented
	}
	return clientInstance, nil
}

// ClearClient 清除全局 GCS 客户端实例
// 用于 Close 方法清理资源，防止后续获取已关闭的客户端
func ClearClient() {
	clientMu.Lock()
	defer clientMu.Unlock()
	clientInstance = nil
	// 重置 sync.Once 以允许重新设置
	clientOnce = sync.Once{}
}
