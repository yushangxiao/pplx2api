package config

import (
	"os"
	"time"
)

// IsRenderEnvironment 检查是否在Render云平台环境
func IsRenderEnvironment() bool {
	return os.Getenv("RENDER") == "true" || os.Getenv("DYNO") != "" || os.Getenv("VERCEL") != "" || os.Getenv("NETLIFY") != ""
}

// GetRenderOptimizedTimeout 获取Render环境优化的超时时间
func GetRenderOptimizedTimeout() time.Duration {
	if IsRenderEnvironment() {
		// Render环境使用较短的超时时间
		return time.Minute * 2
	}
	// 非Render环境使用默认超时
	return time.Minute * 10
}

// GetRenderOptimizedBufferSize 获取Render环境优化的缓冲区大小
func GetRenderOptimizedBufferSize() (int, int) {
	if IsRenderEnvironment() {
		// Render环境使用较小的缓冲区
		return 256 * 1024, 512 * 1024 // 256KB初始，512KB最大
	}
	// 非Render环境使用默认缓冲区
	return 1024 * 1024, 1024 * 1024 // 1MB初始和最大
}

// ShouldEnableSessionUpdater 检查是否应该启用会话更新器
func ShouldEnableSessionUpdater() bool {
	// 在Render等云平台环境中禁用会话自动更新
	return !IsRenderEnvironment()
}
