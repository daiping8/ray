// Package gcs provides the Go client for Ray Global Control Store (GCS).
package gcs

import (
	"sync"

	"github.com/ray-project/ray/go/pkg/ids"
	"github.com/ray-project/ray/go/proto"
)

// 全局单例 GlobalStateAccessor
var (
	stateAccessorInstance GlobalStateAccessor
	stateAccessorOnce     sync.Once
)

// GlobalStateAccessor GCS 全局状态访问器（同步接口）
// 对应 C++ ray::gcs::GlobalStateAccessor
type GlobalStateAccessor interface {
	// Close 关闭访问器连接
	Close() error
	// Connect 建立与 GCS 的连接
	Connect() (bool, error)

	// --- Job ---
	// GetAllJobInfo 获取所有作业信息
	// skipSubmissionJobInfoField: 跳过 JobSubmissionInfo 字段
	// skipIsRunningTasksField: 跳过 IsRunningTasks 字段
	GetAllJobInfo(skipSubmissionJobInfoField, skipIsRunningTasksField bool) ([]*proto.JobTableData, error)
	// GetNextJobID 获取下一个作业 ID
	GetNextJobID() (ids.JobID, error)

	// --- Node ---
	// GetAllNodeInfo 获取所有节点信息
	GetAllNodeInfo() (map[ids.NodeID]*proto.GcsNodeInfo, error)
	// GetNode 获取单个节点信息
	GetNode(nodeID ids.NodeID) (*proto.GcsNodeInfo, error)
	// GetDrainingNodes 获取排水节点及其截止时间
	GetDrainingNodes() (map[ids.NodeID]int64, error)
	// GetNodeToConnectForDriver 获取驱动连接的节点
	GetNodeToConnectForDriver(nodeIPAddress string) (*proto.GcsNodeInfo, error)

	// --- Internal KV ---
	// GetInternalKV 获取内部 KV 值
	GetInternalKV(ns string, key string) ([]byte, error)

	// --- Resource ---
	// GetAllAvailableResources 获取所有可用资源
	GetAllAvailableResources() ([]*proto.AvailableResources, error)
	// GetAllTotalResources 获取所有总资源
	GetAllTotalResources() ([]*proto.TotalResources, error)
	// GetAllResourceUsage 获取所有资源使用情况
	GetAllResourceUsage() (*proto.ResourceUsageBatchData, error)

	// --- Actor ---
	// GetAllActorInfo 获取所有 Actor 信息
	// jobID: 按作业 ID 过滤，nil 表示不过滤
	// actorStateName: 按状态过滤，nil 表示不过滤
	GetAllActorInfo(jobID *ids.JobID, actorStateName *string) ([]*proto.ActorTableData, error)
	// GetActorInfo 获取单个 Actor 信息
	GetActorInfo(actorID ids.ActorID) (*proto.ActorTableData, error)

	// --- Worker ---
	// GetAllWorkerInfo 获取所有 Worker 信息
	GetAllWorkerInfo() ([]*proto.WorkerTableData, error)
	// GetWorkerInfo 获取单个 Worker 信息
	GetWorkerInfo(workerID ids.WorkerID) (*proto.WorkerTableData, error)
	// AddWorkerInfo 添加 Worker 信息
	AddWorkerInfo(data *proto.WorkerTableData) (bool, error)
	// GetWorkerDebuggerPort 获取 Worker 调试器端口
	GetWorkerDebuggerPort(workerID ids.WorkerID) (uint32, error)
	// UpdateWorkerDebuggerPort 更新 Worker 调试器端口
	UpdateWorkerDebuggerPort(workerID ids.WorkerID, debuggerPort uint32) (bool, error)
	// UpdateWorkerNumPausedThreads 更新 Worker 暂停线程数
	UpdateWorkerNumPausedThreads(workerID ids.WorkerID, numPausedThreadsDelta int32) (bool, error)

	// --- Task ---
	// GetAllTaskEvents 获取所有任务事件
	GetAllTaskEvents() ([]*proto.TaskEvents, error)

	// --- Placement Group ---
	// GetAllPlacementGroupInfo 获取所有放置组信息
	GetAllPlacementGroupInfo() ([]*proto.PlacementGroupTableData, error)
	// GetPlacementGroupInfo 按 ID 获取放置组信息
	GetPlacementGroupInfo(pgID ids.PlacementGroupID) (*proto.PlacementGroupTableData, error)
	// GetPlacementGroupByName 按名称获取放置组信息
	GetPlacementGroupByName(name, rayNamespace string) (*proto.PlacementGroupTableData, error)

	// --- System ---
	// GetSystemConfig 获取系统配置
	GetSystemConfig() (string, error)
}

// SetGlobalStateAccessor 设置全局 GlobalStateAccessor 实例
// 由实现包（如 go/internal/gcs/native）在初始化时调用
// 使用 sync.Once 确保只设置一次，避免重复设置
func SetGlobalStateAccessor(accessor GlobalStateAccessor) {
	stateAccessorOnce.Do(func() {
		stateAccessorInstance = accessor
	})
}

// GetGlobalStateAccessor 获取全局 GlobalStateAccessor 实例
// 如果尚未初始化，返回 nil 和 ErrNotImplemented 错误
func GetGlobalStateAccessor() (GlobalStateAccessor, error) {
	if stateAccessorInstance == nil {
		return nil, ErrNotImplemented
	}
	return stateAccessorInstance, nil
}
