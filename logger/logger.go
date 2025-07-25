package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// 日志配置
type Config struct {
	Level          int
	Format         string
	Output         string
	FilePath       string
	FileMaxSize    int64
	FileMaxAge     int
	FileMaxBackups int
}

// 全局日志配置
var logConfig = &Config{
	Level:          INFO,
	Format:         "color",
	Output:         "stdout",
	FilePath:       "",
	FileMaxSize:    100,
	FileMaxAge:     30,
	FileMaxBackups: 3,
}

// 全局日志级别（向后兼容）
var logLevel = INFO

// 日志文件句柄
var logFile *os.File

// Init 初始化日志配置
func Init() {
	// 从环境变量读取配置
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		logConfig.Level = parseLogLevel(levelStr)
		logLevel = logConfig.Level // 保持向后兼容
	}

	if format := os.Getenv("LOG_FORMAT"); format != "" {
		logConfig.Format = format
	}

	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		logConfig.Output = output
	}

	if filePath := os.Getenv("LOG_FILE_PATH"); filePath != "" {
		logConfig.FilePath = filePath
	}

	if maxSizeStr := os.Getenv("LOG_FILE_MAX_SIZE"); maxSizeStr != "" {
		if maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			logConfig.FileMaxSize = maxSize
		}
	}

	if maxAgeStr := os.Getenv("LOG_FILE_MAX_AGE"); maxAgeStr != "" {
		if maxAge, err := strconv.Atoi(maxAgeStr); err == nil {
			logConfig.FileMaxAge = maxAge
		}
	}

	if maxBackupsStr := os.Getenv("LOG_FILE_MAX_BACKUPS"); maxBackupsStr != "" {
		if maxBackups, err := strconv.Atoi(maxBackupsStr); err == nil {
			logConfig.FileMaxBackups = maxBackups
		}
	}

	// 如果需要文件输出，初始化日志文件
	if (logConfig.Output == "file" || logConfig.Output == "both") && logConfig.FilePath != "" {
		initLogFile()
	}
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(levelStr string) int {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// SetLevel 设置日志级别（向后兼容）
func SetLevel(level int) {
	if level >= DEBUG && level <= FATAL {
		logLevel = level
		logConfig.Level = level
	}
}

// GetLevel 获取当前日志级别（向后兼容）
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

// initLogFile 初始化日志文件
func initLogFile() {
	// 确保日志目录存在
	dir := filepath.Dir(logConfig.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		return
	}

	// 打开日志文件（追加模式）
	file, err := os.OpenFile(logConfig.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		return
	}

	logFile = file

	// 启动日志轮转检查
	go checkLogRotation()
}

// checkLogRotation 检查并执行日志轮转
func checkLogRotation() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if logFile == nil || logConfig.FilePath == "" {
			continue
		}

		// 检查文件大小
		if info, err := logFile.Stat(); err == nil {
			// 转换为字节（MB转为字节）
			maxSizeBytes := logConfig.FileMaxSize * 1024 * 1024
			if info.Size() >= maxSizeBytes {
				rotateLog()
			}
		}

		// 检查文件年龄
		if logConfig.FileMaxAge > 0 {
			if info, err := logFile.Stat(); err == nil {
				if time.Since(info.ModTime()) > time.Duration(logConfig.FileMaxAge)*24*time.Hour {
					rotateLog()
				}
			}
		}

		// 清理旧的日志文件
		cleanupOldLogs()
	}
}

// rotateLog 执行日志轮转
func rotateLog() {
	if logFile == nil {
		return
	}

	// 关闭当前日志文件
	logFile.Close()

	// 重命名当前日志文件
	backupName := fmt.Sprintf("%s.%s", logConfig.FilePath, time.Now().Format("2006-01-02T15-04-05"))
	if err := os.Rename(logConfig.FilePath, backupName); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rename log file: %v\n", err)
		return
	}

	// 重新打开日志文件
	initLogFile()
}

// cleanupOldLogs 清理旧的日志文件
func cleanupOldLogs() {
	if logConfig.FilePath == "" || logConfig.FileMaxBackups <= 0 {
		return
	}

	// 获取日志文件目录
	dir := filepath.Dir(logConfig.FilePath)
	baseName := filepath.Base(logConfig.FilePath)

	// 读取目录中的文件
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// 收集匹配的日志文件
	var logFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if strings.HasPrefix(name, baseName+".") && name != baseName {
			logFiles = append(logFiles, filepath.Join(dir, name))
		}
	}

	// 如果文件数量超过最大备份数，删除最旧的文件
	if len(logFiles) > logConfig.FileMaxBackups {
		// 按修改时间排序
		fileInfos := make([]os.FileInfo, len(logFiles))
		for i, file := range logFiles {
			if info, err := os.Stat(file); err == nil {
				fileInfos[i] = info
			}
		}

		// 简单按文件名排序，保留最新的几个文件
		for i := 0; i < len(logFiles)-logConfig.FileMaxBackups; i++ {
			os.Remove(logFiles[i])
		}
	}
}

// formatLog 根据配置格式化日志
func formatLog(level int, format string, args ...interface{}) string {
	now := time.Now().Format("2006-01-02 15:04:05.000")
	levelName := levelNames[level]
	logContent := fmt.Sprintf(format, args...)

	switch logConfig.Format {
	case "json":
		return fmt.Sprintf(`{"time":"%s","level":"%s","message":"%s"}`, now, levelName, logContent)
	case "text":
		return fmt.Sprintf("[%s] [%s] %s", now, levelName, logContent)
	case "color":
		fallthrough
	default:
		colorFunc := levelColors[level]
		logPrefix := fmt.Sprintf("[%s] [%s] ", now, levelName)
		return fmt.Sprintf("%s%s", logPrefix, colorFunc(logContent))
	}
}

// outputLog 输出日志到指定目标
func outputLog(formattedLog string) {
	switch logConfig.Output {
	case "file":
		// 只输出到文件
	case "both":
		fallthrough
	case "stdout":
		fallthrough
	default:
		fmt.Fprintf(os.Stdout, "%s\n", formattedLog)
	}

	// 输出到文件
	if (logConfig.Output == "file" || logConfig.Output == "both") && logConfig.FilePath != "" {
		if logFile != nil {
			fmt.Fprintf(logFile, "%s\n", formattedLog)
			logFile.Sync() // 确保写入磁盘
		} else {
			// 如果文件未初始化，输出到标准错误
			fmt.Fprintf(os.Stderr, "Log file not initialized: %s\n", formattedLog)
		}
	}
}

// 基础日志打印函数
func log(level int, format string, args ...interface{}) {
	if level < logConfig.Level {
		return
	}

	formattedLog := formatLog(level, format, args...)
	outputLog(formattedLog)

	// 如果是致命错误，则退出程序
	if level == FATAL {
		if logFile != nil {
			logFile.Close()
		}
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
