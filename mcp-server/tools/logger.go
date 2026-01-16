// tools/logger.go
// MCP 工具调用日志记录器
package tools

import (
	"encoding/json"
	"log"
	"time"
)

// ToolLogger 工具调用日志记录器
type ToolLogger struct {
	ToolName  string
	StartTime time.Time
	Client    *APIClient // 添加 API 客户端，用于记录到后端
}

// NewToolLogger 创建工具日志记录器（不记录到后端）
// 保留此方法以兼容现有代码
func NewToolLogger(toolName string) *ToolLogger {
	logger := &ToolLogger{
		ToolName:  toolName,
		StartTime: time.Now(),
		Client:    nil, // 不记录到后端
	}
	log.Printf("[工具调用] %s 开始执行", toolName)
	return logger
}

// NewToolLoggerWithClient 创建带客户端的工具日志记录器（会记录到后端）
func NewToolLoggerWithClient(toolName string, client *APIClient) *ToolLogger {
	logger := &ToolLogger{
		ToolName:  toolName,
		StartTime: time.Now(),
		Client:    client,
	}
	log.Printf("[工具调用] %s 开始执行", toolName)
	return logger
}

// LogInput 记录输入参数
func (l *ToolLogger) LogInput(input interface{}) {
	if jsonData, err := json.Marshal(input); err == nil {
		log.Printf("[工具参数] %s - 输入: %s", l.ToolName, string(jsonData))
	}
}

// LogSuccess 记录成功结果
func (l *ToolLogger) LogSuccess(resultSummary string) {
	duration := time.Since(l.StartTime)
	log.Printf("[工具成功] %s - %s - 耗时: %v", l.ToolName, resultSummary, duration)
	
	// 记录到后端数据库
	if l.Client != nil {
		l.recordToBackend(true, "")
	}
}

// LogError 记录错误
func (l *ToolLogger) LogError(err error) {
	duration := time.Since(l.StartTime)
	log.Printf("[工具错误] %s - %v - 耗时: %v", l.ToolName, err, duration)
	
	// 记录到后端数据库
	if l.Client != nil {
		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
		}
		l.recordToBackend(false, errorMsg)
	}
}

// recordToBackend 记录工具调用到后端数据库
func (l *ToolLogger) recordToBackend(success bool, errorMsg string) {
	duration := time.Since(l.StartTime).Milliseconds()
	
	data := map[string]interface{}{
		"toolName": l.ToolName,
		"duration": duration,
		"success":  success,
		"error":    errorMsg,
	}
	
	// 异步调用，避免阻塞工具执行
	go func() {
		_, err := l.Client.Post("/api/v1/mcp/tool-calls/record", data)
		if err != nil {
			log.Printf("[记录失败] 无法记录工具调用 %s: %v", l.ToolName, err)
		}
	}()
}

// LogWarning 记录警告
func (l *ToolLogger) LogWarning(message string) {
	log.Printf("[工具警告] %s - %s", l.ToolName, message)
}
