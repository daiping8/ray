package log

import (
	"errors"
	"testing"

	"github.com/go-logr/logr"
)

func TestNewDelegatingLogSink(t *testing.T) {
	delegate := NewDelegatingLogSink()
	if delegate == nil {
		t.Error("NewDelegatingLogSink should return non-nil")
	}
	if delegate.fulfilled {
		t.Error("new DelegatingLogSink should not be fulfilled")
	}
}

func TestDelegatingLogSink_Fulfill(t *testing.T) {
	delegate := NewDelegatingLogSink()
	logger := logr.New(delegate)

	// fulfill 前记录日志（使用 NullLogSink，无输出）
	logger.Info("before fulfill", "key", "value1")

	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	// fulfill 后记录日志
	logger.Info("after fulfill", "key", "value2")

	if testSink.lastMsg != "after fulfill" {
		t.Errorf("expected 'after fulfill', got '%s'", testSink.lastMsg)
	}
}

func TestDelegatingLogSink_Enabled_BeforeFulfill(t *testing.T) {
	delegate := NewDelegatingLogSink()
	// fulfill 前，使用 NullLogSink，Enabled 返回 false
	if delegate.Enabled(0) {
		t.Error("Enabled should return false before fulfill")
	}
}

func TestDelegatingLogSink_Enabled_AfterFulfill(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	// fulfill 后，使用 testLogSink，Enabled 返回 true
	if !delegate.Enabled(0) {
		t.Error("Enabled should return true after fulfill")
	}
}

func TestDelegatingLogSink_Error(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate)
	testErr := errors.New("test error")
	logger.Error(testErr, "error message", "key", "value")

	if testSink.lastMsg != "error message" {
		t.Errorf("expected 'error message', got '%s'", testSink.lastMsg)
	}
	if testSink.lastErr != testErr {
		t.Error("expected error to be recorded")
	}
}

func TestDelegatingLogSink_WithName(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate)
	loggerWithName := logger.WithName("worker")
	loggerWithName.Info("test message", "key", "value")

	// WithName 应在消息前添加 "worker: "
	if testSink.lastMsg != "worker: test message" {
		t.Errorf("expected 'worker: test message', got '%s'", testSink.lastMsg)
	}
}

func TestDelegatingLogSink_WithValues(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate)
	loggerWithValues := logger.WithValues("preset_key", "preset_value")
	loggerWithValues.Info("test with values", "dynamic_key", "dynamic_value")

	found := false
	for i := 0; i < len(testSink.lastKeysAndValues); i += 2 {
		if testSink.lastKeysAndValues[i] == "preset_key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected preset_key in keysAndValues")
	}
}

func TestDelegatingLogSink_WithNameAndValues(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate)
	loggerWithNameAndValues := logger.WithName("component").WithValues("preset_key", "preset_value")
	loggerWithNameAndValues.Info("test message", "dynamic_key", "dynamic_value")

	// 检查名称前缀
	if testSink.lastMsg != "component: test message" {
		t.Errorf("expected 'component: test message', got '%s'", testSink.lastMsg)
	}

	// 检查预设键值对存在
	found := false
	for i := 0; i < len(testSink.lastKeysAndValues); i += 2 {
		if testSink.lastKeysAndValues[i] == "preset_key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected preset_key in keysAndValues")
	}
}

func TestDelegatingLogSink_DoubleFulfill(t *testing.T) {
	delegate := NewDelegatingLogSink()

	testSink1 := &testLogSink{}
	delegate.Fulfill(testSink1)

	testSink2 := &testLogSink{}
	delegate.Fulfill(testSink2)

	// 第二次 fulfill 应覆盖第一次
	logger := logr.New(delegate)
	logger.Info("test message")

	if testSink2.lastMsg != "test message" {
		t.Error("expected second sink to receive message")
	}
}

func TestDelegatingLogSink_BeforeFulfill_WithValues(t *testing.T) {
	delegate := NewDelegatingLogSink()
	logger := logr.New(delegate)

	// fulfill 前使用 WithValues
	loggerWithValues := logger.WithValues("preset_key", "preset_value")

	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	// fulfill 后记录日志
	loggerWithValues.Info("test message", "dynamic_key", "dynamic_value")

	// 预设的键值对应被包含
	found := false
	for i := 0; i < len(testSink.lastKeysAndValues); i += 2 {
		if testSink.lastKeysAndValues[i] == "preset_key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected preset_key in keysAndValues")
	}
}

func TestDelegatingLogSink_BeforeFulfill_WithName(t *testing.T) {
	delegate := NewDelegatingLogSink()
	logger := logr.New(delegate)

	// fulfill 前使用 WithName
	loggerWithName := logger.WithName("worker")

	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	// fulfill 后记录日志
	loggerWithName.Info("test message")

	if testSink.lastMsg != "worker: test message" {
		t.Errorf("expected 'worker: test message', got '%s'", testSink.lastMsg)
	}
}

// TestDelegatingLogSink_NestedWithName_Error 测试嵌套 WithName 的 Error 方法
func TestDelegatingLogSink_NestedWithName_Error(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate).WithName("worker")
	logger.Error(errors.New("test error"), "error message", "key", "value")

	if testSink.lastErr == nil {
		t.Error("expected error to be recorded")
	}
}

// TestDelegatingLogSink_NestedWithName_WithValues 测试 WithName 后 WithValues
func TestDelegatingLogSink_NestedWithName_WithValues(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate).WithName("worker").WithValues("preset", "value")
	logger.Info("test message", "dynamic", "key")

	// 检查消息包含 name
	if testSink.lastName != "worker" {
		t.Errorf("expected name 'worker', got '%s'", testSink.lastName)
	}
}

// TestDelegatingLogSink_NestedWithValues_WithName 测试 WithValues 后 WithName
func TestDelegatingLogSink_NestedWithValues_WithName(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate).WithValues("preset", "value").WithName("worker")
	logger.Info("test message", "dynamic", "key")

	if testSink.lastName != "worker" {
		t.Errorf("expected name 'worker', got '%s'", testSink.lastName)
	}
}

// TestDelegatingLogSink_ChainedWithValues 测试链式 WithValues
func TestDelegatingLogSink_ChainedWithValues(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate).WithValues("key1", "val1").WithValues("key2", "val2")
	logger.Info("test message")

	// 检查所有键值对存在
	foundKey1 := false
	foundKey2 := false
	for i := 0; i < len(testSink.lastKeysAndValues); i += 2 {
		if testSink.lastKeysAndValues[i] == "key1" {
			foundKey1 = true
		}
		if testSink.lastKeysAndValues[i] == "key2" {
			foundKey2 = true
		}
	}
	if !foundKey1 || !foundKey2 {
		t.Error("expected both key1 and key2 in keysAndValues")
	}
}

// TestDelegatingLogSink_ChainedWithName 测试链式 WithName
func TestDelegatingLogSink_ChainedWithName(t *testing.T) {
	delegate := NewDelegatingLogSink()
	testSink := &testLogSink{}
	delegate.Fulfill(testSink)

	logger := logr.New(delegate).WithName("worker").WithName("task")
	logger.Info("test message")

	if testSink.lastName != "worker.task" {
		t.Errorf("expected name 'worker.task', got '%s'", testSink.lastName)
	}
}