package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/mingrenya/AI-Waf/coraza-spoa/analyzer"
	mongodb "github.com/mingrenya/AI-Waf/pkg/database/mongo"
	"github.com/rs/zerolog"
)

// SimpleLogger 简单的日志实现
type SimpleLogger struct {
	logger zerolog.Logger
}

func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *SimpleLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func main() {
	// 解析命令行参数
	mongoURI := flag.String("mongo", "mongodb://localhost:27017", "MongoDB连接URI")
	database := flag.String("db", "waf", "数据库名称")
	flag.Parse()

	// 创建日志记录器 (输出到stderr,避免污染stdout的JSON-RPC消息)
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	simpleLogger := &SimpleLogger{logger: logger}

	logger.Info().Msg("AI-WAF MCP Server 启动中...")
	logger.Info().Str("mongo", *mongoURI).Str("database", *database).Msg("连接配置")

	// 连接MongoDB
	client, err := mongodb.Connect(*mongoURI)
	if err != nil {
		logger.Fatal().Err(err).Msg("连接MongoDB失败")
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			logger.Error().Err(err).Msg("断开MongoDB连接失败")
		}
	}()

	db := client.Database(*database)

	// 创建MCP服务器
	mcpServer := analyzer.NewMCPServer(db, logger, simpleLogger)

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info().Msg("收到退出信号,正在关闭...")
		cancel()
	}()

	// 启动MCP服务器 (阻塞运行)
	logger.Info().Msg("MCP Server 已启动,等待客户端连接...")
	if err := mcpServer.Run(ctx); err != nil && err != context.Canceled {
		logger.Fatal().Err(err).Msg("MCP Server运行失败")
	}

	logger.Info().Msg("MCP Server 已关闭")
}
