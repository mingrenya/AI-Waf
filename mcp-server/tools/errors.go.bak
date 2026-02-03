// tools/errors.go
// Enhanced Error Handling for MCP Server
// 为 MCP 工具提供友好的、可操作的错误消息
package tools

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrorType 错误类型常量
type ErrorType string

const (
	// 认证错误
	ErrorTypeAuth ErrorType = "authentication_error"
	// 未找到资源
	ErrorTypeNotFound ErrorType = "not_found"
	// 验证错误
	ErrorTypeValidation ErrorType = "validation_error"
	// 网络错误
	ErrorTypeNetwork ErrorType = "network_error"
	// 权限错误
	ErrorTypePermission ErrorType = "permission_error"
	// 速率限制
	ErrorTypeRateLimit ErrorType = "rate_limit"
	// 服务器错误
	ErrorTypeServer ErrorType = "server_error"
	// 超时错误
	ErrorTypeTimeout ErrorType = "timeout_error"
)

// MCPError MCP 增强错误类型
type MCPError struct {
	Type        ErrorType `json:"type"`
	Message     string    `json:"message"`
	Suggestion  string    `json:"suggestion"`
	HTTPStatus  int       `json:"http_status,omitempty"`
	OriginalErr error     `json:"-"`
}

func (e *MCPError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s. 建议: %s", e.Message, e.Suggestion)
	}
	return e.Message
}

func (e *MCPError) Unwrap() error {
	return e.OriginalErr
}

// NewAuthError 创建认证错误
func NewAuthError(message string, err error) error {
	return &MCPError{
		Type:        ErrorTypeAuth,
		Message:     message,
		Suggestion:  "请检查 Authorization token 是否正确配置。可以使用 `scripts/get-mcp-token.sh` 获取新 token。",
		HTTPStatus:  http.StatusUnauthorized,
		OriginalErr: err,
	}
}

// NewNotFoundError 创建资源未找到错误
func NewNotFoundError(resource string, identifier string, err error) error {
	return &MCPError{
		Type:        ErrorTypeNotFound,
		Message:     fmt.Sprintf("未找到%s: %s", resource, identifier),
		Suggestion:  fmt.Sprintf("请确认%s ID 是否正确，或使用列表工具查看所有可用的%s。", resource, resource),
		HTTPStatus:  http.StatusNotFound,
		OriginalErr: err,
	}
}

