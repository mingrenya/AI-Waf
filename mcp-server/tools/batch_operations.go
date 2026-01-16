// tools/batch_operations.go
// 批量操作工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BatchBlockIPsInput 批量封禁IP的输入参数
type BatchBlockIPsInput struct {
	IPs      []string `json:"ips" jsonschema:"要封禁的IP地址列表"`
	Reason   string   `json:"reason" jsonschema:"封禁原因"`
	Duration int      `json:"duration,omitempty" jsonschema:"封禁时长(秒),0表示永久"`
}

// BatchBlockIPsOutput 批量封禁IP输出
type BatchBlockIPsOutput struct {
	SuccessCount int      `json:"successCount" jsonschema:"成功封禁的IP数量"`
	FailedCount  int      `json:"failedCount" jsonschema:"失败的IP数量"`
	FailedIPs    []string `json:"failedIPs,omitempty" jsonschema:"失败的IP列表"`
	Message      string   `json:"message" jsonschema:"操作结果消息"`
}

// CreateBatchBlockIPs 创建批量封禁IP的工具函数
func CreateBatchBlockIPs(client *APIClient) func(context.Context, *mcp.CallToolRequest, BatchBlockIPsInput) (*mcp.CallToolResult, BatchBlockIPsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchBlockIPsInput) (*mcp.CallToolResult, BatchBlockIPsOutput, error) {
		logger := NewToolLogger("batch_block_ips")
		logger.LogInput(input)
		
		if len(input.IPs) == 0 {
			logger.LogWarning("IP列表为空")
			return nil, BatchBlockIPsOutput{
				Message: "IP列表不能为空",
			}, fmt.Errorf("IP列表为空")
		}

		successCount := 0
		failedCount := 0
		var failedIPs []string

		// 逐个封禁IP（因为后端可能没有批量接口，我们在MCP层面实现批量）
		for _, ip := range input.IPs {
			blockData := map[string]interface{}{
				"ip":       ip,
				"reason":   input.Reason,
				"duration": input.Duration,
			}
			
			_, err := client.Post("/api/v1/blocked-ips", blockData)
			if err != nil {
				failedCount++
				failedIPs = append(failedIPs, ip)
				logger.LogWarning(fmt.Sprintf("封禁IP %s 失败: %v", ip, err))
			} else {
				successCount++
			}
		}

		message := fmt.Sprintf("批量封禁完成: 成功 %d 个, 失败 %d 个", successCount, failedCount)
		logger.LogSuccess(message)
		
		return nil, BatchBlockIPsOutput{
			SuccessCount: successCount,
			FailedCount:  failedCount,
			FailedIPs:    failedIPs,
			Message:      message,
		}, nil
	}
}

// BatchUnblockIPsInput 批量解封IP的输入参数
type BatchUnblockIPsInput struct {
	IPs []string `json:"ips" jsonschema:"要解封的IP地址列表"`
}

// BatchUnblockIPsOutput 批量解封IP输出
type BatchUnblockIPsOutput struct {
	SuccessCount int      `json:"successCount" jsonschema:"成功解封的IP数量"`
	FailedCount  int      `json:"failedCount" jsonschema:"失败的IP数量"`
	FailedIPs    []string `json:"failedIPs,omitempty" jsonschema:"失败的IP列表"`
	Message      string   `json:"message" jsonschema:"操作结果消息"`
}

// CreateBatchUnblockIPs 创建批量解封IP的工具函数
func CreateBatchUnblockIPs(client *APIClient) func(context.Context, *mcp.CallToolRequest, BatchUnblockIPsInput) (*mcp.CallToolResult, BatchUnblockIPsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchUnblockIPsInput) (*mcp.CallToolResult, BatchUnblockIPsOutput, error) {
		logger := NewToolLogger("batch_unblock_ips")
		logger.LogInput(input)
		
		if len(input.IPs) == 0 {
			logger.LogWarning("IP列表为空")
			return nil, BatchUnblockIPsOutput{
				Message: "IP列表不能为空",
			}, fmt.Errorf("IP列表为空")
		}

		successCount := 0
		failedCount := 0
		var failedIPs []string

		// 逐个解封IP
		for _, ip := range input.IPs {
			err := client.Delete(fmt.Sprintf("/api/v1/blocked-ips/%s", ip))
			if err != nil {
				failedCount++
				failedIPs = append(failedIPs, ip)
				logger.LogWarning(fmt.Sprintf("解封IP %s 失败: %v", ip, err))
			} else {
				successCount++
			}
		}

		message := fmt.Sprintf("批量解封完成: 成功 %d 个, 失败 %d 个", successCount, failedCount)
		logger.LogSuccess(message)
		
		return nil, BatchUnblockIPsOutput{
			SuccessCount: successCount,
			FailedCount:  failedCount,
			FailedIPs:    failedIPs,
			Message:      message,
		}, nil
	}
}

// BatchCreateRulesInput 批量创建规则的输入参数
type BatchCreateRulesInput struct {
	Rules []RuleCreateRequest `json:"rules" jsonschema:"要创建的规则列表"`
}

