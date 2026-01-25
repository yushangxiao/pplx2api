package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// 日志级别
const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[int]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var levelColors = map[int]func(format string, a ...interface{}) string{
	DEBUG: color.BlueString,
	INFO:  color.GreenString,
	WARN:  color.YellowString,
	ERROR: color.RedString,
	FATAL: color.New(color.FgHiRed, color.Bold).SprintfFunc(),
}

// 全局日志级别，默认为INFO
var logLevel = INFO

// SetLevel 设置日志级别
func SetLevel(level int) {
	if level >= DEBUG && level <= FATAL {
		logLevel = level
	}
}

// GetLevel 获取当前日志级别
func GetLevel() int {
	return logLevel
}

// GetLevelName 获取日志级别名称
func GetLevelName(level int) string {
	if name, ok := levelNames[level]; ok {
		return name
	}
	return "UNKNOWN"
}

// 基础日志打印函数
func log(level int, format string, args ...interface{}) {
	if level < logLevel {
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05.000")
	levelName := levelNames[level]
	colorFunc := levelColors[level]

	logContent := fmt.Sprintf(format, args...)
	logPrefix := fmt.Sprintf("[%s] [%s] ", now, levelName)

	// 使用颜色输出日志级别
	fmt.Fprintf(os.Stdout, "%s%s\n", logPrefix, colorFunc(logContent))

	// 如果是致命错误，则退出程序
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 打印调试日志
func Debug(format string, args ...interface{}) {
	log(DEBUG, format, args...)
}

// Info 打印信息日志
func Info(format string, args ...interface{}) {
	log(INFO, format, args...)
}

// Warn 打印警告日志
func Warn(format string, args ...interface{}) {
	log(WARN, format, args...)
}

// Error 打印错误日志
func Error(format string, args ...interface{}) {
	log(ERROR, format, args...)
}

// Fatal 打印致命错误日志并退出程序
func Fatal(format string, args ...interface{}) {
	log(FATAL, format, args...)
}
