// tools/sites.go
// 站点管理工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListSitesInput 列出站点的输入参数
type ListSitesInput struct{}

// ListSitesOutput 站点列表输出
type ListSitesOutput struct {
	Total int           `json:"total" jsonschema:"站点总数"`
	Sites []interface{} `json:"sites" jsonschema:"站点列表"`
}

// CreateListSites 创建列出站点的工具函数
func CreateListSites(client *APIClient) func(context.Context, *mcp.CallToolRequest, ListSitesInput) (*mcp.CallToolResult, ListSitesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListSitesInput) (*mcp.CallToolResult, ListSitesOutput, error) {
		logger := NewToolLogger("list_sites")
		
		// 使用实际的API路径 /api/v1/site
		data, err := client.Get("/api/v1/site")
		if err != nil {
			logger.LogError(err)
			return nil, ListSitesOutput{}, fmt.Errorf("查询站点失败: %w", err)
		}

		var result struct {
			Data []interface{} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, ListSitesOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess(fmt.Sprintf("返回 %d 个站点", len(result.Data)))
		return nil, ListSitesOutput{
			Total: len(result.Data),
			Sites: result.Data,
		}, nil
	}
}

// GetSiteDetailsInput 获取站点详情的输入参数
type GetSiteDetailsInput struct {
	SiteID string `json:"siteId" jsonschema:"站点ID"`
}

// GetSiteDetailsOutput 站点详情输出
type GetSiteDetailsOutput struct {
	Site interface{} `json:"site" jsonschema:"站点详细信息"`
}

// CreateGetSiteDetails 创建获取站点详情的工具函数
func CreateGetSiteDetails(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetSiteDetailsInput) (*mcp.CallToolResult, GetSiteDetailsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetSiteDetailsInput) (*mcp.CallToolResult, GetSiteDetailsOutput, error) {
		logger := NewToolLogger("get_site_details")
		logger.LogInput(input)
		
		// 使用实际的API路径 /api/v1/site/{id}
		path := fmt.Sprintf("/api/v1/site/%s", input.SiteID)
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, GetSiteDetailsOutput{}, fmt.Errorf("查询站点详情失败: %w", err)
		}

		var result struct {
			Data interface{} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, GetSiteDetailsOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess("获取站点详情成功")
		return nil, GetSiteDetailsOutput{
			Site: result.Data,
		}, nil
	}
}
