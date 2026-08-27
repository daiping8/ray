package common

import (
	"io"
	"os"
	"path/filepath"
)

// InterfaceToStringSlice 将 interface{} 切片转换为 string 切片
// 非 string 元素会被转换为空字符串
func InterfaceToStringSlice(v []interface{}) []string {
	if v == nil {
		return nil
	}
	result := make([]string, len(v))
	for i, item := range v {
		if s, ok := item.(string); ok {
			result[i] = s
		}
	}
	return result
}

// InterfaceMapToStringMap 将 interface{} map 转换为 string map
// 非 string 值会被跳过
func InterfaceMapToStringMap(v map[string]interface{}) map[string]string {
	if v == nil {
		return nil
	}
	result := make(map[string]string)
	for k, val := range v {
		if s, ok := val.(string); ok {
			result[k] = s
		}
	}
	return result
}

// ConvertSlice 将 interface{} 值转换为 []T 类型
// 如果 val 已经是 []T 类型，直接返回
// 如果 val 是 []interface{} 类型，使用 converter 进行转换
// 否则返回默认值
func ConvertSlice[T any](val interface{}, defaultValue []T, converter func([]interface{}) []T) []T {
	if arr, ok := val.([]T); ok {
		return arr
	}
	if arr, ok := val.([]interface{}); ok {
		return converter(arr)
	}
	return defaultValue
}

// ConvertMap 将 interface{} 值转换为 map[string]T 类型
// 如果 val 已经是 map[string]T 类型，直接返回
// 如果 val 是 map[string]interface{} 类型，使用 converter 进行转换
// 否则返回默认值
func ConvertMap[T any](val interface{}, defaultValue map[string]T, converter func(map[string]interface{}) map[string]T) map[string]T {
	if m, ok := val.(map[string]T); ok {
		return m
	}
	if m, ok := val.(map[string]interface{}); ok {
		return converter(m)
	}
	return defaultValue
}

// CopyAll 复制文件或目录
// 如果 src 是目录，则递归复制整个目录
// 如果 src 是文件，则复制单个文件
func CopyAll(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyDir(src, dst)
	}
	return CopyFile(src, dst)
}

// CopyFile 复制单个文件
// 打开源文件并创建目标文件，然后使用 io.Copy 进行内容复制
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CopyDir 递归复制目录
// 创建目标目录并遍历源目录的所有条目，递归复制每个子项
func CopyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if err := CopyAll(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// DirSizeBytes 计算目录总大小（字节）
func DirSizeBytes(dirPath string) (int64, error) {
	var totalSize int64 = 0

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return totalSize, nil
}

// DeduplicateStrings 去重字符串切片，保持原有顺序
func DeduplicateStrings(items []string) []string {
	if items == nil {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
