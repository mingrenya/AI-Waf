package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
)

// GlobalLogger 在 Logger 未初始化时使用
var GlobalLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()

// Logger 全局日志实例
var Logger zerolog.Logger

type LogConfig struct {
	Level  string
	File   string
	Format string
}

func (lc LogConfig) outputWriter() (io.Writer, error) {
	var out io.Writer
	if lc.File == "" || lc.File == "/dev/stdout" {
		out = os.Stdout
	} else if lc.File == "/dev/stderr" {
		out = os.Stderr
	} else if lc.File == "/dev/null" {
		out = io.Discard
	} else {
		f, err := os.OpenFile(lc.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		out = f
	}
	return out, nil
}

func (lc LogConfig) newLogger() (zerolog.Logger, error) {
	out, err := lc.outputWriter()
	if err != nil {
		return GlobalLogger, err
	}

	switch lc.Format {
	case "console":
		out = zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: "2006-01-02 15:04:05",
			FormatLevel: func(i interface{}) string {
				if ll, ok := i.(string); ok {
					switch ll {
					case "debug":
						return "\x1b[36m[DEBUG]\x1b[0m" // 青色
					case "info":
						return "\x1b[32m[ INFO]\x1b[0m" // 绿色
					case "warn":
						return "\x1b[33m[ WARN]\x1b[0m" // 黄色
					case "error":
						return "\x1b[31m[ERROR]\x1b[0m" // 红色
					case "fatal":
						return "\x1b[35m[FATAL]\x1b[0m" // 紫色
					}
				}
				return fmt.Sprintf("[%-5s]", i)
			},
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("\x1b[1m%s\x1b[0m", i) // 加粗消息
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("\x1b[34m%s:\x1b[0m", i) // 蓝色字段名
			},
			FormatFieldValue: func(i interface{}) string {
				return fmt.Sprintf("\x1b[37m%s\x1b[0m", i) // 浅灰色字段值
			},
		}
	case "json":
		// 使用默认的JSON格式
	default:
		return GlobalLogger, fmt.Errorf("unknown log format: %v", lc.Format)
	}

	if lc.Level == "" {
		lc.Level = "info"
	}
	lvl, err := zerolog.ParseLevel(lc.Level)
	if err != nil {
		return GlobalLogger, err
	}

	return zerolog.New(out).Level(lvl).With().Timestamp().Logger(), nil
}

// 基于 全局 Logger 创建包含调用者信息的日志器
// GetLogger 获取包含调用者信息的日志器
func GetLogger() zerolog.Logger {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.With().Str("source", "unknown").Logger()
	}

	// 提取文件名（不含路径）
	fileName := filepath.Base(file)

	return Logger.With().
		Str("file", fileName).
		Int("line", line).
		Logger()
}

// GetServiceLogger 获取服务层日志器
func GetServiceLogger(serviceName string) zerolog.Logger {
	_, file, line, _ := runtime.Caller(1)
	fileName := filepath.Base(file)

	return Logger.With().
		Str("layer", "service").
		Str("service", serviceName).
		Str("file", fileName).
		Int("line", line).
		Logger()
}

// GetControllerLogger 获取控制器层日志器
func GetControllerLogger(controllerName string) zerolog.Logger {
	_, file, line, _ := runtime.Caller(1)
	fileName := filepath.Base(file)

	return Logger.With().
		Str("layer", "controller").
		Str("controller", controllerName).
		Str("file", fileName).
		Int("line", line).
		Logger()
}

// GetRepositoryLogger 获取数据访问层日志器
func GetRepositoryLogger(repoName string) zerolog.Logger {
	_, file, line, _ := runtime.Caller(1)
	fileName := filepath.Base(file)

	return Logger.With().
		Str("layer", "repository").
		Str("repository", repoName).
		Str("file", fileName).
		Int("line", line).
		Logger()
}
