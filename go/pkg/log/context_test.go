package log

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
)

func TestFromContext_NoLogger(t *testing.T) {
	// 重置 Log 为初始状态
	Log = logr.New(NewDelegatingLogSink())

	ctx := context.Background()
	logger := FromContext(ctx)
	// 应返回全局 Log
	if logger.GetSink() == nil {
		t.Error("FromContext should return global Log when no logger in context")
	}
}

func TestFromContext_WithLogger(t *testing.T) {
	ctx := context.Background()
	testSink := &testLogSink{}
	testLogger := logr.New(testSink)

	ctxWithLogger := IntoContext(ctx, testLogger)
	retrievedLogger := FromContext(ctxWithLogger)
	retrievedLogger.Info("test message", "key", "value")

	if testSink.lastMsg != "test message" {
		t.Errorf("expected 'test message', got '%s'", testSink.lastMsg)
	}
}

func TestFromContextOrDiscard_NoLogger(t *testing.T) {
	ctx := context.Background()
	logger := FromContextOrDiscard(ctx)
	// Discard logger 不应 panic
	logger.Info("test message")
	// 无错误即通过
}

func TestFromContextOrDiscard_WithLogger(t *testing.T) {
	ctx := context.Background()
	testSink := &testLogSink{}
	testLogger := logr.New(testSink)

	ctxWithLogger := IntoContext(ctx, testLogger)
	retrievedLogger := FromContextOrDiscard(ctxWithLogger)
	retrievedLogger.Info("test message", "key", "value")

	if testSink.lastMsg != "test message" {
		t.Errorf("expected 'test message', got '%s'", testSink.lastMsg)
	}
}

func TestIntoContext(t *testing.T) {
	ctx := context.Background()
	testSink := &testLogSink{}
	testLogger := logr.New(testSink)

	newCtx := IntoContext(ctx, testLogger)
	if newCtx == ctx {
		t.Error("IntoContext should return new context")
	}

	retrievedLogger, ok := newCtx.Value(loggerKey{}).(logr.Logger)
	if !ok {
		t.Error("context should contain logger")
	}
	if retrievedLogger.GetSink() != testLogger.GetSink() {
		t.Error("retrieved logger should match stored logger")
	}
}

func TestFromContext_ReturnsGlobalLog(t *testing.T) {
	// 重置 Log 为初始状态
	Log = logr.New(NewDelegatingLogSink())

	ctx := context.Background()
	logger := FromContext(ctx)

	// 应该返回全局 Log 的 sink
	if logger.GetSink() != Log.GetSink() {
		t.Error("FromContext should return global Log when context has no logger")
	}
}

func TestIntoContext_Overwrite(t *testing.T) {
	ctx := context.Background()

	testSink1 := &testLogSink{}
	testLogger1 := logr.New(testSink1)

	testSink2 := &testLogSink{}
	testLogger2 := logr.New(testSink2)

	// 第一次设置
	ctx1 := IntoContext(ctx, testLogger1)
	// 第二次设置（覆盖）
	ctx2 := IntoContext(ctx1, testLogger2)

	retrievedLogger := FromContext(ctx2)
	retrievedLogger.Info("test")

	if testSink2.lastMsg != "test" {
		t.Error("expected second logger to be used")
	}
}