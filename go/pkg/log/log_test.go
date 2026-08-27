package log

import (
	"testing"

	"github.com/go-logr/logr"
)

func TestSetLogger_NilLogger(t *testing.T) {
	// 重置 Log 为初始状态
	Log = logr.New(NewDelegatingLogSink())

	SetLogger(logr.Logger{})
	Log.Info("test message", "key", "value")
	Log.Error(nil, "test error", "key", "value")
	// 不应 panic
}

func TestSetLogger_ValidLogger(t *testing.T) {
	// 重置 Log 为初始状态
	Log = logr.New(NewDelegatingLogSink())

	testSink := &testLogSink{}
	testLogger := logr.New(testSink)
	SetLogger(testLogger)
	Log.Info("test message", "key", "value")

	if testSink.lastMsg != "test message" {
		t.Errorf("expected 'test message', got '%s'", testSink.lastMsg)
	}
	if len(testSink.lastKeysAndValues) != 2 {
		t.Errorf("expected 2 keysAndValues, got %d", len(testSink.lastKeysAndValues))
	}
	if testSink.lastKeysAndValues[0] != "key" || testSink.lastKeysAndValues[1] != "value" {
		t.Errorf("expected ['key', 'value'], got %v", testSink.lastKeysAndValues)
	}
}

func TestSetLogger_ReplaceLogger(t *testing.T) {
	// 重置 Log 为初始状态
	Log = logr.New(NewDelegatingLogSink())

	// 第一次设置
	testSink1 := &testLogSink{}
	testLogger1 := logr.New(testSink1)
	SetLogger(testLogger1)

	// 第二次设置（直接替换，不再使用 DelegatingLogSink）
	testSink2 := &testLogSink{}
	testLogger2 := logr.New(testSink2)
	Log = testLogger2

	Log.Info("second logger message")
	if testSink2.lastMsg != "second logger message" {
		t.Errorf("expected 'second logger message', got '%s'", testSink2.lastMsg)
	}
}

// TestSetLogger_NonDelegatingLogger 测试替换非 DelegatingLogSink
func TestSetLogger_NonDelegatingLogger(t *testing.T) {
	// 重置 Log 为初始状态
	Log = logr.New(NewDelegatingLogSink())

	// 先设置一个普通 logger
	testSink1 := &testLogSink{}
	testLogger1 := logr.New(testSink1)
	SetLogger(testLogger1)
	Log.Info("first message", "key", "value1")

	// 替换为另一个 logger
	testSink2 := &testLogSink{}
	testLogger2 := logr.New(testSink2)
	SetLogger(testLogger2)
	Log.Info("second message", "key", "value2")

	if testSink2.lastMsg != "second message" {
		t.Errorf("expected 'second message', got '%s'", testSink2.lastMsg)
	}
}

func TestWithName(t *testing.T) {
	// 重置 Log 为初始状态
	Log = logr.New(NewDelegatingLogSink())

	testSink := &testLogSink{}
	testLogger := logr.New(testSink)
	SetLogger(testLogger)

	namedLogger := WithName("worker")
	namedLogger.Info("named message")

	// WithName 会将名称添加到消息前缀
	if testSink.lastMsg != "worker: named message" {
		t.Errorf("expected 'worker: named message', got '%s'", testSink.lastMsg)
	}
}

func TestWithValues(t *testing.T) {
	// 重置 Log 为初始状态
	Log = logr.New(NewDelegatingLogSink())

	testSink := &testLogSink{}
	testLogger := logr.New(testSink)
	SetLogger(testLogger)

	valuedLogger := WithValues("component", "worker", "id", "123")
	valuedLogger.Info("valued message")

	if testSink.lastMsg != "valued message" {
		t.Errorf("expected 'valued message', got '%s'", testSink.lastMsg)
	}
	// WithValues 添加的键值对会自动附加到每条日志
	if len(testSink.lastKeysAndValues) < 4 {
		t.Errorf("expected at least 4 keysAndValues, got %d", len(testSink.lastKeysAndValues))
	}
}