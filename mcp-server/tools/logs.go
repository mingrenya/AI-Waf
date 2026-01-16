// tools/logs.go
// WAF日志查询工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListAttackLogsInput 查询攻击日志的输入参数
type ListAttackLogsInput struct {
	Hours    int    `json:"hours" jsonschema:"查询最近N小时的日志,默认24"`
	Type     string `json:"type,omitempty" jsonschema:"攻击类型过滤,如sql_injection,xss,path_traversal"`
	Severity string `json:"severity,omitempty" jsonschema:"严重程度过滤,如critical,high,medium,low"`
	Limit    int    `json:"limit,omitempty" jsonschema:"返回结果数量限制,默认50"`
}

// ListAttackLogsOutput 攻击日志输出
type ListAttackLogsOutput struct {
	Count int           `json:"count" jsonschema:"日志总数"`
	Logs  []interface{} `json:"logs" jsonschema:"日志详细信息"`
}

// CreateListAttackLogs 创建查询攻击日志的工具函数
func CreateListAttackLogs(client *APIClient) func(context.Context, *mcp.CallToolRequest, ListAttackLogsInput) (*mcp.CallToolResult, ListAttackLogsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListAttackLogsInput) (*mcp.CallToolResult, ListAttackLogsOutput, error) {
		logger := NewToolLoggerWithClient("list_attack_logs", client)
		logger.LogInput(input)

		// 设置默认值
		if input.Hours == 0 {
			input.Hours = 24
		}
		if input.Limit == 0 {
			input.Limit = 50
		}

		// 构建查询参数 - 使用实际的API路径
		path := fmt.Sprintf("/api/v1/waf/logs?page=1&pageSize=%d", input.Limit)
		if input.Type != "" {
			path += "&attackType=" + input.Type
		}
		if input.Severity != "" {
			path += "&severity=" + input.Severity
		}

		// 调用后端API
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, ListAttackLogsOutput{}, fmt.Errorf("查询日志失败: %w", err)
		}

		// 解析响应
		var result struct {
			Data struct {
				List  []interface{} `json:"list"`
				Total int           `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, ListAttackLogsOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess(fmt.Sprintf("返回 %d 条日志", result.Data.Total))
		return nil, ListAttackLogsOutput{
			Count: result.Data.Total,
			Logs:  result.Data.List,
		}, nil
	}
}

// GetLogStatsInput 获取日志统计的输入参数
type GetLogStatsInput struct {
	Hours int `json:"hours" jsonschema:"统计最近N小时的数据,默认24"`
}

// GetLogStatsOutput 日志统计输出
type GetLogStatsOutput struct {
	TotalAttacks   int                    `json:"totalAttacks" jsonschema:"总攻击次数"`
	AttackTypes    map[string]int         `json:"attackTypes" jsonschema:"各类型攻击数量"`
	TopSourceIPs   []map[string]interface{} `json:"topSourceIPs" jsonschema:"攻击来源IP TOP10"`
	SeverityDistribution map[string]int   `json:"severityDistribution" jsonschema:"严重程度分布"`
}

// CreateGetLogStats 创建获取日志统计的工具函数
func CreateGetLogStats(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetLogStatsInput) (*mcp.CallToolResult, GetLogStatsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetLogStatsInput) (*mcp.CallToolResult, GetLogStatsOutput, error) {
		if input.Hours == 0 {
			input.Hours = 24
		}

		path := fmt.Sprintf("/api/waf-logs/stats?hours=%d", input.Hours)
		data, err := client.Get(path)
		if err != nil {
			return nil, GetLogStatsOutput{}, fmt.Errorf("获取统计失败: %w", err)
		}

		var result struct {
			Data struct {
				TotalAttacks         int                      `json:"totalAttacks"`
				AttackTypes          map[string]int           `json:"attackTypes"`
				TopSourceIPs         []map[string]interface{} `json:"topSourceIPs"`
				SeverityDistribution map[string]int           `json:"severityDistribution"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, GetLogStatsOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		return nil, GetLogStatsOutput{
			TotalAttacks:         result.Data.TotalAttacks,
			AttackTypes:          result.Data.AttackTypes,
			TopSourceIPs:         result.Data.TopSourceIPs,
			SeverityDistribution: result.Data.SeverityDistribution,
		}, nil
	}
}
