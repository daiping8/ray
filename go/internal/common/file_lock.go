package common

import (
	"context"
	"time"

	"github.com/gofrs/flock"
)

// AsyncFileLock 异步文件锁，是可复用的并发控制组件
type AsyncFileLock struct {
	lock *flock.Flock
}

// NewAsyncFileLock 创建新的异步文件锁
func NewAsyncFileLock(lockFile string) *AsyncFileLock {
	return &AsyncFileLock{
		lock: flock.New(lockFile),
	}
}

// Acquire 获取锁（带上下文支持）
// 使用 TryLockContext 方法，周期性重试直到获取锁或上下文取消
func (l *AsyncFileLock) Acquire(ctx context.Context) error {
	// 使用 100ms 间隔尝试获取锁
	retryDuration := 100 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			acquired, err := l.lock.TryLockContext(ctx, retryDuration)
			if err != nil {
				return err
			}
			if acquired {
				return nil
			}
		}
	}
}

// Release 释放锁
func (l *AsyncFileLock) Release() {
	if l.lock != nil {
		_ = l.lock.Unlock()
	}
}
