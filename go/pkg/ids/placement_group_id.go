package ids

import (
	"bytes"
	"errors"
)

// PlacementGroupID 18 字节的 Placement Group ID 类型
type PlacementGroupID struct {
	data [PlacementGroupIDSize]byte
}

var nilPlacementGroupID = PlacementGroupID{data: [PlacementGroupIDSize]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
}}

// NilPlacementGroupID 返回空的 PlacementGroupID
func NilPlacementGroupID() PlacementGroupID { return nilPlacementGroupID }

// OfPlacementGroupID 从 JobID 创建 PlacementGroupID
func OfPlacementGroupID(jobID JobID) PlacementGroupID {
	var id PlacementGroupID
	fillRandom(id.data[:PlacementGroupIDUniqueBytesSize])
	copy(id.data[PlacementGroupIDUniqueBytesSize:], jobID.data[:])
	return id
}

// PlacementGroupIDFromBinary 从字节数组创建 PlacementGroupID
func PlacementGroupIDFromBinary(data []byte) (PlacementGroupID, error) {
	if len(data) != PlacementGroupIDSize {
		return nilPlacementGroupID, errors.New("invalid PlacementGroupID length")
	}
	var id PlacementGroupID
	copy(id.data[:], data)
	return id, nil
}

// PlacementGroupIDFromHex 从十六进制字符串创建 PlacementGroupID
// 优化：直接解码到结构体数组，避免双重分配
func PlacementGroupIDFromHex(hexStr string) (PlacementGroupID, error) {
	var id PlacementGroupID
	if err := decodeHexToBytes(id.data[:], hexStr); err != nil {
		return nilPlacementGroupID, err
	}
	return id, nil
}

// IsNil 检查 PlacementGroupID 是否为空
func (id PlacementGroupID) IsNil() bool {
	return bytes.Equal(id.data[:], nilPlacementGroupID.data[:])
}

// JobID 从 PlacementGroupID 提取 JobID
func (id PlacementGroupID) JobID() JobID {
	if id.IsNil() {
		return nilJobID
	}
	var jobID JobID
	copy(jobID.data[:], id.data[PlacementGroupIDUniqueBytesSize:])
	return jobID
}

// Binary 返回 PlacementGroupID 的字节数组表示
func (id PlacementGroupID) Binary() []byte { return id.data[:] }

// Hex 返回 PlacementGroupID 的十六进制字符串表示
func (id PlacementGroupID) Hex() string { return idToHex(id.data[:]) }

// String 返回 PlacementGroupID 的字符串表示
func (id PlacementGroupID) String() string {
	if id.IsNil() {
		return "NIL_ID"
	}
	return id.Hex()
}

// Hash 计算 PlacementGroupID 的哈希值
func (id PlacementGroupID) Hash() uint64 { return murmurHash64A(id.data[:], 0) }

// Size 返回 PlacementGroupID 的大小
func (id PlacementGroupID) Size() int { return PlacementGroupIDSize }

// Equal 比较两个 PlacementGroupID 是否相等
func (id PlacementGroupID) Equal(other PlacementGroupID) bool {
	return bytes.Equal(id.data[:], other.data[:])
}