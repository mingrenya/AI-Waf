package dto

import "time"

// MCPStatusResponse MCP连接状态响应
type MCPStatusResponse struct {
	Connected       bool     `json:"connected"`
	LastConnectedAt *string  `json:"lastConnectedAt,omitempty"`
	ServerVersion   string   `json:"serverVersion,omitempty"`
	TotalTools      int      `json:"totalTools"`
	AvailableTools  []string `json:"availableTools"`
	Error           string   `json:"error,omitempty"`
}

// MCPToolsResponse MCP工具列表响应
type MCPToolsResponse struct {
	Tools []string `json:"tools"`
}

// MCPToolCallHistoryRequest 工具调用历史查询请求
type MCPToolCallHistoryRequest struct {
	Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int `form:"offset" binding:"omitempty,min=0"`
}

// MCPToolCallRecord 工具调用记录
type MCPToolCallRecord struct {
	ID        string    `json:"id"`
	ToolName  string    `json:"toolName"`
	Timestamp time.Time `json:"timestamp"`
	Duration  int64     `json:"duration"` // milliseconds
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// MCPToolCallHistoryResponse 工具调用历史响应
type MCPToolCallHistoryResponse struct {
	Data  []MCPToolCallRecord `json:"data"`
	Total int64               `json:"total"`
}

// RecordToolCallRequest 记录工具调用请求
type RecordToolCallRequest struct {
	ToolName string `json:"toolName" binding:"required"`
	Duration int64  `json:"duration" binding:"required,min=0"` // milliseconds
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}
