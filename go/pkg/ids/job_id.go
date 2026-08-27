package ids

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// JobID 4 字节的 Job ID 类型
type JobID struct {
	data [JobIDSize]byte
}

var nilJobID = JobID{data: [4]byte{0xff, 0xff, 0xff, 0xff}}

// NilJobID 返回空的 JobID
func NilJobID() JobID { return nilJobID }

// NewJobID 生成一个新的随机 JobID
func NewJobID() JobID {
	var id JobID
	fillRandom(id.data[:])
	return id
}

// JobIDFromInt 从 uint32 整数创建 JobID
func JobIDFromInt(value uint32) JobID {
	var id JobID
	binary.LittleEndian.PutUint32(id.data[:], value)
	return id
}

// JobIDFromBinary 从字节数组创建 JobID
func JobIDFromBinary(data []byte) (JobID, error) {
	if len(data) != JobIDSize {
		return nilJobID, errors.New("invalid JobID length")
	}
	var id JobID
	copy(id.data[:], data)
	return id, nil
}

// JobIDFromHex 从十六进制字符串创建 JobID
// 优化：直接解码到结构体数组，避免双重分配
func JobIDFromHex(hexStr string) (JobID, error) {
	var id JobID
	if err := decodeHexToBytes(id.data[:], hexStr); err != nil {
		return nilJobID, err
	}
	return id, nil
}

// IsNil 检查 JobID 是否为空
func (id JobID) IsNil() bool {
	return bytes.Equal(id.data[:], nilJobID.data[:])
}

// ToInt 将 JobID 转换为 uint32 整数
func (id JobID) ToInt() uint32 {
	return binary.LittleEndian.Uint32(id.data[:])
}

// Binary 返回 JobID 的字节数组表示
func (id JobID) Binary() []byte {
	return id.data[:]
}

// Hex 返回 JobID 的十六进制字符串表示
func (id JobID) Hex() string {
	return idToHex(id.data[:])
}

// String 返回 JobID 的字符串表示
func (id JobID) String() string {
	if id.IsNil() {
		return "NIL_ID"
	}
	return id.Hex()
}

// Hash 计算 JobID 的哈希值
func (id JobID) Hash() uint64 {
	return murmurHash64A(id.data[:], 0)
}

// Size 返回 JobID 的大小
func (id JobID) Size() int {
	return JobIDSize
}

// Equal 比较两个 JobID 是否相等
func (id JobID) Equal(other JobID) bool {
	return bytes.Equal(id.data[:], other.data[:])
}
