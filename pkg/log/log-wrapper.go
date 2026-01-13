package log

import (
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
)

// LoggerWrapper 是对zerolog.Logger的包装
type LoggerWrapper struct {
	logger zerolog.Logger
}

// NewLoggerWrapper 创建一个新的LoggerWrapper
func NewLoggerWrapper(logger zerolog.Logger) *LoggerWrapper {
	return &LoggerWrapper{
		logger: logger,
	}
}

// Debug 创建Debug级别的日志，并添加调用位置信息
func (lw *LoggerWrapper) Debug() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Debug().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Debug().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Info 创建Info级别的日志，并添加调用位置信息
func (lw *LoggerWrapper) Info() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Info().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Info().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Warn 创建Warn级别的日志，并添加调用位置信息
func (lw *LoggerWrapper) Warn() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Warn().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Warn().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Error 创建Error级别的日志，并添加调用位置信息
func (lw *LoggerWrapper) Error() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Error().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Error().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Fatal 创建Fatal级别的日志，并添加调用位置信息
func (lw *LoggerWrapper) Fatal() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Fatal().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Fatal().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// example
// func example() {
// 	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
// 	log := NewLoggerWrapper(logger)
// 	log.Debug().Msg("debug message")
// 	log.Info().Msg("info message")
// }
