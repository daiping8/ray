package ids

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// LeaseID 32 字节的 Lease ID 类型
type LeaseID struct {
	data [LeaseIDSize]byte
}

var nilLeaseID = LeaseID{data: [LeaseIDSize]byte{
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
}}

// NilLeaseID 返回空的 LeaseID
func NilLeaseID() LeaseID { return nilLeaseID }

// LeaseIDFromWorker 从 WorkerID 和 counter 创建 LeaseID
func LeaseIDFromWorker(workerID WorkerID, counter uint32) LeaseID {
	var id LeaseID
	binary.LittleEndian.PutUint32(id.data[:LeaseIDUniqueBytesSize], counter)
	copy(id.data[LeaseIDUniqueBytesSize:], workerID.data[:])
	return id
}

// NewLeaseID 生成一个新的随机 LeaseID
func NewLeaseID() LeaseID {
	var id LeaseID
	fillRandom(id.data[:])
	return id
}

// LeaseIDFromBinary 从字节数组创建 LeaseID
func LeaseIDFromBinary(data []byte) (LeaseID, error) {
	if len(data) != LeaseIDSize {
		return nilLeaseID, errors.New("invalid LeaseID length")
	}
	var id LeaseID
	copy(id.data[:], data)
	return id, nil
}

// LeaseIDFromHex 从十六进制字符串创建 LeaseID
// 优化：直接解码到结构体数组，避免双重分配
func LeaseIDFromHex(hexStr string) (LeaseID, error) {
	var id LeaseID
	if err := decodeHexToBytes(id.data[:], hexStr); err != nil {
		return nilLeaseID, err
	}
	return id, nil
}

// IsNil 检查 LeaseID 是否为空
func (id LeaseID) IsNil() bool {
	return bytes.Equal(id.data[:], nilLeaseID.data[:])
}

// WorkerID 从 LeaseID 提取 WorkerID
func (id LeaseID) WorkerID() WorkerID {
	var workerID WorkerID
	copy(workerID.data[:], id.data[LeaseIDUniqueBytesSize:])
	return workerID
}

// Binary 返回 LeaseID 的字节数组表示
func (id LeaseID) Binary() []byte { return id.data[:] }

// Hex 返回 LeaseID 的十六进制字符串表示
func (id LeaseID) Hex() string { return idToHex(id.data[:]) }

// String 返回 LeaseID 的字符串表示
func (id LeaseID) String() string {
	if id.IsNil() {
		return "NIL_ID"
	}
	return id.Hex()
}

// Hash 计算 LeaseID 的哈希值
func (id LeaseID) Hash() uint64 { return murmurHash64A(id.data[:], 0) }

// Size 返回 LeaseID 的大小
func (id LeaseID) Size() int { return LeaseIDSize }

// Equal 比较两个 LeaseID 是否相等
func (id LeaseID) Equal(other LeaseID) bool {
	return bytes.Equal(id.data[:], other.data[:])
}
