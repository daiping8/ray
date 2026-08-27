package common

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ray-project/ray/go/pkg/gcs"
	"github.com/ray-project/ray/go/pkg/ids"
	"github.com/ray-project/ray/go/proto"
)

var State = NewGlobalState()

// GlobalState 全局状态管理器
type GlobalState struct {
	gcsOptions          *gcs.ClientOptions
	globalStateAccessor gcs.GlobalStateAccessor
	initLock            sync.Mutex
}

// NewGlobalState 创建新的GlobalState实例
//
// 初始化逻辑：
// 1. 创建空的GlobalState结构体
// 2. gcsOptions初始为nil，由后续InitializeGlobalState设置（对应Python的_initialize_global_state）
// 3. globalStateAccessor初始为nil，首次使用时通过ConnectAndGetAccessor创建
// 4. initLock用于保护并发访问，确保线程安全
func NewGlobalState() *GlobalState {
	return &GlobalState{
		gcsOptions:          nil,
		globalStateAccessor: nil,
		initLock:            sync.Mutex{},
	}
}

// InitializeGlobalState 初始化GlobalState的GCS选项
func (s *GlobalState) InitializeGlobalState(opts *gcs.ClientOptions) {
	s.initLock.Lock()
	defer s.initLock.Unlock()
	s.gcsOptions = opts
}

// ConnectAndGetAccessor 懒加载并返回已连接的GCS状态访问器
//
// 核心逻辑：
// 1. 线程安全：使用initLock确保多线程环境下的原子操作
// 2. 缓存检查：如果已有访问器则直接返回，避免重复连接
// 3. 前置验证：检查gcsOptions是否已设置（即是否已调用InitializeGlobalState）
// 4. 获取访问器：从全局单例获取GlobalStateAccessor
// 5. 连接验证：尝试连接GCS服务器，失败时清理状态并返回错误
func (s *GlobalState) ConnectAndGetAccessor() (gcs.GlobalStateAccessor, error) {
	s.initLock.Lock()
	defer s.initLock.Unlock()

	// 缓存检查：如果已有访问器则直接返回
	if s.globalStateAccessor != nil {
		return s.globalStateAccessor, nil
	}

	// 前置验证：检查gcsOptions是否已设置
	if s.gcsOptions == nil {
		return nil, errors.New("Ray has not been started yet. Trying to use state API before InitializeGlobalState has been called")
	}

	// 获取全局访问器实例
	accessor, err := gcs.GetGlobalStateAccessor()
	if err != nil {
		return nil, err
	}

	// 尝试连接GCS服务器
	connected, err := accessor.Connect()
	if err != nil || !connected {
		s.globalStateAccessor = nil
		// 连接失败时清理状态
		if err == nil {
			err = errors.New("failed to connect to GCS server")
		}
		return nil, err
	}

	// 缓存访问器引用
	s.globalStateAccessor = accessor
	return s.globalStateAccessor, nil
}

// Disconnect 断开与GCS的连接并清理资源
//
// 清理操作：
// 1. 重置gcsOptions为nil，标记需要重新初始化
// 2. 释放globalStateAccessor引用，允许垃圾回收
func (s *GlobalState) Disconnect() error {
	s.initLock.Lock()
	defer s.initLock.Unlock()

	s.gcsOptions = nil
	if s.globalStateAccessor != nil {
		// 关闭访问器连接
		s.globalStateAccessor.Close()
		s.globalStateAccessor = nil
	}

	return nil
}

// AddWorker 向集群中添加Worker信息
//
// 核心功能：
// 1. 获取GCS状态访问器（懒加载连接）
// 2. 构建WorkerTableData protobuf消息
// 3. 设置Worker的基本属性（存活状态、ID、类型）
// 4. 将workerInfo map转换为protobuf的map字段
// 5. 调用accessor.AddWorkerInfo将数据写入GCS
func (s *GlobalState) AddWorker(workerID ids.WorkerID, workerType proto.WorkerType, workerInfo map[string]string) error {
	accessor, err := s.ConnectAndGetAccessor()
	if err != nil {
		return fmt.Errorf("failed to get GCS accessor: %w", err)
	}

	// 转换workerInfo为[]byte格式（protobuf要求）
	byteWorkerInfo := make(map[string][]byte)
	for k, v := range workerInfo {
		byteWorkerInfo[k] = []byte(v)
	}
	// 构建WorkerTableData消息
	workerData := &proto.WorkerTableData{
		IsAlive:    true,
		WorkerType: workerType,
		WorkerInfo: byteWorkerInfo,
		Timestamp:  time.Now().UnixNano() / 1e6, // 毫秒时间戳
	}

	// 设置Worker地址（包含Worker ID）
	// 注意：WorkerId字段是[]byte类型，需要从UniqueID获取二进制数据
	workerData.WorkerAddress = &proto.Address{
		WorkerId: workerID.Binary(),
	}

	// 直接调用accessor.AddWorkerInfo将数据写入GCS
	success, err := accessor.AddWorkerInfo(workerData)
	if err != nil {
		return fmt.Errorf("failed to add worker info to GCS: %w", err)
	}
	if !success {
		return errors.New("add worker info returned false")
	}

	return nil
}
