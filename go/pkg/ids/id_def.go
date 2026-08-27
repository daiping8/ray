package ids

import (
	"bytes"
	"errors"
)

// =============================================================================
// 基础 Nil ID - 所有派生类型的空值基础
// =============================================================================

// nilIDBytes 是通用的空 ID 字节数组 (28 个 0xff)
var nilIDBytes = [UniqueIDSize]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
}

// =============================================================================
// WorkerID - UniqueID 派生类型，标识 Worker 进程
// =============================================================================

// WorkerID - UniqueID 派生类型，标识 Worker 进程
type WorkerID struct {
	UniqueID
}

var nilWorkerID = WorkerID{UniqueID: UniqueID{data: nilIDBytes}}

// NilWorkerID 返回空的 WorkerID
func NilWorkerID() WorkerID { return nilWorkerID }

// NewWorkerID 生成一个新的随机 WorkerID
func NewWorkerID() WorkerID {
	return WorkerID{UniqueID: NewUniqueID()}
}

// WorkerIDFromBinary 从字节数组创建 WorkerID
func WorkerIDFromBinary(data []byte) (WorkerID, error) {
	if len(data) != UniqueIDSize {
		return nilWorkerID, errors.New("invalid WorkerID length")
	}
	var id WorkerID
	copy(id.UniqueID.data[:], data)
	return id, nil
}

// WorkerIDFromHex 从十六进制字符串创建 WorkerID
func WorkerIDFromHex(hexStr string) (WorkerID, error) {
	uid, err := UniqueIDFromHex(hexStr)
	if err != nil {
		return nilWorkerID, err
	}
	return WorkerID{UniqueID: uid}, nil
}

// ComputeDriverIdFromJob 从 JobID 计算 Driver ID
func ComputeDriverIdFromJob(jobID JobID) WorkerID {
	var id WorkerID
	copy(id.UniqueID.data[:JobIDSize], jobID.data[:])
	copy(id.UniqueID.data[JobIDSize:], nilWorkerID.UniqueID.data[JobIDSize:])
	return id
}

// IsNil 检查 WorkerID 是否为空
func (id WorkerID) IsNil() bool {
	return bytes.Equal(id.UniqueID.data[:], nilWorkerID.UniqueID.data[:])
}

// Equal 比较两个 WorkerID 是否相等
func (id WorkerID) Equal(other WorkerID) bool {
	return bytes.Equal(id.UniqueID.data[:], other.UniqueID.data[:])
}

// =============================================================================
// NodeID - UniqueID 派生类型，标识物理节点
// =============================================================================

// NodeID - UniqueID 派生类型，标识物理节点
type NodeID struct {
	UniqueID
}

var nilNodeID = NodeID{UniqueID: UniqueID{data: nilIDBytes}}

// kGCSNodeID - 特殊常量，GCS NodeID 使用 28 个 0x00
var kGCSNodeID = NodeID{UniqueID: UniqueID{data: [UniqueIDSize]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
}}}

// NilNodeID 返回空的 NodeID
func NilNodeID() NodeID { return nilNodeID }

// GCSNodeID 返回 GCS NodeID
func GCSNodeID() NodeID { return kGCSNodeID }

// NewNodeID 生成一个新的随机 NodeID
func NewNodeID() NodeID {
	return NodeID{UniqueID: NewUniqueID()}
}

// NodeIDFromBinary 从字节数组创建 NodeID
func NodeIDFromBinary(data []byte) (NodeID, error) {
	if len(data) != UniqueIDSize {
		return nilNodeID, errors.New("invalid NodeID length")
	}
	var id NodeID
	copy(id.UniqueID.data[:], data)
	return id, nil
}

// NodeIDFromHex 从十六进制字符串创建 NodeID
func NodeIDFromHex(hexStr string) (NodeID, error) {
	uid, err := UniqueIDFromHex(hexStr)
	if err != nil {
		return nilNodeID, err
	}
	return NodeID{UniqueID: uid}, nil
}

// IsNil 检查 NodeID 是否为空
func (id NodeID) IsNil() bool {
	return bytes.Equal(id.UniqueID.data[:], nilNodeID.UniqueID.data[:])
}

// Equal 比较两个 NodeID 是否相等
func (id NodeID) Equal(other NodeID) bool {
	return bytes.Equal(id.UniqueID.data[:], other.UniqueID.data[:])
}

// =============================================================================
// ClusterID - UniqueID 派生类型，标识集群
// =============================================================================

// ClusterID - UniqueID 派生类型，标识集群
type ClusterID struct {
	UniqueID
}

var nilClusterID = ClusterID{UniqueID: UniqueID{data: nilIDBytes}}

// NilClusterID 返回空的 ClusterID
func NilClusterID() ClusterID { return nilClusterID }

// NewClusterID 生成一个新的随机 ClusterID
func NewClusterID() ClusterID {
	return ClusterID{UniqueID: NewUniqueID()}
}

// ClusterIDFromBinary 从字节数组创建 ClusterID
func ClusterIDFromBinary(data []byte) (ClusterID, error) {
	if len(data) != UniqueIDSize {
		return nilClusterID, errors.New("invalid ClusterID length")
	}
	var id ClusterID
	copy(id.UniqueID.data[:], data)
	return id, nil
}

// ClusterIDFromHex 从十六进制字符串创建 ClusterID
func ClusterIDFromHex(hexStr string) (ClusterID, error) {
	uid, err := UniqueIDFromHex(hexStr)
	if err != nil {
		return nilClusterID, err
	}
	return ClusterID{UniqueID: uid}, nil
}

