// tools/rules.go
// MicroRule规则管理工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListMicroRulesInput 列出规则的输入参数
type ListMicroRulesInput struct {
	Page int `json:"page,omitempty" jsonschema:"页码,默认1"`
	Size int `json:"size,omitempty" jsonschema:"每页数量,默认20"`
}

// ListMicroRulesOutput 规则列表输出
type ListMicroRulesOutput struct {
	Total int           `json:"total" jsonschema:"规则总数"`
	Rules []interface{} `json:"rules" jsonschema:"规则列表"`
}

// CreateListMicroRules 创建列出规则的工具函数
func CreateListMicroRules(client *APIClient) func(context.Context, *mcp.CallToolRequest, ListMicroRulesInput) (*mcp.CallToolResult, ListMicroRulesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListMicroRulesInput) (*mcp.CallToolResult, ListMicroRulesOutput, error) {
		logger := NewToolLogger("list_micro_rules")
		logger.LogInput(input)
		
		if input.Page == 0 {
			input.Page = 1
		}
		if input.Size == 0 {
			input.Size = 20
		}

		// 使用实际的API路径 /api/v1/micro-rules
		path := fmt.Sprintf("/api/v1/micro-rules?page=%d&pageSize=%d", input.Page, input.Size)
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, ListMicroRulesOutput{}, fmt.Errorf("查询规则失败: %w", err)
		}

		var result struct {
			Data struct {
				List  []interface{} `json:"list"`
				Total int           `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, ListMicroRulesOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		return nil, ListMicroRulesOutput{
			Total: result.Data.Total,
			Rules: result.Data.List,
		}, nil
	}
}

// CreateMicroRuleInput 创建规则的输入参数
type CreateMicroRuleInput struct {
	Name        string      `json:"name" jsonschema:"规则名称"`
	Description string      `json:"description,omitempty" jsonschema:"规则描述"`
	Type        string      `json:"type" jsonschema:"规则类型: blacklist或whitelist"`
	Enabled     bool        `json:"enabled" jsonschema:"是否启用"`
	Priority    int         `json:"priority,omitempty" jsonschema:"优先级,数字越大优先级越高"`
	Conditions  interface{} `json:"conditions" jsonschema:"规则条件,JSON格式"`
}

// CreateMicroRuleOutput 创建规则的输出
type CreateMicroRuleOutput struct {
	RuleID  string `json:"ruleId" jsonschema:"创建的规则ID"`
	Message string `json:"message" jsonschema:"创建结果消息"`
}

// CreateCreateMicroRule 创建新规则的工具函数
func CreateCreateMicroRule(client *APIClient) func(context.Context, *mcp.CallToolRequest, CreateMicroRuleInput) (*mcp.CallToolResult, CreateMicroRuleOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateMicroRuleInput) (*mcp.CallToolResult, CreateMicroRuleOutput, error) {
		data, err := client.Post("/api/rules/micro-rule", input)
		if err != nil {
			return nil, CreateMicroRuleOutput{}, fmt.Errorf("创建规则失败: %w", err)
		}

		var result struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, CreateMicroRuleOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		return nil, CreateMicroRuleOutput{
			RuleID:  result.Data.ID,
			Message: "规则创建成功",
		}, nil
	}
}

// UpdateMicroRuleInput 更新规则的输入参数
type UpdateMicroRuleInput struct {
	RuleID      string      `json:"ruleId" jsonschema:"要更新的规则ID"`
	Name        string      `json:"name,omitempty" jsonschema:"规则名称"`
	Description string      `json:"description,omitempty" jsonschema:"规则描述"`
	Enabled     *bool       `json:"enabled,omitempty" jsonschema:"是否启用"`
	Priority    *int        `json:"priority,omitempty" jsonschema:"优先级"`
	Conditions  interface{} `json:"conditions,omitempty" jsonschema:"规则条件"`
}

// UpdateMicroRuleOutput 更新规则的输出
type UpdateMicroRuleOutput struct {
	Message string `json:"message" jsonschema:"更新结果消息"`
}

// CreateUpdateMicroRule 创建更新规则的工具函数
func CreateUpdateMicroRule(client *APIClient) func(context.Context, *mcp.CallToolRequest, UpdateMicroRuleInput) (*mcp.CallToolResult, UpdateMicroRuleOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateMicroRuleInput) (*mcp.CallToolResult, UpdateMicroRuleOutput, error) {
		path := fmt.Sprintf("/api/rules/micro-rule/%s", input.RuleID)
		_, err := client.Put(path, input)
		if err != nil {
			return nil, UpdateMicroRuleOutput{}, fmt.Errorf("更新规则失败: %w", err)
		}

		return nil, UpdateMicroRuleOutput{
			Message: "规则更新成功",
		}, nil
	}
}

// DeleteMicroRuleInput 删除规则的输入参数
type DeleteMicroRuleInput struct {
	RuleID string `json:"ruleId" jsonschema:"要删除的规则ID"`
}

// DeleteMicroRuleOutput 删除规则的输出
type DeleteMicroRuleOutput struct {
	Message string `json:"message" jsonschema:"删除结果消息"`
}

// CreateDeleteMicroRule 创建删除规则的工具函数
func CreateDeleteMicroRule(client *APIClient) func(context.Context, *mcp.CallToolRequest, DeleteMicroRuleInput) (*mcp.CallToolResult, DeleteMicroRuleOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteMicroRuleInput) (*mcp.CallToolResult, DeleteMicroRuleOutput, error) {
		path := fmt.Sprintf("/api/rules/micro-rule/%s", input.RuleID)
		err := client.Delete(path)
		if err != nil {
			return nil, DeleteMicroRuleOutput{}, fmt.Errorf("删除规则失败: %w", err)
		}

		return nil, DeleteMicroRuleOutput{
			Message: "规则删除成功",
		}, nil
	}
}
