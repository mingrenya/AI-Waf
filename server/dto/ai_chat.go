package dto

import "time"

// AIChatRequest AI聊天请求
type AIChatRequest struct {
	Message  string          `json:"message" binding:"required"`
	Stream   bool            `json:"stream"`
	Messages []AIChatMessage `json:"messages"` // 历史消息上下文
}

// AIChatMessage AI聊天消息
type AIChatMessage struct {
	Role    string `json:"role" binding:"required,oneof=user assistant system"`
	Content string `json:"content" binding:"required"`
}

// AIChatResponse AI聊天响应
type AIChatResponse struct {
	Message   string    `json:"message"`
	ToolCalls []string  `json:"toolCalls,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// AIChatStreamResponse AI聊天流式响应
type AIChatStreamResponse struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}
