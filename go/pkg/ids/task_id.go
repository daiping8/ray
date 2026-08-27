package ids

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// TaskIDAttemptNumberMask 用于清除 TaskID unique bytes 的最低字节
// 以便设置 attempt number
const TaskIDAttemptNumberMask = 0xFFFFFFFFFFFFFF00

// TaskID 24 字节的 Task ID 类型
type TaskID struct {
	data [TaskIDSize]byte
}

var nilTaskID = TaskID{data: [TaskIDSize]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
}}

// NilTaskID 返回空的 TaskID
func NilTaskID() TaskID { return nilTaskID }

// TaskIDForDriverTask 从 JobID 创建 Driver Task 的 TaskID
func TaskIDForDriverTask(jobID JobID) TaskID {
	var id TaskID
	// unique bytes 全部为 0xff
	copy(id.data[:TaskIDUniqueBytesSize], nilTaskID.data[:TaskIDUniqueBytesSize])
	// 使用 dummy actor ID
	dummyActorID := ActorIDNilFromJob(jobID)
	copy(id.data[TaskIDUniqueBytesSize:], dummyActorID.data[:])
	return id
}

// TaskIDForActorCreationTask 从 ActorID 创建 Actor Creation Task 的 TaskID
func TaskIDForActorCreationTask(actorID ActorID) TaskID {
	var id TaskID
	// unique bytes 全部为 0xff
	copy(id.data[:TaskIDUniqueBytesSize], nilTaskID.data[:TaskIDUniqueBytesSize])
	copy(id.data[TaskIDUniqueBytesSize:], actorID.data[:])
	return id
}

// TaskIDForActorTask 创建 Actor Task 的 TaskID
func TaskIDForActorTask(jobID JobID, parentTaskID TaskID,
	counter uint64, actorID ActorID,
) TaskID {
	var id TaskID
	// 使用 generateUniqueBytesInto 直接写入目标，避免双重分配
	generateUniqueBytesInto(id.data[:TaskIDUniqueBytesSize], jobID, parentTaskID, counter, 0)
	copy(id.data[TaskIDUniqueBytesSize:], actorID.data[:])
	return id
}

// TaskIDForNormalTask 创建普通 Task 的 TaskID
func TaskIDForNormalTask(jobID JobID, parentTaskID TaskID, counter uint64) TaskID {
	var id TaskID
	// 使用 generateUniqueBytesInto 直接写入目标，避免双重分配
	generateUniqueBytesInto(id.data[:TaskIDUniqueBytesSize], jobID, parentTaskID, counter, 0)
	// 使用 dummy actor ID
	dummyActorID := ActorIDNilFromJob(jobID)
	copy(id.data[TaskIDUniqueBytesSize:], dummyActorID.data[:])
	return id
}

// TaskIDForExecutionAttempt 创建执行尝试（重试）的 TaskID
func TaskIDForExecutionAttempt(taskID TaskID, attemptNumber uint64) TaskID {
	var newID TaskID
	copy(newID.data[:], taskID.data[:])

	// 修改 unique bytes 部分（前 8 字节）
	uniqueBytes := binary.LittleEndian.Uint64(newID.data[:8])
	uniqueBytes &= TaskIDAttemptNumberMask // 清除最低字节
	uniqueBytes += attemptNumber
	binary.LittleEndian.PutUint64(newID.data[:8], uniqueBytes)

	return newID
}

// TaskIDFromBinary 从字节数组创建 TaskID
func TaskIDFromBinary(data []byte) (TaskID, error) {
	if len(data) != TaskIDSize {
		return nilTaskID, errors.New("invalid TaskID length")
	}
	var id TaskID
	copy(id.data[:], data)
	return id, nil
}

// TaskIDFromHex 从十六进制字符串创建 TaskID
// 优化：直接解码到结构体数组，避免双重分配
func TaskIDFromHex(hexStr string) (TaskID, error) {
	var id TaskID
	if err := decodeHexToBytes(id.data[:], hexStr); err != nil {
		return nilTaskID, err
	}
	return id, nil
}

// IsNil 检查 TaskID 是否为空
func (id TaskID) IsNil() bool {
	return bytes.Equal(id.data[:], nilTaskID.data[:])
}

// ActorID 从 TaskID 提取 ActorID
func (id TaskID) ActorID() ActorID {
	var actorID ActorID
	copy(actorID.data[:], id.data[TaskIDUniqueBytesSize:])
	return actorID
}

// JobID 从 TaskID 提取 JobID
func (id TaskID) JobID() JobID {
	var jobID JobID
	copy(jobID.data[:], id.data[TaskIDUniqueBytesSize+ActorIDUniqueBytesSize:])
	return jobID
}

// IsForActorCreationTask 检查是否为 Actor Creation Task
func (id TaskID) IsForActorCreationTask() bool {
	// unique bytes 全为 0xff 且 actor ID 非 nil
	return bytes.Equal(id.data[:TaskIDUniqueBytesSize], nilTaskID.data[:TaskIDUniqueBytesSize]) && !id.ActorID().IsNil()
}

// Binary 返回 TaskID 的字节数组表示
func (id TaskID) Binary() []byte { return id.data[:] }

// Hex 返回 TaskID 的十六进制字符串表示
func (id TaskID) Hex() string { return idToHex(id.data[:]) }

// String 返回 TaskID 的字符串表示
func (id TaskID) String() string {
	if id.IsNil() {
		return "NIL_ID"
	}
	return id.Hex()
}

// Hash 计算 TaskID 的哈希值
func (id TaskID) Hash() uint64 { return murmurHash64A(id.data[:], 0) }

// Size 返回 TaskID 的大小
func (id TaskID) Size() int { return TaskIDSize }

// Equal 比较两个 TaskID 是否相等
func (id TaskID) Equal(other TaskID) bool {
	return bytes.Equal(id.data[:], other.data[:])
}
