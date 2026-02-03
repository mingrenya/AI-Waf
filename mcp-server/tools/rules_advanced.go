// tools/rules_advanced.go
// 高级规则管理工具：导入/导出、批量操作等
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

// ========== 规则导入/导出工具 ==========

// ExportRulesInput 导出规则的输入参数
type ExportRulesInput struct {
	Format          string            `json:"format" jsonschema:"required,Export format (json or yaml)"`
	Filter          map[string]string `json:"filter,omitempty" jsonschema:"Filter conditions (optional fields: type, status, priority_min)"`
	IncludeDisabled bool              `json:"include_disabled" jsonschema:"Whether to include disabled rules"`
}

// ExportRulesOutput 导出规则的输出
type ExportRulesOutput struct {
	Content   string `json:"content" jsonschema:"Exported rules content in JSON or YAML format"`
	RuleCount int    `json:"rule_count" jsonschema:"Number of exported rules"`
	Format    string `json:"format" jsonschema:"Export format"`
	Message   string `json:"message" jsonschema:"Operation result message"`
}

// CreateExportRules 创建导出规则的工具函数
func CreateExportRules(client *APIClient) func(context.Context, *mcp.CallToolRequest, ExportRulesInput) (*mcp.CallToolResult, ExportRulesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ExportRulesInput) (*mcp.CallToolResult, ExportRulesOutput, error) {
		logger := NewToolLogger("export_rules")
		logger.LogInput(input)

		// 构建查询参数 - 使用大的size值以获取所有规则
		query := "?size=1000&page=1&"
		if input.IncludeDisabled {
			query += "include_disabled=true&"
		}
		for key, value := range input.Filter {
			query += fmt.Sprintf("%s=%s&", key, value)
		}

		// 获取所有规则
		path := "/api/v1/micro-rules" + query
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, ExportRulesOutput{}, WrapError(err, "获取规则")
		}

		var result struct {
			Data struct {
				Items []interface{} `json:"items"`
				Total int64         `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, ExportRulesOutput{}, FormatParseError("响应", err)
		}

		// 格式化输出
		var content string
		if input.Format == "yaml" {
			// 使用yaml.v3库进行真正的YAML格式化
			contentBytes, err := yaml.Marshal(result.Data.Items)
			if err != nil {
				logger.LogError(err)
				return nil, ExportRulesOutput{}, FormatParseError("YAML序列化", err)
			}
			content = string(contentBytes)
		} else {
			// JSON格式输出
			contentBytes, err := json.MarshalIndent(result.Data.Items, "", "  ")
			if err != nil {
				logger.LogError(err)
				return nil, ExportRulesOutput{}, FormatParseError("JSON序列化", err)
			}
			content = string(contentBytes)
		}

		logger.LogSuccess(fmt.Sprintf("已导出 %d 条规则", len(result.Data.Items)))

		return nil, ExportRulesOutput{
			Content:   content,
			RuleCount: len(result.Data.Items),
			Format:    input.Format,
			Message:   fmt.Sprintf("成功导出 %d 条规则", len(result.Data.Items)),
		}, nil
	}
}

// ImportRulesInput 导入规则的输入参数
type ImportRulesInput struct {
	RulesContent string `json:"rules_content" jsonschema:"required,Rules content in JSON or YAML format"`
	Format       string `json:"format" jsonschema:"required,Import format (json or yaml)"`
	MergeMode    string `json:"merge_mode" jsonschema:"Merge mode (replace, merge, or skip_duplicate, default merge)"`
	DryRun       bool   `json:"dry_run" jsonschema:"Whether to only validate without importing"`
}

// ImportRulesOutput 导入规则的输出
type ImportRulesOutput struct {
	ImportedCount int      `json:"imported_count" jsonschema:"Number of successfully imported rules"`
	SkippedCount  int      `json:"skipped_count" jsonschema:"Number of skipped rules"`
	FailedCount   int      `json:"failed_count" jsonschema:"Number of failed rules"`
	FailedRules   []string `json:"failed_rules,omitempty" jsonschema:"List of failed rule names"`
	Message       string   `json:"message" jsonschema:"Operation result message"`
}

// CreateImportRules 创建导入规则的工具函数
func CreateImportRules(client *APIClient) func(context.Context, *mcp.CallToolRequest, ImportRulesInput) (*mcp.CallToolResult, ImportRulesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ImportRulesInput) (*mcp.CallToolResult, ImportRulesOutput, error) {
		logger := NewToolLogger("import_rules")
		logger.LogInput(input)

		if input.MergeMode == "" {
			input.MergeMode = "merge"
		}

		// 解析规则内容
		var rules []map[string]interface{}
		if err := json.Unmarshal([]byte(input.RulesContent), &rules); err != nil {
			return nil, ImportRulesOutput{}, FormatParseError("规则内容", err)
		}

		if input.DryRun {
			logger.LogSuccess(fmt.Sprintf("验证模式: 共 %d 条规则待导入", len(rules)))
			return nil, ImportRulesOutput{
				ImportedCount: 0,
				SkippedCount:  0,
				Message:       fmt.Sprintf("验证通过，共 %d 条规则可导入", len(rules)),
			}, nil
		}

		// 逐个导入规则
		imported := 0
		skipped := 0
		failed := 0
		var failedRules []string

		for _, rule := range rules {
			// 根据合并模式处理
			if input.MergeMode == "skip_duplicate" {
				// 检查是否存在同名规则
				if name, ok := rule["name"].(string); ok {
					existingRules, err := client.Get(fmt.Sprintf("/api/v1/micro-rules?name=%s", name))
					if err == nil {
						var checkResult struct {
							Data struct {
								Total int `json:"total"`
							} `json:"data"`
						}
						json.Unmarshal(existingRules, &checkResult)
						if checkResult.Data.Total > 0 {
							skipped++
							continue
						}
					}
				}
			}

			// 创建规则
			_, err := client.Post("/api/v1/micro-rules", rule)
			if err != nil {
				failed++
				if name, ok := rule["name"].(string); ok {
					failedRules = append(failedRules, name)
				}
				logger.LogWarning(fmt.Sprintf("导入规则失败: %v", err))
			} else {
				imported++
			}
		}

		message := fmt.Sprintf("导入完成: 成功 %d 条, 跳过 %d 条, 失败 %d 条", imported, skipped, failed)
		logger.LogSuccess(message)

		return nil, ImportRulesOutput{
			ImportedCount: imported,
			SkippedCount:  skipped,
			FailedCount:   failed,
			FailedRules:   failedRules,
			Message:       message,
		}, nil
	}
}

// ========== 批量更新规则工具 ==========

// BatchUpdateRulesInput 批量更新规则的输入参数
type BatchUpdateRulesInput struct {
	RuleIDs  []string               `json:"rule_ids" jsonschema:"required,List of rule IDs to update"`
	Updates  map[string]interface{} `json:"updates" jsonschema:"required,Fields to update (e.g. status, priority)"`
	Rollback bool                   `json:"rollback" jsonschema:"Whether to rollback on failure"`
}

// BatchUpdateRulesOutput 批量更新规则输出
type BatchUpdateRulesOutput struct {
	SuccessCount int      `json:"success_count" jsonschema:"Number of successfully updated rules"`
	FailedCount  int      `json:"failed_count" jsonschema:"Number of failed rules"`
	FailedIDs    []string `json:"failed_ids,omitempty" jsonschema:"List of failed rule IDs"`
	Message      string   `json:"message" jsonschema:"Operation result message"`
}

// CreateBatchUpdateRules 创建批量更新规则的工具函数
func CreateBatchUpdateRules(client *APIClient) func(context.Context, *mcp.CallToolRequest, BatchUpdateRulesInput) (*mcp.CallToolResult, BatchUpdateRulesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchUpdateRulesInput) (*mcp.CallToolResult, BatchUpdateRulesOutput, error) {
		logger := NewToolLogger("batch_update_rules")
		logger.LogInput(input)

		successCount := 0
		failedCount := 0
		var failedIDs []string
		var successIDs []string

		// 逐个更新规则
		for _, ruleID := range input.RuleIDs {
			_, err := client.Put(fmt.Sprintf("/api/v1/micro-rules/%s", ruleID), input.Updates)
			if err != nil {
				failedCount++
				failedIDs = append(failedIDs, ruleID)
				logger.LogWarning(fmt.Sprintf("更新规则 %s 失败: %v", ruleID, err))
			} else {
				successCount++
				successIDs = append(successIDs, ruleID)
			}
		}

		// 失败时回滚
		if input.Rollback && failedCount > 0 {
			logger.LogWarning("检测到失败，开始回滚...")
			// 这里应该实现回滚逻辑（需要记录原始状态）
			logger.LogWarning("回滚功能需要实现状态快照机制")
		}

		message := fmt.Sprintf("批量更新完成: 成功 %d 条, 失败 %d 条", successCount, failedCount)
		logger.LogSuccess(message)

		return nil, BatchUpdateRulesOutput{
			SuccessCount: successCount,
			FailedCount:  failedCount,
			FailedIDs:    failedIDs,
			Message:      message,
		}, nil
	}
}

// ========== 规则测试工具 ==========

// TestRuleInput 测试规则的输入参数
type TestRuleInput struct {
	RuleID     string                   `json:"rule_id,omitempty" jsonschema:"Rule ID for testing existing rules"`
	RuleConfig map[string]interface{}   `json:"rule_config,omitempty" jsonschema:"Rule configuration for testing unsaved rules"`
	TestCases  []map[string]interface{} `json:"test_cases" jsonschema:"required,List of test cases"`
}

// TestRuleOutput 测试规则的输出
type TestRuleOutput struct {
	TotalTests  int                      `json:"total_tests" jsonschema:"Total number of tests"`
	PassedTests int                      `json:"passed_tests" jsonschema:"Number of passed tests"`
	FailedTests int                      `json:"failed_tests" jsonschema:"Number of failed tests"`
	Results     []map[string]interface{} `json:"results" jsonschema:"Detailed test results"`
	Message     string                   `json:"message" jsonschema:"Test result summary"`
}

// CreateTestRule 创建测试规则的工具函数
func CreateTestRule(client *APIClient) func(context.Context, *mcp.CallToolRequest, TestRuleInput) (*mcp.CallToolResult, TestRuleOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input TestRuleInput) (*mcp.CallToolResult, TestRuleOutput, error) {
		logger := NewToolLogger("test_rule")
		logger.LogInput(input)

		// 实现规则测试逻辑
		// 这里应该调用后端的规则测试API（如果有的话）
		// 或者在MCP层面实现简单的规则评估逻辑

		totalTests := len(input.TestCases)
		passedTests := 0
		failedTests := 0
		results := make([]map[string]interface{}, 0)

		// 简化的测试逻辑示例
		for i, testCase := range input.TestCases {
			// 这里应该根据规则配置评估测试用例
			result := map[string]interface{}{
				"test_number": i + 1,
				"input":       testCase,
				"expected":    testCase["expected"],
				"actual":      "to_be_implemented",
				"passed":      false,
			}
			results = append(results, result)
		}

		message := fmt.Sprintf("测试完成: 总计 %d, 通过 %d, 失败 %d", totalTests, passedTests, failedTests)
		logger.LogSuccess(message)

		return nil, TestRuleOutput{
			TotalTests:  totalTests,
			PassedTests: passedTests,
			FailedTests: failedTests,
			Results:     results,
			Message:     message,
		}, nil
	}
}
