package ids

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// ObjectIDIndexType ObjectID 索引类型
type ObjectIDIndexType = uint32

// ObjectID 28 字节的 Object ID 类型
type ObjectID struct {
	data [ObjectIDSize]byte
}

var nilObjectID = ObjectID{data: [ObjectIDSize]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
}}

// NilObjectID 返回空的 ObjectID
func NilObjectID() ObjectID { return nilObjectID }

// ObjectIDFromIndex 从 TaskID 和 index 构造 ObjectID
func ObjectIDFromIndex(taskID TaskID, index ObjectIDIndexType) ObjectID {
	if index < 1 || index > uint32(MaxObjectIndex) {
		panic("invalid object index")
	}

	var id ObjectID
	copy(id.data[:TaskIDSize], taskID.data[:])
	binary.LittleEndian.PutUint32(id.data[TaskIDSize:], index)
	return id
}

// ObjectIDForActorHandle 从 ActorID 创建 Actor Handle 的 ObjectID
func ObjectIDForActorHandle(actorID ActorID) ObjectID {
	creationTaskID := TaskIDForActorCreationTask(actorID)
	return ObjectIDFromIndex(creationTaskID, 1)
}

// ObjectIDFromBinary 从字节数组创建 ObjectID
func ObjectIDFromBinary(data []byte) (ObjectID, error) {
	if len(data) != ObjectIDSize {
		return nilObjectID, errors.New("invalid ObjectID length")
	}
	var id ObjectID
	copy(id.data[:], data)
	return id, nil
}

// ObjectIDFromHex 从十六进制字符串创建 ObjectID
// 优化：直接解码到结构体数组，避免双重分配
func ObjectIDFromHex(hexStr string) (ObjectID, error) {
	var id ObjectID
	if err := decodeHexToBytes(id.data[:], hexStr); err != nil {
		return nilObjectID, err
	}
	return id, nil
}

// NewObjectID generates a random ObjectID for local mode testing.
// Consistent with Java's ObjectId.fromRandom() and C++ ObjectID::FromRandom.
func NewObjectID() ObjectID {
	var id ObjectID
	fillRandom(id.data[:])
	binary.LittleEndian.PutUint32(id.data[TaskIDSize:], 0)
	return id
}

// IsNil 检查 ObjectID 是否为空
func (id ObjectID) IsNil() bool {
	return bytes.Equal(id.data[:], nilObjectID.data[:])
}

// TaskID 从 ObjectID 提取 TaskID
func (id ObjectID) TaskID() TaskID {
	var taskID TaskID
	copy(taskID.data[:], id.data[:TaskIDSize])
	return taskID
}

// ObjectIndex 从 ObjectID 提取对象索引
func (id ObjectID) ObjectIndex() ObjectIDIndexType {
	return binary.LittleEndian.Uint32(id.data[TaskIDSize:])
}

// IsActorID 检查 ObjectID 是否为 Actor ID
func (id ObjectID) IsActorID() bool {
	taskID := id.TaskID()
	// 检查 TaskID 的前 8 字节（unique bytes）是否全为 0xff
	return bytes.Equal(taskID.data[:TaskIDUniqueBytesSize], nilTaskID.data[:TaskIDUniqueBytesSize])
}

// ToActorID 将 ObjectID 转换为 ActorID
func (id ObjectID) ToActorID() ActorID {
	// ActorID 位于 ObjectID.data[8:24]
	var actorID ActorID
	copy(actorID.data[:], id.data[8:24])
	return actorID
}

// Binary 返回 ObjectID 的字节数组表示
func (id ObjectID) Binary() []byte { return id.data[:] }

// Hex 返回 ObjectID 的十六进制字符串表示
func (id ObjectID) Hex() string { return idToHex(id.data[:]) }

// String 返回 ObjectID 的字符串表示
func (id ObjectID) String() string {
	if id.IsNil() {
		return "NIL_ID"
	}
	return id.Hex()
}

// Hash 计算 ObjectID 的哈希值
func (id ObjectID) Hash() uint64 { return murmurHash64A(id.data[:], 0) }

// Size 返回 ObjectID 的大小
func (id ObjectID) Size() int { return ObjectIDSize }

// Equal 比较两个 ObjectID 是否相等
func (id ObjectID) Equal(other ObjectID) bool {
	return bytes.Equal(id.data[:], other.data[:])
}
