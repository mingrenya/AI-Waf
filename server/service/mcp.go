package service

import (
	"context"
	//"fmt"
	//"os"
	"time"

	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/repository"
	//"go.mongodb.org/mongo-driver/v2/bson"
)

type MCPService struct {
	mcpRepo *repository.MCPRepository
}

func NewMCPService(mcpRepo *repository.MCPRepository) *MCPService {
	return &MCPService{
		mcpRepo: mcpRepo,
	}
}

// MCP工具列表 - 与 mcp-server/main.go 中注册的工具保持一致
var mcpTools = []string{
	// 日志查询工具
	"ai_waf_list_attack_logs",
	"ai_waf_get_log_stats",
	// 规则管理工具
	"ai_waf_list_micro_rules",
	"ai_waf_create_micro_rule",
	"ai_waf_update_micro_rule",
	"ai_waf_delete_micro_rule",
	// IP封禁管理工具
	"ai_waf_list_blocked_ips",
	"ai_waf_get_blocked_ip_stats",
	// 站点管理工具
	"ai_waf_list_sites",
	"ai_waf_get_site_details",
	// AI分析器工具
	"ai_waf_list_attack_patterns",
	"ai_waf_list_generated_rules",
	"ai_waf_trigger_analysis",
	"ai_waf_review_rule",
	"ai_waf_deploy_rule",
	// 配置管理工具
	"ai_waf_get_config",
	"ai_waf_update_config",
	// 批量操作工具
	"ai_waf_batch_block_ips",
	"ai_waf_batch_unblock_ips",
	"ai_waf_batch_create_rules",
	"ai_waf_batch_delete_rules",
	// 实时监控工具
	"ai_waf_get_realtime_qps",
	"ai_waf_get_time_series_data",
	"ai_waf_get_security_metrics",
	"ai_waf_get_system_health",
	// 高级AI分析工具
	"ai_waf_analyze_attack_patterns",
	"ai_waf_generate_rule_from_pattern",
	"ai_waf_evaluate_rule_effectiveness",
	"ai_waf_optimize_rule",
	"ai_waf_compare_rules",
	// 高级规则管理工具
	"ai_waf_export_rules",
	"ai_waf_import_rules",
	"ai_waf_batch_update_rules",
	"ai_waf_test_rule",
	// 扩展工具
	"ai_waf_generate_security_report",
	"ai_waf_predict_threats",
	"ai_waf_auto_remediate",
	"ai_waf_export_audit_log",
	"ai_waf_smart_rule_suggestion",
	"ai_waf_setup_alert_policy",
	"ai_waf_get_incident_status",
	"ai_waf_compliance_check",
	"ai_waf_audit_trail_validation",
	"ai_waf_capacity_planning",
}

// GetMCPStatus 获取MCP服务器连接状态
func (s *MCPService) GetMCPStatus(ctx context.Context) (*dto.MCPStatusResponse, error) {
	// 检查MCP服务器是否在运行
	// 这里简化处理，实际应该检查MCP Server进程或健康检查端点
	connected := s.checkMCPServerConnection()

	status := &dto.MCPStatusResponse{
		Connected:      connected,
		ServerVersion:  "v1.0.0",
		TotalTools:     len(mcpTools),
		AvailableTools: mcpTools,
	}

	// 获取最后连接时间（从数据库获取最近的工具调用时间）
	if connected {
		lastCall, err := s.mcpRepo.GetLastToolCall(ctx)
		if err == nil && lastCall != nil {
			timestamp := lastCall.Timestamp.Format(time.RFC3339)
			status.LastConnectedAt = &timestamp
		}
	}

	return status, nil
}

// GetMCPTools 获取MCP工具列表
func (s *MCPService) GetMCPTools(ctx context.Context) ([]string, error) {
	return mcpTools, nil
}

// GetToolCallHistory 获取工具调用历史
func (s *MCPService) GetToolCallHistory(ctx context.Context, limit, offset int) ([]dto.MCPToolCallRecord, int64, error) {
	calls, total, err := s.mcpRepo.GetToolCallHistory(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	records := make([]dto.MCPToolCallRecord, len(calls))
	for i, call := range calls {
		records[i] = dto.MCPToolCallRecord{
			ID:        call.ID.Hex(),
			ToolName:  call.ToolName,
			Timestamp: call.Timestamp,
			Duration:  call.Duration,
			Success:   call.Success,
			Error:     call.Error,
		}
	}

	return records, total, nil
}

// RecordToolCall 记录工具调用
func (s *MCPService) RecordToolCall(ctx context.Context, toolName string, duration int64, success bool, errorMsg string) error {
	return s.mcpRepo.RecordToolCall(ctx, toolName, duration, success, errorMsg)
}

// checkMCPServerConnection 检查MCP服务器连接
func (s *MCPService) checkMCPServerConnection() bool {
	// MCP功能可用性的判断逻辑：
	// MCP Server 是 stdio 进程，无法直接通过网络检测
	// 这里返回 true 表示后端 API（MCP 功能实现）正常运行
	// "connected: true" 表示 MCP 功能可用，而不是 MCP Server 的网络连接状态
	return true
}
