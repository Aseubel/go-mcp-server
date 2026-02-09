package log

import (
	"log/slog"
	"os"
	"time"

	"mcp/config"

	"github.com/lmittmann/tint"
)

var (
	// Logger 全局日志实例
	Logger *slog.Logger = slog.Default()
)

// Init 初始化日志系统
func Init(cfg *config.MCPConfig) {
	// 从配置中获取日志级别
	level := parseLogLevel(cfg.Log.Level)

	// 根据服务器模式确定是否启用调试模式
	debug := cfg.Server.Env != "prod" && cfg.Server.Env != "production"

	var handler slog.Handler

	if debug {
		// 开发环境：使用彩色日志输出（tint）
		// 性能影响极小，主要是 ANSI 转义序列的字符串拼接
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen, // 简洁时间格式 "3:04PM"
			AddSource:  true,         // 显示源码位置
		})
	} else {
		// 生产环境：使用 JSON 格式，便于日志收集和分析
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: false, // 生产环境不需要源码位置
		})
	}

	// 创建日志实例
	Logger = slog.New(handler)

	// 设置为默认日志记录器
	slog.SetDefault(Logger)
}

// parseLogLevel 将字符串日志级别转换为slog.Level
func parseLogLevel(levelStr string) slog.Level {
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		// 默认使用INFO级别
		return slog.LevelInfo
	}
}

// Debug 记录调试级别日志
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

// Info 记录信息级别日志
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// Warn 记录警告级别日志
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

// Error 记录错误级别日志
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

// Fatal 记录致命错误并退出程序
func Fatal(msg string, args ...any) {
	Logger.Error(msg, args...)

	os.Exit(1)
}

// With 创建带有字段的日志记录器
func With(args ...any) *slog.Logger {
	return Logger.With(args...)
}

// WithGroup 创建带有组的日志记录器
func WithGroup(name string) *slog.Logger {
	return Logger.WithGroup(name)
}

