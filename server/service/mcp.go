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
	"list_attack_logs",
	"get_log_stats",
	// 规则管理工具
	"list_micro_rules",
	"create_micro_rule",
	"update_micro_rule",
	"delete_micro_rule",
	// IP封禁工具
	"list_blocked_ips",
	"get_blocked_ip_stats",
	// 站点管理工具
	"list_sites",
	"get_site_details",
	// AI分析工具
	"list_attack_patterns",
	"list_generated_rules",
	"trigger_ai_analysis",
	"review_rule",
	"deploy_rule",
	// 配置管理工具
	"get_waf_config",
	"update_waf_config",
	"get_stats_overview",
	// 批量操作工具
	"batch_block_ips",
	"batch_unblock_ips",
	"batch_create_rules",
	"batch_delete_rules",
	// 监控工具
	"get_realtime_qps",
	"get_time_series_data",
	"get_security_metrics",
	"get_system_health",
	// 高级AI分析工具
	"analyze_attack_patterns",
	"generate_rule_from_pattern",
	"evaluate_rule_effectiveness",
	"optimize_rule",
	"compare_rules",
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

	if connected {
		// 获取最后连接时间（从数据库获取最近的工具调用时间）
		lastCall, err := s.mcpRepo.GetLastToolCall(ctx)
		if err == nil && lastCall != nil {
			timestamp := lastCall.Timestamp.Format(time.RFC3339)
			status.LastConnectedAt = &timestamp
		}
	} else {
		status.Error = "MCP Server未被使用或无调用记录（最近5分钟内无工具调用）"
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
	// MCP连接状态的判断逻辑：
	// 1. MCP Server 是独立的 stdio 进程，被 AnythingLLM/Claude Desktop 调用
	// 2. MCP Server 通过 HTTP 调用后端 API
	// 3. 如果最近有工具调用记录，说明 MCP Server 正在被使用
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	lastCall, err := s.mcpRepo.GetLastToolCall(ctx)
	if err != nil || lastCall == nil {
		// 数据库中没有调用记录，可能是：
		// 1. MCP Server 从未被使用
		// 2. 工具调用没有被记录（需要添加记录逻辑）
		return false
	}

	// 如果最近5分钟内有工具调用，认为 MCP Server 处于活跃状态
	return time.Since(lastCall.Timestamp) < 5*time.Minute
}
