// tools/ai_analyzer.go
// AI分析器工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListAttackPatternsInput 列出攻击模式的输入参数
type ListAttackPatternsInput struct {
	Page     int    `json:"page,omitempty" jsonschema:"页码,默认1"`
	Size     int    `json:"size,omitempty" jsonschema:"每页数量,默认20"`
	Severity string `json:"severity,omitempty" jsonschema:"严重程度过滤"`
}

// ListAttackPatternsOutput 攻击模式列表输出
type ListAttackPatternsOutput struct {
	Total    int           `json:"total" jsonschema:"模式总数"`
	Patterns []interface{} `json:"patterns" jsonschema:"攻击模式列表"`
}

// CreateListAttackPatterns 创建列出攻击模式的工具函数
func CreateListAttackPatterns(client *APIClient) func(context.Context, *mcp.CallToolRequest, ListAttackPatternsInput) (*mcp.CallToolResult, ListAttackPatternsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListAttackPatternsInput) (*mcp.CallToolResult, ListAttackPatternsOutput, error) {
		if input.Page == 0 {
			input.Page = 1
		}
		if input.Size == 0 {
			input.Size = 20
		}

		path := fmt.Sprintf("/api/ai-analyzer/patterns?page=%d&size=%d", input.Page, input.Size)
		if input.Severity != "" {
			path += "&severity=" + input.Severity
		}

		data, err := client.Get(path)
		if err != nil {
			return nil, ListAttackPatternsOutput{}, fmt.Errorf("查询攻击模式失败: %w", err)
		}

		var result struct {
			Data struct {
				List  []interface{} `json:"list"`
				Total int           `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, ListAttackPatternsOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		return nil, ListAttackPatternsOutput{
			Total:    result.Data.Total,
			Patterns: result.Data.List,
		}, nil
	}
}

// ListGeneratedRulesInput 列出生成规则的输入参数
type ListGeneratedRulesInput struct {
	Page   int    `json:"page,omitempty" jsonschema:"页码,默认1"`
	Size   int    `json:"size,omitempty" jsonschema:"每页数量,默认20"`
	Status string `json:"status,omitempty" jsonschema:"状态过滤: pending,approved,rejected,deployed"`
}

// ListGeneratedRulesOutput 生成规则列表输出
type ListGeneratedRulesOutput struct {
	Total int           `json:"total" jsonschema:"规则总数"`
	Rules []interface{} `json:"rules" jsonschema:"规则列表"`
}

// CreateListGeneratedRules 创建列出生成规则的工具函数
func CreateListGeneratedRules(client *APIClient) func(context.Context, *mcp.CallToolRequest, ListGeneratedRulesInput) (*mcp.CallToolResult, ListGeneratedRulesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListGeneratedRulesInput) (*mcp.CallToolResult, ListGeneratedRulesOutput, error) {
		if input.Page == 0 {
			input.Page = 1
		}
		if input.Size == 0 {
			input.Size = 20
		}

		path := fmt.Sprintf("/api/ai-analyzer/rules?page=%d&size=%d", input.Page, input.Size)
		if input.Status != "" {
			path += "&status=" + input.Status
		}

		data, err := client.Get(path)
		if err != nil {
			return nil, ListGeneratedRulesOutput{}, fmt.Errorf("查询生成规则失败: %w", err)
		}

		var result struct {
			Data struct {
				List  []interface{} `json:"list"`
				Total int           `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, ListGeneratedRulesOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		return nil, ListGeneratedRulesOutput{
			Total: result.Data.Total,
			Rules: result.Data.List,
		}, nil
	}
}

// TriggerAIAnalysisInput 触发AI分析的输入参数
type TriggerAIAnalysisInput struct {
	Force bool `json:"force,omitempty" jsonschema:"是否强制立即分析"`
}

// TriggerAIAnalysisOutput AI分析输出
type TriggerAIAnalysisOutput struct {
	Message string `json:"message" jsonschema:"分析结果消息"`
}

// CreateTriggerAIAnalysis 创建触发AI分析的工具函数
func CreateTriggerAIAnalysis(client *APIClient) func(context.Context, *mcp.CallToolRequest, TriggerAIAnalysisInput) (*mcp.CallToolResult, TriggerAIAnalysisOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input TriggerAIAnalysisInput) (*mcp.CallToolResult, TriggerAIAnalysisOutput, error) {
		_, err := client.Post("/api/ai-analyzer/trigger", map[string]bool{"force": input.Force})
		if err != nil {
			return nil, TriggerAIAnalysisOutput{}, fmt.Errorf("触发分析失败: %w", err)
		}

		return nil, TriggerAIAnalysisOutput{
			Message: "AI分析任务已启动，请稍后查看结果",
		}, nil
	}
}

// ReviewRuleInput 审核规则的输入参数
type ReviewRuleInput struct {
	RuleID  string `json:"ruleId" jsonschema:"规则ID"`
	Action  string `json:"action" jsonschema:"审核动作: approve或reject"`
	Comment string `json:"comment,omitempty" jsonschema:"审核意见"`
}

// ReviewRuleOutput 审核规则输出
type ReviewRuleOutput struct {
	Message string `json:"message" jsonschema:"审核结果消息"`
}

// CreateReviewRule 创建审核规则的工具函数
func CreateReviewRule(client *APIClient) func(context.Context, *mcp.CallToolRequest, ReviewRuleInput) (*mcp.CallToolResult, ReviewRuleOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ReviewRuleInput) (*mcp.CallToolResult, ReviewRuleOutput, error) {
		if input.Action != "approve" && input.Action != "reject" {
			return nil, ReviewRuleOutput{}, fmt.Errorf("无效的审核动作，必须是approve或reject")
		}

		path := fmt.Sprintf("/api/ai-analyzer/rules/%s/review", input.RuleID)
		_, err := client.Post(path, map[string]string{
			"action":  input.Action,
			"comment": input.Comment,
		})
		if err != nil {
			return nil, ReviewRuleOutput{}, fmt.Errorf("审核规则失败: %w", err)
		}

		message := "规则已批准"
		if input.Action == "reject" {
			message = "规则已拒绝"
		}

		return nil, ReviewRuleOutput{
			Message: message,
		}, nil
	}
}

// DeployRuleInput 部署规则的输入参数
type DeployRuleInput struct {
	RuleID string `json:"ruleId" jsonschema:"要部署的规则ID"`
}

// DeployRuleOutput 部署规则输出
type DeployRuleOutput struct {
	Message string `json:"message" jsonschema:"部署结果消息"`
}

// CreateDeployRule 创建部署规则的工具函数
func CreateDeployRule(client *APIClient) func(context.Context, *mcp.CallToolRequest, DeployRuleInput) (*mcp.CallToolResult, DeployRuleOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeployRuleInput) (*mcp.CallToolResult, DeployRuleOutput, error) {
		path := fmt.Sprintf("/api/ai-analyzer/rules/%s/deploy", input.RuleID)
		_, err := client.Post(path, nil)
		if err != nil {
			return nil, DeployRuleOutput{}, fmt.Errorf("部署规则失败: %w", err)
		}

		return nil, DeployRuleOutput{
			Message: "规则已成功部署到生产环境",
		}, nil
	}
}
