package common

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRayHomeAndPath 测试 RAY_PATH 变量的初始化
func TestRayHomeAndPath(t *testing.T) {
	// 验证 RAY_PATH 不为空
	if RAY_PATH == "" {
		t.Error("RAY_PATH should not be empty")
	}

	// 验证 RAY_PATH 是一个有效的目录
	info, err := os.Stat(RAY_PATH)
	if err != nil {
		t.Errorf("RAY_PATH should be a valid directory: %v", err)
	} else if !info.IsDir() {
		t.Error("RAY_PATH should be a directory")
	}
}

// TestGetRayJarsDir 测试 GetRayJarsDir 函数
func TestGetRayJarsDir(t *testing.T) {
	// 保存原始的 RAY_PATH
	originalRayPath := RAY_PATH
	defer func() {
		RAY_PATH = originalRayPath
	}()

	// 测试 1: 当 jars 目录不存在时，应该返回错误
	t.Run("jars directory not exists", func(t *testing.T) {
		// 设置一个不存在的临时目录作为 RAY_PATH
		tempDir := t.TempDir()
		RAY_PATH = tempDir

		_, err := GetRayJarsDir()
		if err == nil {
			t.Error("GetRayJarsDir should return error when jars directory does not exist")
		}
	})

	// 测试 2: 当 jars 目录存在时，应该返回正确的路径
	t.Run("jars directory exists", func(t *testing.T) {
		// 创建临时目录结构
		tempDir := t.TempDir()
		jarsDir := filepath.Join(tempDir, "jars")
		err := os.Mkdir(jarsDir, 0755)
		if err != nil {
			t.Fatalf("failed to create jars dir: %v", err)
		}

		RAY_PATH = tempDir

		result, err := GetRayJarsDir()
		if err != nil {
			t.Errorf("GetRayJarsDir should not return error: %v", err)
		}

		expectedPath, _ := filepath.Abs(jarsDir)
		if result != expectedPath {
			t.Errorf("GetRayJarsDir = %q, want %q", result, expectedPath)
		}
	})
}

// TestGetRayJarsDirWithEnv 测试使用环境变量设置 RAY_PATH 的情况
func TestGetRayJarsDirWithEnv(t *testing.T) {
	// 保存原始的环境变量和 RAY_PATH
	originalRayPath := os.Getenv("RAY_PATH")
	originalRayPathVar := RAY_PATH
	defer func() {
		if originalRayPath == "" {
			os.Unsetenv("RAY_PATH")
		} else {
			os.Setenv("RAY_PATH", originalRayPath)
		}
		RAY_PATH = originalRayPathVar
	}()

	// 设置临时的 RAY_PATH 环境变量
	tempDir := t.TempDir()
	jarsDir := filepath.Join(tempDir, "jars")
	err := os.Mkdir(jarsDir, 0755)
	if err != nil {
		t.Fatalf("failed to create jars dir: %v", err)
	}

	os.Setenv("RAY_PATH", tempDir)

	// 重新初始化 RAY_PATH（模拟 init 函数的行为）
	RAY_PATH = os.Getenv("RAY_PATH")

	result, err := GetRayJarsDir()
	if err != nil {
		t.Errorf("GetRayJarsDir should not return error: %v", err)
	}

	expectedPath, _ := filepath.Abs(jarsDir)
	if result != expectedPath {
		t.Errorf("GetRayJarsDir = %q, want %q", result, expectedPath)
	}
}
