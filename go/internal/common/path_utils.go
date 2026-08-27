package common

import (
	"net/url"
	"runtime"
	"strings"
)

// IsPath 检查字符串是否是文件路径而非 URI
//
// 返回 True 如果输入是路径，否则返回 False。
//
// Windows 路径以驱动器名称开头，这可能会被 urlparse 解释为 URI scheme，
// 因此需要与 POSIX 路径区别对待。
//
// 例如：创建目录会返回路径 'C:\Users\mp5n6ul72w\working_dir'，
// 它的 scheme 将是 'C:'。
func IsPath(pathOrURI string) bool {
	parsedURI, err := url.Parse(pathOrURI)
	if err != nil {
		return true
	}

	// 根据运行时操作系统判断路径类型
	if runtime.GOOS == "windows" {
		// Windows 平台：使用 PureWindowsPath 逻辑
		drive := getWindowsDrive(pathOrURI)
		if drive != "" {
			// 盘符路径：scheme 应等于盘符字母（小写），例如 "c" 对应 "C:"
			// pathlib.PureWindowsPath("C:\\path").drive == "C:"
			// urlparse("C:\\path").scheme == "c"
			return strings.ToLower(parsedURI.Scheme) == strings.ToLower(strings.TrimSuffix(drive, ":"))
		}
		// 其他 Windows 路径（包含反斜杠）：没有 scheme 则是路径
		if strings.Contains(pathOrURI, `\`) {
			return parsedURI.Scheme == ""
		}
		// 其他情况：没有 scheme 则是路径
		return parsedURI.Scheme == ""
	}

	// POSIX 平台：使用 PurePosixPath 逻辑 - 没有 scheme 则是路径
	return parsedURI.Scheme == ""
}

// getWindowsDrive 从 Windows 路径获取盘符（如 "C:"）
// 如果是有效的 Windows 盘符路径（如 C:\ 或 C:/），返回盘符加冒号；否则返回空字符串
func getWindowsDrive(path string) string {
	if len(path) < 2 {
		return ""
	}
	c := path[0]
	isLetter := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	if !isLetter || path[1] != ':' {
		return ""
	}
	// 检查第三个字符是否是路径分隔符（\ 或 /）或者已经到字符串末尾
	if len(path) == 2 || path[2] == '\\' || path[2] == '/' {
		return path[:2]
	}
	return ""
}
