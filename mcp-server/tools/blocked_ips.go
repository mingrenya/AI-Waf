// tools/blocked_ips.go
// IP封禁管理工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListBlockedIPsInput 列出封禁IP的输入参数
type ListBlockedIPsInput struct {
	Page int `json:"page,omitempty" jsonschema:"页码,默认1"`
	Size int `json:"size,omitempty" jsonschema:"每页数量,默认20"`
}

// ListBlockedIPsOutput 封禁IP列表输出
type ListBlockedIPsOutput struct {
	Total int           `json:"total" jsonschema:"封禁IP总数"`
	IPs   []interface{} `json:"ips" jsonschema:"封禁IP详细信息"`
}

// CreateListBlockedIPs 创建列出封禁IP的工具函数
func CreateListBlockedIPs(client *APIClient) func(context.Context, *mcp.CallToolRequest, ListBlockedIPsInput) (*mcp.CallToolResult, ListBlockedIPsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListBlockedIPsInput) (*mcp.CallToolResult, ListBlockedIPsOutput, error) {
		logger := NewToolLogger("list_blocked_ips")
		logger.LogInput(input)
		
		if input.Page == 0 {
			input.Page = 1
		}
		if input.Size == 0 {
			input.Size = 20
		}

		// 使用实际的API路径 /api/v1/blocked-ips
		path := fmt.Sprintf("/api/v1/blocked-ips?page=%d&pageSize=%d", input.Page, input.Size)
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, ListBlockedIPsOutput{}, fmt.Errorf("查询封禁IP失败: %w", err)
		}

		var result struct {
			Data struct {
				List  []interface{} `json:"list"`
				Total int           `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, ListBlockedIPsOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		return nil, ListBlockedIPsOutput{
			Total: result.Data.Total,
			IPs:   result.Data.List,
		}, nil
	}
}

// GetBlockedIPStatsInput 获取封禁IP统计的输入参数
type GetBlockedIPStatsInput struct{}

// GetBlockedIPStatsOutput 封禁IP统计输出
type GetBlockedIPStatsOutput struct {
	TotalBlocked int            `json:"totalBlocked" jsonschema:"总封禁IP数"`
	ActiveBlocks int            `json:"activeBlocks" jsonschema:"当前活跃封禁数"`
	ReasonStats  map[string]int `json:"reasonStats" jsonschema:"封禁原因统计"`
}

// CreateGetBlockedIPStats 创建获取封禁IP统计的工具函数
func CreateGetBlockedIPStats(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetBlockedIPStatsInput) (*mcp.CallToolResult, GetBlockedIPStatsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetBlockedIPStatsInput) (*mcp.CallToolResult, GetBlockedIPStatsOutput, error) {
		data, err := client.Get("/api/flow-control/blocked-ips/stats")
		if err != nil {
			return nil, GetBlockedIPStatsOutput{}, fmt.Errorf("获取统计失败: %w", err)
		}

		var result struct {
			Data struct {
				TotalBlocked int            `json:"totalBlocked"`
				ActiveBlocks int            `json:"activeBlocks"`
				ReasonStats  map[string]int `json:"reasonStats"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, GetBlockedIPStatsOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		return nil, GetBlockedIPStatsOutput{
			TotalBlocked: result.Data.TotalBlocked,
			ActiveBlocks: result.Data.ActiveBlocks,
			ReasonStats:  result.Data.ReasonStats,
		}, nil
	}
}
