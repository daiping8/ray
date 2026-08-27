package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RAY_PATH Ray Go 包绝对路径，用于构建 jars 等资源路径
// 对应 Python: os.path.abspath(os.path.dirname(os.path.dirname(__file__)))
var RAY_PATH string

func init() {
	// 优先使用环境变量 RAY_PATH
	rayPath := os.Getenv("RAY_PATH")
	if rayPath == "" {
		// 使用 runtime.Caller(0) 获取当前文件路径，比 os.Args[0] 更可靠
		_, currentFile, _, ok := runtime.Caller(0)
		var currentDir string
		if ok {
			currentDir = filepath.Dir(currentFile)
		} else {
			// 如果无法获取，使用当前工作目录
			currentDir, _ = os.Getwd()
		}
		rayPath = filepath.Dir(filepath.Dir(filepath.Dir(currentDir)))
	}

	RAY_PATH = rayPath
}

// GetRayJarsDir 返回一个目录，其中包含所有 Ray 相关的 jars 及其依赖
// 对应 Python: get_ray_jars_dir()
// 返回：jars 目录的绝对路径
// 错误：如果 jars 目录不存在，返回错误
func GetRayJarsDir() (string, error) {
	jarsDir := filepath.Join(RAY_PATH, "jars")
	absJarsDir, err := filepath.Abs(jarsDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for jars dir: %w", err)
	}

	// 检查 jars 目录是否存在
	if _, err := os.Stat(absJarsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("jars directory does not exist: %s", absJarsDir)
	}

	return absJarsDir, nil
}

// ExpandUser 展开路径中的 ~ 符号为用户主目录
func ExpandUser(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" || path[0] != '~' {
		return path, nil
	}

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// 构建完整路径
	// 移除 ~
	relativePath := path[1:]
	relativePath = strings.TrimSpace(relativePath)
	if len(relativePath) > 0 && relativePath[0] == '/' {
		relativePath = relativePath[1:]
	}

	return filepath.Join(homeDir, relativePath), nil
}

// PathExists 检查路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// PathIsDir 检查路径是否是目录
func PathIsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
