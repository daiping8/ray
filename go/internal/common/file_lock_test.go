package common

import (
	"context"
	"testing"
	"time"
)

func TestNewAsyncFileLock(t *testing.T) {
	// 创建一个临时文件用于锁测试
	tmpDir := t.TempDir()
	lockFile := tmpDir + "/test.lock"

	lock := NewAsyncFileLock(lockFile)
	if lock == nil {
		t.Fatal("expected NewAsyncFileLock to return non-nil lock")
	}
	if lock.lock == nil {
		t.Error("expected lock.lock to be initialized")
	}
}

func TestAsyncFileLock_AcquireAndRelease(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := tmpDir + "/test.lock"

	lock := NewAsyncFileLock(lockFile)
	ctx := context.Background()

	// 获取锁
	err := lock.Acquire(ctx)
	if err != nil {
		t.Errorf("expected no error acquiring lock, got %v", err)
	}

	// 释放锁
	lock.Release()
	// Release 不应该 panic，即使已经释放
	lock.Release()
}

func TestAsyncFileLock_AcquireWithContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := tmpDir + "/test.lock"

	lock1 := NewAsyncFileLock(lockFile)
	lock2 := NewAsyncFileLock(lockFile)

	ctx := context.Background()

	// 第一个锁获取锁
	err := lock1.Acquire(ctx)
	if err != nil {
		t.Fatalf("expected no error acquiring first lock, got %v", err)
	}
	defer lock1.Release()

	// 创建一个带超时的上下文用于第二个锁
	ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 第二个锁尝试获取锁，应该超时
	err = lock2.Acquire(ctx2)
	if err == nil {
		t.Error("expected timeout error when acquiring lock held by another process")
	}
}

func TestAsyncFileLock_AcquireWithCancelledContext(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := tmpDir + "/test.lock"

	lock := NewAsyncFileLock(lockFile)

	// 使用已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	err := lock.Acquire(ctx)
	if err == nil {
		t.Error("expected error when acquiring lock with cancelled context")
	}
}
