// tools/batch_operations.go
// 批量操作工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BatchBlockIPsInput 批量封禁IP的输入参数
type BatchBlockIPsInput struct {
	IPs      []string `json:"ips" jsonschema:"List of IP addresses to block"`
	Reason   string   `json:"reason" jsonschema:"Reason for blocking"`
	Duration int      `json:"duration,omitempty" jsonschema:"Block duration in seconds (0 for permanent)"`
}

// BatchBlockIPsOutput 批量封禁IP输出
type BatchBlockIPsOutput struct {
	SuccessCount int      `json:"successCount" jsonschema:"Number of successfully blocked IPs"`
	FailedCount  int      `json:"failedCount" jsonschema:"Number of failed IP blocks"`
	FailedIPs    []string `json:"failedIPs,omitempty" jsonschema:"List of failed IP addresses"`
	Message      string   `json:"message" jsonschema:"Operation result message"`
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
				}, NewValidationErrorWithSuggestion(
					"ips",
					"IP列表不能为空",
					"请提供至少一个需要封禁的 IP 地址。批量操作最多支持 100 个 IP。",
				)
		}

		successCount := 0
		failedCount := 0
		var failedIPs []string
		var mu sync.Mutex // 保护共享变量
		var wg sync.WaitGroup

		// 并发封禁IP以提升性能，最多10个并发，使用context进行超时控制
		semaphore := make(chan struct{}, 10)
		for _, ip := range input.IPs {
			wg.Add(1)
			go func(targetIP string) {
				defer wg.Done()
				semaphore <- struct{}{}        // 获取信号量
				defer func() { <-semaphore }() // 释放信号量

				blockData := map[string]interface{}{
					"ip":       targetIP,
					"reason":   input.Reason,
					"duration": input.Duration,
				}

				// 使用带超时的context
				_, err := client.PostWithContext(ctx, "/api/v1/blocked-ips", blockData)
				mu.Lock()
				if err != nil {
					failedCount++
					failedIPs = append(failedIPs, targetIP)
					logger.LogWarning(fmt.Sprintf("封禁IP %s 失败: %v", targetIP, err))
				} else {
					successCount++
				}
				mu.Unlock()
			}(ip)
		}
		wg.Wait() // 等待所有goroutine完成

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
	IPs []string `json:"ips" jsonschema:"List of IP addresses to unblock"`
}

// BatchUnblockIPsOutput 批量解封IP输出
type BatchUnblockIPsOutput struct {
	SuccessCount int      `json:"successCount" jsonschema:"Number of successfully unblocked IPs"`
	FailedCount  int      `json:"failedCount" jsonschema:"Number of failed IP unblocks"`
	FailedIPs    []string `json:"failedIPs,omitempty" jsonschema:"List of failed IP addresses"`
	Message      string   `json:"message" jsonschema:"Operation result message"`
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
				}, NewValidationErrorWithSuggestion(
					"ips",
					"IP列表不能为空",
					"请提供至少一个需要解封的 IP 地址。批量操作最多支持 100 个 IP。",
				)
		}

		successCount := 0
		failedCount := 0
		var failedIPs []string
		var mu sync.Mutex // 保护共享变量
		var wg sync.WaitGroup

		// 并发解封IP以提升性能，最多10个并发，使用context进行超时控制
		semaphore := make(chan struct{}, 10)
		for _, ip := range input.IPs {
			wg.Add(1)
			go func(targetIP string) {
				defer wg.Done()
				semaphore <- struct{}{}        // 获取信号量
				defer func() { <-semaphore }() // 释放信号量

				// 使用带超时的context
				err := client.DeleteWithContext(ctx, fmt.Sprintf("/api/v1/blocked-ips/%s", targetIP))
				mu.Lock()
				if err != nil {
					failedCount++
					failedIPs = append(failedIPs, targetIP)
					logger.LogWarning(fmt.Sprintf("解封IP %s 失败: %v", targetIP, err))
				} else {
					successCount++
				}
				mu.Unlock()
			}(ip)
		}
		wg.Wait() // 等待所有goroutine完成

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
	Rules []RuleCreateRequest `json:"rules" jsonschema:"List of rules to create"`
}

// RuleCreateRequest 单个规则创建请求
type RuleCreateRequest struct {
	Name      string      `json:"name" jsonschema:"Rule name"`
	Type      string      `json:"type" jsonschema:"Rule type (whitelist or blacklist)"`
	Status    string      `json:"status" jsonschema:"Rule status (enabled or disabled)"`
	Priority  int         `json:"priority" jsonschema:"Priority level (higher number = higher priority)"`
	Condition interface{} `json:"condition" jsonschema:"Rule condition as JSON object"`
}

// BatchCreateRulesOutput 批量创建规则输出
type BatchCreateRulesOutput struct {
	SuccessCount int      `json:"successCount" jsonschema:"Number of successfully created rules"`
	FailedCount  int      `json:"failedCount" jsonschema:"Number of failed rule creations"`
	FailedRules  []string `json:"failedRules,omitempty" jsonschema:"List of failed rule names"`
	CreatedIDs   []string `json:"createdIds,omitempty" jsonschema:"List of successfully created rule IDs"`
	Message      string   `json:"message" jsonschema:"Operation result message"`
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
				}, NewValidationErrorWithSuggestion(
					"rules",
					"规则列表不能为空",
					"请提供至少一个需要创建的规则。批量操作最多支持 50 个规则。",
				)
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
	RuleIDs []string `json:"ruleIds" jsonschema:"List of rule IDs to delete"`
}

// BatchDeleteRulesOutput 批量删除规则输出
type BatchDeleteRulesOutput struct {
	SuccessCount int      `json:"successCount" jsonschema:"Number of successfully deleted rules"`
	FailedCount  int      `json:"failedCount" jsonschema:"Number of failed rule deletions"`
	FailedIDs    []string `json:"failedIDs,omitempty" jsonschema:"List of failed rule IDs"`
	Message      string   `json:"message" jsonschema:"Operation result message"`
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
				}, NewValidationErrorWithSuggestion(
					"ruleIds",
					"规则ID列表不能为空",
					"请提供至少一个需要删除的规则 ID。批量操作最多支持 50 个规则。",
				)
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
