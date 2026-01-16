// tools/config.go
// WAF配置管理工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetWAFConfigInput 获取WAF配置的输入参数
type GetWAFConfigInput struct{}

// GetWAFConfigOutput WAF配置输出
type GetWAFConfigOutput struct {
	Config interface{} `json:"config" jsonschema:"WAF配置详细信息"`
}

// CreateGetWAFConfig 创建获取WAF配置的工具函数
func CreateGetWAFConfig(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetWAFConfigInput) (*mcp.CallToolResult, GetWAFConfigOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetWAFConfigInput) (*mcp.CallToolResult, GetWAFConfigOutput, error) {
		logger := NewToolLogger("get_waf_config")
		
		// 使用实际的API路径 /api/v1/config
		data, err := client.Get("/api/v1/config")
		if err != nil {
			logger.LogError(err)
			return nil, GetWAFConfigOutput{}, fmt.Errorf("获取配置失败: %w", err)
		}

		var result struct {
			Data interface{} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, GetWAFConfigOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess("获取WAF配置成功")
		return nil, GetWAFConfigOutput{
			Config: result.Data,
		}, nil
	}
}

// UpdateWAFConfigInput 更新WAF配置的输入参数
type UpdateWAFConfigInput struct {
	LogLevel            *string `json:"logLevel,omitempty" jsonschema:"日志级别：debug,info,warn,error"`
	MaxRequestBodySize  *int    `json:"maxRequestBodySize,omitempty" jsonschema:"最大请求体大小(字节)"`
	EnableGeoIP         *bool   `json:"enableGeoIP,omitempty" jsonschema:"是否启用GeoIP"`
	EnableRateLimit     *bool   `json:"enableRateLimit,omitempty" jsonschema:"是否启用速率限制"`
	RateLimitRPM        *int    `json:"rateLimitRPM,omitempty" jsonschema:"速率限制(每分钟请求数)"`
	BlockDuration       *int    `json:"blockDuration,omitempty" jsonschema:"封禁时长(秒)"`
	EnableAIAnalysis    *bool   `json:"enableAIAnalysis,omitempty" jsonschema:"是否启用AI分析"`
}

// UpdateWAFConfigOutput 更新配置输出
type UpdateWAFConfigOutput struct {
	UpdatedConfig interface{} `json:"updatedConfig" jsonschema:"更新后的配置"`
	Message       string      `json:"message" jsonschema:"更新结果消息"`
}

// CreateUpdateWAFConfig 创建更新WAF配置的工具函数
func CreateUpdateWAFConfig(client *APIClient) func(context.Context, *mcp.CallToolRequest, UpdateWAFConfigInput) (*mcp.CallToolResult, UpdateWAFConfigOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateWAFConfigInput) (*mcp.CallToolResult, UpdateWAFConfigOutput, error) {
		logger := NewToolLogger("update_waf_config")
		logger.LogInput(input)
		
		// 使用PATCH方法更新配置 /api/v1/config
		data, err := client.Patch("/api/v1/config", input)
		if err != nil {
			logger.LogError(err)
			return nil, UpdateWAFConfigOutput{}, fmt.Errorf("更新配置失败: %w", err)
		}

		var result struct {
			Data interface{} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, UpdateWAFConfigOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess("更新WAF配置成功")
		return nil, UpdateWAFConfigOutput{
			UpdatedConfig: result.Data,
			Message:       "配置更新成功",
		}, nil
	}
}

// GetStatsOverviewInput 获取统计概览的输入参数
type GetStatsOverviewInput struct {
	TimeRange string `json:"timeRange" jsonschema:"时间范围：24h,7d,30d,默认24h"`
}

// GetStatsOverviewOutput 统计概览输出
type GetStatsOverviewOutput struct {
	TotalRequests   int64                  `json:"totalRequests" jsonschema:"总请求数"`
	BlockedRequests int64                  `json:"blockedRequests" jsonschema:"被阻止的请求数"`
	ErrorRate       float64                `json:"errorRate" jsonschema:"错误率"`
	AvgResponseTime float64                `json:"avgResponseTime" jsonschema:"平均响应时间(ms)"`
	Traffic         map[string]interface{} `json:"traffic" jsonschema:"流量统计"`
}

// CreateGetStatsOverview 创建获取统计概览的工具函数
func CreateGetStatsOverview(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetStatsOverviewInput) (*mcp.CallToolResult, GetStatsOverviewOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetStatsOverviewInput) (*mcp.CallToolResult, GetStatsOverviewOutput, error) {
		logger := NewToolLogger("get_stats_overview")
		logger.LogInput(input)
		
		if input.TimeRange == "" {
			input.TimeRange = "24h"
		}
		
		// 使用实际的API路径 /api/v1/stats/overview
		path := fmt.Sprintf("/api/v1/stats/overview?timeRange=%s", input.TimeRange)
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, GetStatsOverviewOutput{}, fmt.Errorf("获取统计概览失败: %w", err)
		}

		var result struct {
			Data GetStatsOverviewOutput `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, GetStatsOverviewOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess(fmt.Sprintf("获取%s统计概览成功", input.TimeRange))
		return nil, result.Data, nil
	}
}
