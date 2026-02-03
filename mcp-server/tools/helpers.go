// tools/helpers.go
// 公共辅助函数
package tools

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// PaginationParams 分页参数
type PaginationParams struct {
	Page int
	Size int
}

// ValidatePagination 验证并规范化分页参数
func ValidatePagination(page, size int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return PaginationParams{Page: page, Size: size}
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	List  []interface{} `json:"list"`
	Total int           `json:"total"`
}

// GetPaginatedList 获取分页列表的通用函数
func GetPaginatedList(client *APIClient, path string) ([]interface{}, int, error) {
	data, err := client.Get(path)
	if err != nil {
		return nil, 0, WrapError(err, "API请求")
	}

	var result struct {
		Data PaginatedResponse `json:"data"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, 0, FormatParseError("响应", err)
	}

	return result.Data.List, result.Data.Total, nil
}

// ValidateLimit 验证限制参数（如QPS数据点数量）
func ValidateLimit(limit, defaultVal, maxVal int) int {
	if limit <= 0 {
		return defaultVal
	}
	if limit > maxVal {
		return maxVal
	}
	return limit
}

// StandardAPIResponse 标准API响应结构
type StandardAPIResponse struct {
	Data interface{} `json:"data"`
}

// ParseAPIResponse 解析标准API响应
func ParseAPIResponse(data []byte, target interface{}) error {
	var result struct {
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return FormatParseError("外层响应", err)
	}

	if err := json.Unmarshal(result.Data, target); err != nil {
		return FormatParseError("数据字段", err)
	}

	return nil
}

// URLBuilder API URL构建器
type URLBuilder struct {
	path   string
	params url.Values
}

// NewURLBuilder 创建新的URL构建器
// 示例: NewURLBuilder("/api/v1/logs").AddParam("page", 1).AddParam("size", 20).Build()
func NewURLBuilder(path string) *URLBuilder {
	return &URLBuilder{
		path:   path,
		params: url.Values{},
	}
}

// AddParam 添加查询参数
func (b *URLBuilder) AddParam(key string, value interface{}) *URLBuilder {
	if value == nil {
		return b
	}

	switch v := value.(type) {
	case string:
		if v != "" {
			b.params.Add(key, v)
		}
	case int:
		if v != 0 {
			b.params.Add(key, strconv.Itoa(v))
		}
	case bool:
		b.params.Add(key, strconv.FormatBool(v))
	default:
		b.params.Add(key, fmt.Sprintf("%v", v))
	}
	return b
}

// AddParamIfNotEmpty 只在值非空时添加参数
func (b *URLBuilder) AddParamIfNotEmpty(key string, value string) *URLBuilder {
	if value != "" {
		b.params.Add(key, value)
	}
	return b
}

// Build 构建最终的URL路径（带查询参数）
func (b *URLBuilder) Build() string {
	if len(b.params) == 0 {
		return b.path
	}
	return b.path + "?" + b.params.Encode()
}

// Validator 输入验证接口
type Validator interface {
	Validate() error
}

// ValidateInput 执行输入验证（如果实现了Validator接口）
func ValidateInput(input interface{}) error {
	if validator, ok := input.(Validator); ok {
		return validator.Validate()
	}
	return nil
}

// ValidationError 验证错误类型
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidationError 创建验证错误
func NewValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
