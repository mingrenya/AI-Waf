// AI-Waf MCP Server
// 提供MCP协议接口，让AI应用（如Claude Desktop）能够访问WAF功能
package main

import (
	"context"
	"log"
	"os"
	//"time"

	"github.com/mingrenya/AI-Waf/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// 获取环境变量配置
	backendURL := os.Getenv("WAF_BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:2333"
	}

	apiToken := os.Getenv("WAF_API_TOKEN")
	if apiToken == "" {
		log.Println("警告: 未设置 WAF_API_TOKEN 环境变量")
	}

	// 创建后端API客户端
	client := tools.NewAPIClient(backendURL, apiToken)

	// 创建MCP Server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ai-waf",
		Version: "v1.0.0",
	}, nil)

	// 注册所有工具
	log.Println("注册MCP工具...")

	// 1. 日志查询工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_attack_logs",
		Description: "查询WAF攻击日志，支持按时间范围、攻击类型、严重程度过滤",
	}, tools.CreateListAttackLogs(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_log_stats",
		Description: "获取攻击日志统计信息，包括攻击类型分布、来源IP统计等",
	}, tools.CreateGetLogStats(client))

	// 2. 规则管理工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_micro_rules",
		Description: "列出所有MicroRule规则",
	}, tools.CreateListMicroRules(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_micro_rule",
		Description: "创建新的MicroRule规则，用于自定义访问控制",
	}, tools.CreateCreateMicroRule(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_micro_rule",
		Description: "更新现有的MicroRule规则",
	}, tools.CreateUpdateMicroRule(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_micro_rule",
		Description: "删除指定的MicroRule规则",
	}, tools.CreateDeleteMicroRule(client))

	// 3. IP封禁管理工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_blocked_ips",
		Description: "列出被封禁的IP地址及封禁原因",
	}, tools.CreateListBlockedIPs(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_blocked_ip_stats",
		Description: "获取IP封禁统计信息",
	}, tools.CreateGetBlockedIPStats(client))

	// 4. 站点管理工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sites",
		Description: "列出所有受保护的站点配置",
	}, tools.CreateListSites(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_site_details",
		Description: "获取指定站点的详细配置信息",
	}, tools.CreateGetSiteDetails(client))

	// 5. AI分析器工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_attack_patterns",
		Description: "列出AI检测到的攻击模式",
	}, tools.CreateListAttackPatterns(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_generated_rules",
		Description: "列出AI生成的防护规则",
	}, tools.CreateListGeneratedRules(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "trigger_ai_analysis",
		Description: "手动触发AI分析任务",
	}, tools.CreateTriggerAIAnalysis(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "review_rule",
		Description: "审核AI生成的规则（批准或拒绝）",
	}, tools.CreateReviewRule(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deploy_rule",
		Description: "部署已审核通过的规则到生产环境",
	}, tools.CreateDeployRule(client))

	// 6. 配置管理工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_waf_config",
		Description: "获取WAF系统配置信息",
	}, tools.CreateGetWAFConfig(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_waf_config",
		Description: "更新WAF系统配置（支持部分更新）",
	}, tools.CreateUpdateWAFConfig(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_stats_overview",
		Description: "获取WAF统计概览（请求数、阻止率、错误率等）",
	}, tools.CreateGetStatsOverview(client))

	// 7. 批量操作工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_block_ips",
		Description: "批量封禁多个IP地址",
	}, tools.CreateBatchBlockIPs(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_unblock_ips",
		Description: "批量解封多个IP地址",
	}, tools.CreateBatchUnblockIPs(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_create_rules",
		Description: "批量创建多个MicroRule规则",
	}, tools.CreateBatchCreateRules(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_delete_rules",
		Description: "批量删除多个MicroRule规则",
	}, tools.CreateBatchDeleteRules(client))

	// 8. 实时监控工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_realtime_qps",
		Description: "获取实时QPS数据（最近的请求速率）",
	}, tools.CreateGetRealtimeQPS(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_time_series_data",
		Description: "获取时间序列监控数据（请求数、错误数、响应时间等）",
	}, tools.CreateGetTimeSeriesData(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_security_metrics",
		Description: "获取安全指标（攻击统计、威胁等级等）",
	}, tools.CreateGetSecurityMetrics(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_system_health",
		Description: "获取系统健康状态（服务状态、资源使用率等）",
	}, tools.CreateGetSystemHealth(client))

	// 9. 高级AI分析工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_attack_patterns",
		Description: "分析攻击模式：使用机器学习算法(聚类、异常检测)识别攻击模式，自动生成防护规则",
	}, tools.CreateAnalyzeAttackPatterns(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_rule_from_pattern",
		Description: "从检测到的攻击模式自动生成防护规则（MicroRule或ModSecurity规则），支持自动审核",
	}, tools.CreateGenerateRuleFromPattern(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "evaluate_rule_effectiveness",
		Description: "评估已部署规则的效果：计算误报率、真阳率、性能影响、安全效果，提供优化建议",
	}, tools.CreateEvaluateRuleEffectiveness(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "optimize_rule",
		Description: "优化现有规则：根据历史数据和效果评估自动优化规则参数，提升准确性和性能",
	}, tools.CreateOptimizeRule(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "compare_rules",
		Description: "对比两条规则的效果：对比性能指标、安全效果、误报率等，帮助选择最佳规则",
	}, tools.CreateCompareRules(client))

	// 添加中间件（可选，用于调试和追踪）
	// 注意：stdio 模式下，日志会输出到 stderr，不会干扰 JSON-RPC 通信
	if os.Getenv("MCP_DEBUG") == "1" {
		server.AddReceivingMiddleware(createLoggingMiddleware())
		log.Println("[MCP] Debug mode enabled")
	}
	if os.Getenv("MCP_TRACK") == "1" {
		server.AddReceivingMiddleware(createTrackingMiddleware(client))
		log.Println("[MCP] Tracking mode enabled")
	}

	// 通过stdio与MCP客户端通信
	log.Println("================================")
	log.Println("AI-Waf MCP Server 启动成功")
	log.Printf("后端URL: %s\n", backendURL)
	log.Println("已注册31个MCP工具（日志2 + 规则4 + IP封禁2 + 站点2 + AI分析5 + 配置3 + 批量操作4 + 监控4 + 高级AI分析5）")
	log.Println("等待MCP客户端连接...")
	log.Println("提示: 看到JSON-RPC消息(如 {\"jsonrpc\":\"2.0\"...}) 即表示客户端已成功连接")
	log.Println("================================")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
