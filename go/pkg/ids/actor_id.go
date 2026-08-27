package ids

import (
	"bytes"
	"errors"
	"time"
)

// ActorID 16 字节的 Actor ID 类型
type ActorID struct {
	data [ActorIDSize]byte
}

var nilActorID = ActorID{data: [ActorIDSize]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
}}

// NilActorID 返回空的 ActorID
func NilActorID() ActorID { return nilActorID }

// OfActorID 从 JobID、parent TaskID 和 counter 创建 ActorID
func OfActorID(jobID JobID, parentTaskID TaskID, counter uint64) ActorID {
	// 使用当前时间戳作为 extra 参数，避免相同参数组合产生相同 ID
	extra := time.Now().UnixNano()

	var id ActorID
	// 使用 generateUniqueBytesInto 直接写入目标，避免双重分配
	generateUniqueBytesInto(id.data[:ActorIDUniqueBytesSize], jobID, parentTaskID, counter, extra)
	copy(id.data[ActorIDUniqueBytesSize:], jobID.data[:])
	return id
}

// ActorIDNilFromJob 从 JobID 创建空的 ActorID
func ActorIDNilFromJob(jobID JobID) ActorID {
	var id ActorID
	copy(id.data[:ActorIDUniqueBytesSize], nilActorID.data[:ActorIDUniqueBytesSize])
	copy(id.data[ActorIDUniqueBytesSize:], jobID.data[:])
	return id
}

// ActorIDFromBinary 从字节数组创建 ActorID
func ActorIDFromBinary(data []byte) (ActorID, error) {
	if len(data) != ActorIDSize {
		return nilActorID, errors.New("invalid ActorID length")
	}
	var id ActorID
	copy(id.data[:], data)
	return id, nil
}

// ActorIDFromHex 从十六进制字符串创建 ActorID
// 优化：直接解码到结构体数组，避免双重分配
func ActorIDFromHex(hexStr string) (ActorID, error) {
	var id ActorID
	if err := decodeHexToBytes(id.data[:], hexStr); err != nil {
		return nilActorID, err
	}
	return id, nil
}

// IsNil 检查 ActorID 是否为空
func (id ActorID) IsNil() bool {
	return bytes.Equal(id.data[:], nilActorID.data[:])
}

// JobID 从 ActorID 提取 JobID
func (id ActorID) JobID() JobID {
	if id.IsNil() {
		return nilJobID
	}
	var jobID JobID
	copy(jobID.data[:], id.data[ActorIDUniqueBytesSize:])
	return jobID
}

// Binary 返回 ActorID 的字节数组表示
func (id ActorID) Binary() []byte { return id.data[:] }

// Hex 返回 ActorID 的十六进制字符串表示
func (id ActorID) Hex() string { return idToHex(id.data[:]) }

// String 返回 ActorID 的字符串表示
func (id ActorID) String() string {
	if id.IsNil() {
		return "NIL_ID"
	}
	return id.Hex()
}

// Hash 计算 ActorID 的哈希值
func (id ActorID) Hash() uint64 { return murmurHash64A(id.data[:], 0) }

// Size 返回 ActorID 的大小
func (id ActorID) Size() int { return ActorIDSize }

// Equal 比较两个 ActorID 是否相等
func (id ActorID) Equal(other ActorID) bool {
	return bytes.Equal(id.data[:], other.data[:])
}