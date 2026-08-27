package ids

import (
	"bytes"
	"errors"
)

// UniqueID 28 字节的唯一 ID 类型
type UniqueID struct {
	data [UniqueIDSize]byte
}

// nilUniqueID 使用 nilIDBytes 作为空值基础
var nilUniqueID = UniqueID{data: nilIDBytes}

// NilUniqueID 返回空的 UniqueID
func NilUniqueID() UniqueID { return nilUniqueID }

// NewUniqueID 生成一个新的随机 UniqueID
func NewUniqueID() UniqueID {
	var id UniqueID
	fillRandom(id.data[:])
	return id
}

// UniqueIDFromBinary 从字节数组创建 UniqueID
func UniqueIDFromBinary(data []byte) (UniqueID, error) {
	if len(data) != UniqueIDSize {
		return nilUniqueID, errors.New("invalid UniqueID length")
	}
	var id UniqueID
	copy(id.data[:], data)
	return id, nil
}

// UniqueIDFromHex 从十六进制字符串创建 UniqueID
// 优化：直接解码到结构体数组，避免双重分配
func UniqueIDFromHex(hexStr string) (UniqueID, error) {
	var id UniqueID
	if err := decodeHexToBytes(id.data[:], hexStr); err != nil {
		return nilUniqueID, err
	}
	return id, nil
}

// IsNil 检查 UniqueID 是否为空
func (id UniqueID) IsNil() bool {
	return bytes.Equal(id.data[:], nilUniqueID.data[:])
}

// Binary 返回 UniqueID 的字节数组表示
func (id UniqueID) Binary() []byte {
	return id.data[:]
}

// Hex 返回 UniqueID 的十六进制字符串表示
func (id UniqueID) Hex() string {
	return idToHex(id.data[:])
}

// String 返回 UniqueID 的字符串表示
func (id UniqueID) String() string {
	if id.IsNil() {
		return "NIL_ID"
	}
	return id.Hex()
}

// Hash 计算 UniqueID 的哈希值
func (id UniqueID) Hash() uint64 {
	return murmurHash64A(id.data[:], 0)
}

// Size 返回 UniqueID 的大小
func (id UniqueID) Size() int {
	return UniqueIDSize
}

// Equal 比较两个 UniqueID 是否相等
func (id UniqueID) Equal(other UniqueID) bool {
	return bytes.Equal(id.data[:], other.data[:])
}