// NewValidationError 创建验证错误（增强版）
func NewValidationErrorWithSuggestion(field, message, suggestion string) error {
	return &MCPError{
		Type:       ErrorTypeValidation,
		Message:    fmt.Sprintf("%s: %s", field, message),
		Suggestion: suggestion,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(operation string, err error) error {
	suggestion := "请检查网络连接状态和 WAF 后端服务是否正常运行"
	if isTimeoutError(err) {
		suggestion = "请求超时。请检查网络延迟或考虑增加超时时间。"
	}
	return &MCPError{
		Type:        ErrorTypeNetwork,
		Message:     fmt.Sprintf("%s时发生网络错误: %v", operation, err),
		Suggestion:  suggestion,
		OriginalErr: err,
	}
}

// NewPermissionError 创建权限错误
func NewPermissionError(action string, err error) error {
	return &MCPError{
		Type:        ErrorTypePermission,
		Message:     fmt.Sprintf("无权限执行操作: %s", action),
		Suggestion:  "请检查当前用户是否有足够的权限。可能需要管理员权限。",
		HTTPStatus:  http.StatusForbidden,
		OriginalErr: err,
	}
}

// NewRateLimitError 创建速率限制错误
func NewRateLimitError(retryAfter string, err error) error {
	suggestion := "请求过于频繁，请稍后再试"
	if retryAfter != "" {
		suggestion = fmt.Sprintf("请在 %s 秒后重试", retryAfter)
	}
	return &MCPError{
		Type:        ErrorTypeRateLimit,
		Message:     "已超过 API 速率限制",
		Suggestion:  suggestion,
		HTTPStatus:  http.StatusTooManyRequests,
		OriginalErr: err,
	}
}

// NewServerError 创建服务器错误
func NewServerError(operation string, err error) error {
	return &MCPError{
		Type:        ErrorTypeServer,
		Message:     fmt.Sprintf("%s失败: 服务器内部错误", operation),
		Suggestion:  "这是服务器端问题。请检查 WAF 后端服务日志，或联系管理员。",
		HTTPStatus:  http.StatusInternalServerError,
		OriginalErr: err,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(operation string, err error) error {
	return &MCPError{
		Type:        ErrorTypeTimeout,
		Message:     fmt.Sprintf("%s超时", operation),
		Suggestion:  "操作超时。请检查网络连接，或对于大量数据操作，考虑分批处理。",
		OriginalErr: err,
	}
}

// FormatAPIError 根据 HTTP 状态码格式化 API 错误
func FormatAPIError(operation string, statusCode int, responseBody []byte, originalErr error) error {
	switch {
	case statusCode == http.StatusUnauthorized:
		return NewAuthError(
			fmt.Sprintf("%s失败: 认证失败", operation),
			originalErr,
		)
	case statusCode == http.StatusForbidden:
		return NewPermissionError(operation, originalErr)
	case statusCode == http.StatusNotFound:
		return &MCPError{
			Type:        ErrorTypeNotFound,
			Message:     fmt.Sprintf("%s失败: 资源不存在", operation),
			Suggestion:  "请检查请求的资源 ID 或路径是否正确。",
			HTTPStatus:  statusCode,
			OriginalErr: originalErr,
		}
	case statusCode == http.StatusTooManyRequests:
		return NewRateLimitError("", originalErr)
	case statusCode >= 500:
		return NewServerError(operation, originalErr)
	case statusCode >= 400:
		// 尝试解析错误响应
		errorMsg := string(responseBody)
		if errorMsg == "" {
			errorMsg = fmt.Sprintf("HTTP %d", statusCode)
		}
		return &MCPError{
			Type:        ErrorTypeValidation,
			Message:     fmt.Sprintf("%s失败: %s", operation, errorMsg),
			Suggestion:  "请检查请求参数是否符合要求。可以查看工具的 JSON Schema 了解详细约束。",
			HTTPStatus:  statusCode,
			OriginalErr: originalErr,
		}
	default:
		return originalErr
	}
}

// FormatParseError 格式化 JSON 解析错误
func FormatParseError(dataType string, err error) error {
	return &MCPError{
		Type:        ErrorTypeServer,
		Message:     fmt.Sprintf("解析%s数据失败", dataType),
		Suggestion:  "服务器返回的数据格式不正确。这可能是后端 API 版本不匹配。请检查 MCP Server 和 WAF 后端的版本兼容性。",
		OriginalErr: err,
	}
}

// isTimeoutError 检查是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	// 检查 context.DeadlineExceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// 检查错误消息中是否包含 timeout
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline exceeded") ||
		strings.Contains(errMsg, "i/o timeout")
}

// WrapError 包装错误并添加上下文
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}

	// 如果已经是 MCPError，保留类型和建议
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return &MCPError{
			Type:        mcpErr.Type,
			Message:     fmt.Sprintf("%s: %s", context, mcpErr.Message),
			Suggestion:  mcpErr.Suggestion,
			HTTPStatus:  mcpErr.HTTPStatus,
			OriginalErr: mcpErr.OriginalErr,
		}
	}

	// 根据错误类型自动分类
	if isTimeoutError(err) {
		return NewTimeoutError(context, err)
	}

	// 默认返回服务器错误
	return &MCPError{
		Type:        ErrorTypeServer,
		Message:     fmt.Sprintf("%s: %v", context, err),
		Suggestion:  "请查看错误详情并重试。如问题持续，请联系管理员。",
		OriginalErr: err,
	}
}

// ExtractSuggestion 从 MCPError 中提取建议
func ExtractSuggestion(err error) string {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Suggestion
	}
	return ""
}