// RuleCreateRequest 单个规则创建请求
type RuleCreateRequest struct {
	Name        string      `json:"name" jsonschema:"规则名称"`
	Condition   interface{} `json:"condition" jsonschema:"规则条件"`
	Action      string      `json:"action" jsonschema:"规则动作：allow,deny,log"`
	Priority    int         `json:"priority,omitempty" jsonschema:"优先级,数字越大优先级越高"`
	Enabled     bool        `json:"enabled" jsonschema:"是否启用"`
	Description string      `json:"description,omitempty" jsonschema:"规则描述"`
}

// BatchCreateRulesOutput 批量创建规则输出
type BatchCreateRulesOutput struct {
	SuccessCount int      `json:"successCount" jsonschema:"成功创建的规则数量"`
	FailedCount  int      `json:"failedCount" jsonschema:"失败的规则数量"`
	FailedRules  []string `json:"failedRules,omitempty" jsonschema:"失败的规则名称列表"`
	CreatedIDs   []string `json:"createdIds,omitempty" jsonschema:"创建成功的规则ID列表"`
	Message      string   `json:"message" jsonschema:"操作结果消息"`
}

// CreateBatchCreateRules 创建批量创建规则的工具函数
func CreateBatchCreateRules(client *APIClient) func(context.Context, *mcp.CallToolRequest, BatchCreateRulesInput) (*mcp.CallToolResult, BatchCreateRulesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchCreateRulesInput) (*mcp.CallToolResult, BatchCreateRulesOutput, error) {
		logger := NewToolLogger("batch_create_rules")
		logger.LogInput(fmt.Sprintf("创建 %d 个规则", len(input.Rules)))
		
		if len(input.Rules) == 0 {
			logger.LogWarning("规则列表为空")
			return nil, BatchCreateRulesOutput{
				Message: "规则列表不能为空",
			}, fmt.Errorf("规则列表为空")
		}

		successCount := 0
		failedCount := 0
		var failedRules []string
		var createdIDs []string

		// 逐个创建规则
		for _, rule := range input.Rules {
			data, err := client.Post("/api/v1/micro-rules", rule)
			if err != nil {
				failedCount++
				failedRules = append(failedRules, rule.Name)
				logger.LogWarning(fmt.Sprintf("创建规则 %s 失败: %v", rule.Name, err))
				continue
			}
			
			// 解析返回的规则ID
			var result struct {
				Data struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			if err := json.Unmarshal(data, &result); err == nil && result.Data.ID != "" {
				createdIDs = append(createdIDs, result.Data.ID)
			}
			
			successCount++
		}

		message := fmt.Sprintf("批量创建规则完成: 成功 %d 个, 失败 %d 个", successCount, failedCount)
		logger.LogSuccess(message)
		
		return nil, BatchCreateRulesOutput{
			SuccessCount: successCount,
			FailedCount:  failedCount,
			FailedRules:  failedRules,
			CreatedIDs:   createdIDs,
			Message:      message,
		}, nil
	}
}

// BatchDeleteRulesInput 批量删除规则的输入参数
type BatchDeleteRulesInput struct {
	RuleIDs []string `json:"ruleIds" jsonschema:"要删除的规则ID列表"`
}

// BatchDeleteRulesOutput 批量删除规则输出
type BatchDeleteRulesOutput struct {
	SuccessCount int      `json:"successCount" jsonschema:"成功删除的规则数量"`
	FailedCount  int      `json:"failedCount" jsonschema:"失败的规则数量"`
	FailedIDs    []string `json:"failedIds,omitempty" jsonschema:"失败的规则ID列表"`
	Message      string   `json:"message" jsonschema:"操作结果消息"`
}

// CreateBatchDeleteRules 创建批量删除规则的工具函数
func CreateBatchDeleteRules(client *APIClient) func(context.Context, *mcp.CallToolRequest, BatchDeleteRulesInput) (*mcp.CallToolResult, BatchDeleteRulesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchDeleteRulesInput) (*mcp.CallToolResult, BatchDeleteRulesOutput, error) {
		logger := NewToolLogger("batch_delete_rules")
		logger.LogInput(fmt.Sprintf("删除 %d 个规则", len(input.RuleIDs)))
		
		if len(input.RuleIDs) == 0 {
			logger.LogWarning("规则ID列表为空")
			return nil, BatchDeleteRulesOutput{
				Message: "规则ID列表不能为空",
			}, fmt.Errorf("规则ID列表为空")
		}

		successCount := 0
		failedCount := 0
		var failedIDs []string

		// 逐个删除规则
		for _, ruleID := range input.RuleIDs {
			err := client.Delete(fmt.Sprintf("/api/v1/micro-rules/%s", ruleID))
			if err != nil {
				failedCount++
				failedIDs = append(failedIDs, ruleID)
				logger.LogWarning(fmt.Sprintf("删除规则 %s 失败: %v", ruleID, err))
			} else {
				successCount++
			}
		}

		message := fmt.Sprintf("批量删除规则完成: 成功 %d 个, 失败 %d 个", successCount, failedCount)
		logger.LogSuccess(message)
		
		return nil, BatchDeleteRulesOutput{
			SuccessCount: successCount,
			FailedCount:  failedCount,
			FailedIDs:    failedIDs,
			Message:      message,
		}, nil
	}
}
