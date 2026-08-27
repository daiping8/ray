package common

import (
	"os"
	"runtime"
	"strconv"
)

// EnvInteger 从环境变量中获取整数值，如果不存在或无法转换则返回默认值
func EnvInteger(key string, defaultVal int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// EnvFloat 从环境变量中获取浮点数值，如果不存在或无法转换则返回默认值
func EnvFloat(key string, defaultVal float64) float64 {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultVal
	}
	return floatVal
}

// EnvBool 从环境变量中获取布尔值，支持 "true"/"1" 为 true，其他为 false
func EnvBool(key string, defaultVal bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}
	return value == "true" || value == "1"
}

// EnvSetByUser 检查环境变量是否被用户设置
func EnvSetByUser(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

// EnvString 从环境变量中获取字符串值，如果不存在则返回默认值
func EnvString(key string, defaultVal string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}
	return value
}

// 日志级别配置 (与 Python ray_constants.LOGGER_LEVEL 对齐)
var (
	// LoggerLevel 默认日志级别 (对应 Python: os.environ.get("RAY_LOGGER_LEVEL", "info"))
	LoggerLevel = EnvString("RAY_LOGGER_LEVEL", "info")
	// LoggerLevelChoices 可选的日志级别列表
	LoggerLevelChoices = []string{"debug", "info", "warning", "error", "critical"}
)

const (
	// DefaultRuntimeEnvTimeoutSeconds 默认运行时环境超时时间（秒）
	DefaultRuntimeEnvTimeoutSeconds = 600

	// Keep in sync with max_grpc_message_size in ray_config_def.h.
	GRPC_CPP_MAX_MESSAGE_SIZE = 250 * 1024 * 1024
)

// 日志格式 (与 Python ray_constants.LOGGER_FORMAT 对齐)
const (
	// LoggerFormat 默认日志格式
	LoggerFormat     = "%(asctime)s\t%(levelname)s %(filename)s:%(lineno)s -- %(message)s"
	LoggerFormatHelp = "The logging format."
)

const (
	MONITOR_LOG_FILE_NAME = "monitor.log"

	// DefaultLoggingDevelopment 默认日志开发/生成模式
	DefaultLoggingDevelopment = true
	// 默认日志轮转大小 (与 Python LOGGING_ROTATE_BYTES 对齐)
	LOGGING_ROTATE_BYTES = 512 * 1024 * 1024 // 512MB
	// 默认日志备份数量
	LOGGING_ROTATE_BACKUP_COUNT = 5

	REDIS_DEFAULT_USERNAME = ""
	REDIS_DEFAULT_PASSWORD = ""

	DEFAULT_DASHBOARD_IP                = "127.0.0.1"
	DEFAULT_DASHBOARD_PORT              = 8265
	DASHBOARD_ADDRESS                   = "dashboard"
	DASHBOARD_CLIENT_MAX_SIZE           = 100 * 1024 * 1024
	PROMETHEUS_SERVICE_DISCOVERY_FILE   = "prom_metrics_service_discovery.json"
	DEFAULT_DASHBOARD_AGENT_LISTEN_PORT = 52365
)

const (
	// IS_WINDOWS_OR_OSX 表示当前运行平台是否为Windows或macOS
	// 注意：这个常量在Go中总是false，因为它是为Python代码准备的
	// Go版本的检测应该在运行时通过runtime.GOOS判断
	IS_WINDOWS_OR_OSX           = false
	ENABLE_RAY_CLUSTERS_ENV_VAR = "RAY_ENABLE_WINDOWS_OR_OSX_CLUSTER"
)

// IsWindowsOrOSX 检查当前平台是否为Windows或macOS
func IsWindowsOrOSX() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "windows"
}

// EnableRayCluster 返回是否启用Ray集群模式
// 默认情况下，Windows和macOS不启用集群模式（除非通过环境变量覆盖）
func EnableRayCluster() bool {
	defaultValue := !IsWindowsOrOSX()
	return EnvBool(ENABLE_RAY_CLUSTERS_ENV_VAR, defaultValue)
}

// MonitorLogRotateBytes 获取日志轮转大小，支持环境变量覆盖
func MonitorLogRotateBytes() int {
	return EnvInteger("RAY_MONITOR_LOG_ROTATE_BYTES", LOGGING_ROTATE_BYTES)
}

// MonitorLogRotateBackupCount 获取日志备份数量，支持环境变量覆盖
func MonitorLogRotateBackupCount() int {
	return EnvInteger("RAY_MONITOR_LOG_ROTATE_BACKUP_COUNT", LOGGING_ROTATE_BACKUP_COUNT)
}
