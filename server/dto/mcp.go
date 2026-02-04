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
	ID                string    `json:"id"`
	ToolName          string    `json:"toolName"`
	Timestamp         time.Time `json:"timestamp"`
	Duration          int64     `json:"duration"` // milliseconds
	Success           bool      `json:"success"`
	Error             string    `json:"error,omitempty"`
	ParentMessageUUID string    `json:"parentMessageUUID,omitempty"`
}

// MCPToolCallHistoryResponse 工具调用历史响应
type MCPToolCallHistoryResponse struct {
	Data  []MCPToolCallRecord `json:"data"`
	Total int64               `json:"total"`
}

// RecordToolCallRequest 记录工具调用请求
type RecordToolCallRequest struct {
	ToolName          string `json:"toolName" binding:"required"`
	Duration          int64  `json:"duration" binding:"required,min=0"` // milliseconds
	Success           bool   `json:"success"`
	Error             string `json:"error,omitempty"`
	ParentMessageUUID string `json:"parent_message_uuid,omitempty"`
}

// AIRuleSuggestion AI生成的规则建议
// 用于前端 MCP 规则建议接口
type AIRuleSuggestion struct {
	ID             string                 `json:"id"`
	PatternID      string                 `json:"patternId,omitempty"`
	PatternName    string                 `json:"patternName"`
	RuleName       string                 `json:"ruleName"`
	RuleType       string                 `json:"ruleType"`
	Confidence     float64                `json:"confidence"`
	Severity       string                 `json:"severity"`
	Description    string                 `json:"description"`
	Recommendation string                 `json:"recommendation"`
	RuleContent    map[string]interface{} `json:"ruleContent"`
	Status         string                 `json:"status"`
	CreatedAt      time.Time              `json:"createdAt"`
	ReviewedAt     *time.Time             `json:"reviewedAt,omitempty"`
	DeployedAt     *time.Time             `json:"deployedAt,omitempty"`
}

// AIRuleSuggestionListResponse 规则建议列表响应
// 对齐前端: { data: AIRuleSuggestion[], total: number }
type AIRuleSuggestionListResponse struct {
	Data  []AIRuleSuggestion `json:"data"`
	Total int64              `json:"total"`
}

// AIAnalysisResult AI分析结果
// 对齐前端: totalPatterns, highSeverityPatterns, suggestedRules, processingTime, timestamp
type AIAnalysisResult struct {
	TotalPatterns        int64     `json:"totalPatterns"`
	HighSeverityPatterns int64     `json:"highSeverityPatterns"`
	SuggestedRules       int64     `json:"suggestedRules"`
	ProcessingTime       float64   `json:"processingTime"`
	Timestamp            time.Time `json:"timestamp"`
}

// TriggerAnalysisParams MCP触发分析请求
// 对齐前端 /ai-analyzer/analyze/patterns
// 当前实现仅用于接收参数，后端暂不依赖该参数进行计算
type TriggerAnalysisParams struct {
	TimeRange        string  `json:"timeRange"`
	MinSamples       int     `json:"minSamples"`
	AnomalyThreshold float64 `json:"anomalyThreshold"`
	ClusteringMethod string  `json:"clusteringMethod"`
}