// IsNil 检查 ClusterID 是否为空
func (id ClusterID) IsNil() bool {
	return bytes.Equal(id.UniqueID.data[:], nilClusterID.UniqueID.data[:])
}

// Equal 比较两个 ClusterID 是否相等
func (id ClusterID) Equal(other ClusterID) bool {
	return bytes.Equal(id.UniqueID.data[:], other.UniqueID.data[:])
}

// =============================================================================
// FunctionID - UniqueID 派生类型，标识函数
// =============================================================================

// FunctionID - UniqueID 派生类型，标识函数
type FunctionID struct {
	UniqueID
}

var nilFunctionID = FunctionID{UniqueID: UniqueID{data: nilIDBytes}}

// NilFunctionID 返回空的 FunctionID
func NilFunctionID() FunctionID { return nilFunctionID }

// NewFunctionID 生成一个新的随机 FunctionID
func NewFunctionID() FunctionID {
	return FunctionID{UniqueID: NewUniqueID()}
}

// FunctionIDFromBinary 从字节数组创建 FunctionID
func FunctionIDFromBinary(data []byte) (FunctionID, error) {
	if len(data) != UniqueIDSize {
		return nilFunctionID, errors.New("invalid FunctionID length")
	}
	var id FunctionID
	copy(id.UniqueID.data[:], data)
	return id, nil
}

// FunctionIDFromHex 从十六进制字符串创建 FunctionID
func FunctionIDFromHex(hexStr string) (FunctionID, error) {
	uid, err := UniqueIDFromHex(hexStr)
	if err != nil {
		return nilFunctionID, err
	}
	return FunctionID{UniqueID: uid}, nil
}

// IsNil 检查 FunctionID 是否为空
func (id FunctionID) IsNil() bool {
	return bytes.Equal(id.UniqueID.data[:], nilFunctionID.UniqueID.data[:])
}

// Equal 比较两个 FunctionID 是否相等
func (id FunctionID) Equal(other FunctionID) bool {
	return bytes.Equal(id.UniqueID.data[:], other.UniqueID.data[:])
}

// =============================================================================
// ActorClassID - UniqueID 派生类型，标识 Actor 类
// =============================================================================

// ActorClassID - UniqueID 派生类型，标识 Actor 类
type ActorClassID struct {
	UniqueID
}

var nilActorClassID = ActorClassID{UniqueID: UniqueID{data: nilIDBytes}}

// NilActorClassID 返回空的 ActorClassID
func NilActorClassID() ActorClassID { return nilActorClassID }

// NewActorClassID 生成一个新的随机 ActorClassID
func NewActorClassID() ActorClassID {
	return ActorClassID{UniqueID: NewUniqueID()}
}

// ActorClassIDFromBinary 从字节数组创建 ActorClassID
func ActorClassIDFromBinary(data []byte) (ActorClassID, error) {
	if len(data) != UniqueIDSize {
		return nilActorClassID, errors.New("invalid ActorClassID length")
	}
	var id ActorClassID
	copy(id.UniqueID.data[:], data)
	return id, nil
}

// ActorClassIDFromHex 从十六进制字符串创建 ActorClassID
func ActorClassIDFromHex(hexStr string) (ActorClassID, error) {
	uid, err := UniqueIDFromHex(hexStr)
	if err != nil {
		return nilActorClassID, err
	}
	return ActorClassID{UniqueID: uid}, nil
}

// IsNil 检查 ActorClassID 是否为空
func (id ActorClassID) IsNil() bool {
	return bytes.Equal(id.UniqueID.data[:], nilActorClassID.UniqueID.data[:])
}

// Equal 比较两个 ActorClassID 是否相等
func (id ActorClassID) Equal(other ActorClassID) bool {
	return bytes.Equal(id.UniqueID.data[:], other.UniqueID.data[:])
}

// =============================================================================
// ConfigID - UniqueID 派生类型，标识配置
// =============================================================================

// ConfigID - UniqueID 派生类型，标识配置
type ConfigID struct {
	UniqueID
}

var nilConfigID = ConfigID{UniqueID: UniqueID{data: nilIDBytes}}

// NilConfigID 返回空的 ConfigID
func NilConfigID() ConfigID { return nilConfigID }

// NewConfigID 生成一个新的随机 ConfigID
func NewConfigID() ConfigID {
	return ConfigID{UniqueID: NewUniqueID()}
}

// ConfigIDFromBinary 从字节数组创建 ConfigID
func ConfigIDFromBinary(data []byte) (ConfigID, error) {
	if len(data) != UniqueIDSize {
		return nilConfigID, errors.New("invalid ConfigID length")
	}
	var id ConfigID
	copy(id.UniqueID.data[:], data)
	return id, nil
}

// ConfigIDFromHex 从十六进制字符串创建 ConfigID
func ConfigIDFromHex(hexStr string) (ConfigID, error) {
	uid, err := UniqueIDFromHex(hexStr)
	if err != nil {
		return nilConfigID, err
	}
	return ConfigID{UniqueID: uid}, nil
}

// IsNil 检查 ConfigID 是否为空
func (id ConfigID) IsNil() bool {
	return bytes.Equal(id.UniqueID.data[:], nilConfigID.UniqueID.data[:])
}

// Equal 比较两个 ConfigID 是否相等
func (id ConfigID) Equal(other ConfigID) bool {
	return bytes.Equal(id.UniqueID.data[:], other.UniqueID.data[:])
}