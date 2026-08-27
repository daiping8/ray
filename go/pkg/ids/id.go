package ids

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash"
	"sync"
	"unsafe"
)

// ID 大小常量定义
const (
	UniqueIDSize                    = 28
	JobIDSize                       = 4
	ActorIDSize                     = 16
	ActorIDUniqueBytesSize          = 12
	TaskIDSize                      = 24
	TaskIDUniqueBytesSize           = 8
	ObjectIDSize                    = 28
	ObjectIDIndexSize               = 4
	PlacementGroupIDSize            = 18
	PlacementGroupIDUniqueBytesSize = 14
	LeaseIDSize                     = 32
	LeaseIDUniqueBytesSize          = 4
)

// MaxObjectIndex 最大对象索引值
const MaxObjectIndex int64 = (1 << ObjectIDIndexSize) - 1

// idToHex 将字节数组转换为十六进制字符串
func idToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// decodeHexToBytes 将十六进制字符串解码到目标字节数组
// 直接解码到目标数组，避免双重分配
// 优化：使用 unsafe 避免堆分配（只读场景）
func decodeHexToBytes(dst []byte, hexStr string) error {
	expectedLen := len(dst)
	if len(hexStr) != 2*expectedLen {
		return errors.New("invalid hex string length")
	}
	// 使用 unsafe 避免堆分配（只读场景）
	hexBytes := unsafe.Slice(unsafe.StringData(hexStr), len(hexStr))
	n, err := hex.Decode(dst, hexBytes)
	if err != nil {
		return err
	}
	if n != expectedLen {
		return errors.New("hex decode length mismatch")
	}
	return nil
}

// fillRandom 用随机字节填充给定切片
func fillRandom(data []byte) {
	_, err := rand.Read(data)
	if err != nil {
		panic("failed to generate random bytes: " + err.Error())
	}
}

// hasherPool 用于复用 sha256 hasher，减少高吞吐场景下的 GC 压力
var hasherPool = sync.Pool{
	New: func() interface{} { return sha256.New() },
}

// generateUniqueBytesInto 直接将唯一字节写入目标数组
// 优化：避免 slice 逃逸到堆导致的双重分配
func generateUniqueBytesInto(dst []byte, jobID JobID, parentTaskID TaskID,
	counter uint64, extra int64,
) {
	h := hasherPool.Get().(hash.Hash)
	defer hasherPool.Put(h)
	h.Reset()

	// 写入 jobID
	h.Write(jobID.data[:])

	// 写入 parentTaskID
	h.Write(parentTaskID.data[:])

	// 写入 counter（使用栈分配避免堆分配）
	var counterBytes [8]byte
	binary.LittleEndian.PutUint64(counterBytes[:], counter)
	h.Write(counterBytes[:])

	// 写入 extra（可选，使用栈分配避免堆分配）
	if extra != 0 {
		var extraBytes [8]byte
		binary.LittleEndian.PutUint64(extraBytes[:], uint64(extra))
		h.Write(extraBytes[:])
	}

	// 使用栈分配的数组接收哈希值
	var hashBuf [sha256.Size]byte
	h.Sum(hashBuf[:0])
	copy(dst, hashBuf[:len(dst)])
}

// generateUniqueBytes 生成唯一字节用于 ActorID、TaskID 等
// 使用 SHA256 替代 C++ SHA256 实现
// 注意：此函数返回的 slice 会逃逸到堆，建议使用 generateUniqueBytesInto 直接写入目标
func generateUniqueBytes(jobID JobID, parentTaskID TaskID,
	counter uint64, extra int64, length int,
) []byte {
	var buf [sha256.Size]byte
	generateUniqueBytesInto(buf[:length], jobID, parentTaskID, counter, extra)
	return buf[:length]
}
