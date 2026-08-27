// Package gcs provides the Go client for Ray Global Control Store (GCS).
package gcs

import (
	"context"

	"github.com/ray-project/ray/go/pkg/ids"
	"github.com/ray-project/ray/go/proto"
)

// InternalKVInterface 内部 KV 存储接口
type InternalKVInterface interface {
	// Get 获取 KV 值
	Get(ctx context.Context, ns, key string) ([]byte, error)
	// MultiGet 批量获取 KV 值
	MultiGet(ctx context.Context, ns string, keys []string) (map[string][]byte, error)
	// Put 放置 KV 值
	Put(ctx context.Context, ns, key string, value []byte, overwrite bool) (bool, error)
	// Del 删除 KV 值
	Del(ctx context.Context, ns, key string, delByPrefix bool) (int, error)
	// Keys 获取键列表
	Keys(ctx context.Context, ns, prefix string) ([]string, error)
	// Exists 检查键是否存在
	Exists(ctx context.Context, ns, key string) (bool, error)
}

// NodeInfoInterface 节点信息接口
type NodeInfoInterface interface {
	// CheckAlive 检查节点是否存活
	CheckAlive(ctx context.Context, nodeIDs []ids.NodeID) ([]bool, error)
	// GetAll 获取节点信息
	GetAll(ctx context.Context, nodeIDs []ids.NodeID) (map[ids.NodeID]*proto.GcsNodeInfo, error)
	// DrainNodes 排水节点
	DrainNodes(ctx context.Context, nodeIDs []ids.NodeID) ([]ids.NodeID, error)
	// GetNodeToConnect 获取驱动节点连接信息
	GetNodeToConnect(ctx context.Context, nodeIpAddress string) (*proto.GcsNodeInfo, error)
}

// NodeResourceInterface 节点资源接口
type NodeResourceInterface interface {
	// GetAvailableResources 获取可用资源
	GetAvailableResources(ctx context.Context, nodeID ids.NodeID) (*proto.AvailableResources, error)
	// GetTotalResources 获取总资源
	GetTotalResources(ctx context.Context, nodeID ids.NodeID) (*proto.TotalResources, error)
}

// ActorInfoInterface Actor 信息接口
type ActorInfoInterface interface {
	// GetActorInfo 获取 Actor 信息
	GetActorInfo(ctx context.Context, actorID ids.ActorID) (*proto.ActorTableData, error)
	// ListActors 列出所有 Actor
	ListActors(ctx context.Context, jobID *ids.JobID) ([]*proto.ActorTableData, error)
}

// JobInfoInterface 作业信息接口
type JobInfoInterface interface {
	// GetJobInfo 获取作业信息
	GetJobInfo(ctx context.Context, jobID ids.JobID) (*proto.JobTableData, error)
	// ListJobs 列出所有作业
	ListJobs(ctx context.Context) ([]*proto.JobTableData, error)
	// NextJobID 获取下一个作业 ID
	NextJobID(ctx context.Context) (ids.JobID, error)
}

// WorkerInfoInterface Worker 信息接口
type WorkerInfoInterface interface {
	// GetWorkerInfo 获取 Worker 信息
	GetWorkerInfo(ctx context.Context, workerID ids.WorkerID) (*proto.WorkerTableData, error)
	// ListWorkers 列出所有 Worker
	ListWorkers(ctx context.Context) ([]*proto.WorkerTableData, error)
}

// PlacementGroupInterface 放置组接口
type PlacementGroupInterface interface {
	// GetPlacementGroup 获取放置组信息
	GetPlacementGroup(ctx context.Context, pgID ids.PlacementGroupID) (*proto.PlacementGroupTableData, error)
	// ListPlacementGroups 列出所有放置组
	ListPlacementGroups(ctx context.Context) ([]*proto.PlacementGroupTableData, error)
}

// PublisherInterface 发布器接口
type PublisherInterface interface {
	// PublishErrors 发布错误流
	PublishErrors(ctx context.Context) (<-chan ErrorData, error)
	// PublishLogs 发布日志流
	PublishLogs(ctx context.Context) (<-chan LogData, error)
}

// AutoscalerInterface 自动扩缩器接口
type AutoscalerInterface interface {
	// GetAutoscalerStatus 获取自动扩缩器状态（返回反序列化后的 protobuf 对象）
	GetAutoscalerStatus(ctx context.Context) (*proto.GetClusterStatusReply, error)
}
