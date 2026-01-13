package log

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
)

var Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

// LoggerWithCaller 返回一个包含当前调用者位置信息的日志Hook
func LoggerWithCaller() zerolog.Logger {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.With().Str("source", "unknown").Logger()
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.With().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName).
		Logger()
}

// Debug 创建一个带有调用位置的Debug级别日志
func Debug() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Debug().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Debug().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Info 创建一个带有调用位置的Info级别日志
func Info() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Info().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Info().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Warn 创建一个带有调用位置的Warn级别日志
func Warn() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Warn().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Warn().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Error 创建一个带有调用位置的Error级别日志
func Error() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Error().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Error().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Fatal 创建一个带有调用位置的Fatal级别日志
func Fatal() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Fatal().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Fatal().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}
