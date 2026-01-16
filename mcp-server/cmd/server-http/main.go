// AI-Waf MCP Server (HTTP版本)
// 提供HTTP接口，让后端能够检测MCP Server连接状态
// 使用方式: cd cmd/server-http && go run main.go
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mingrenya/AI-Waf/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	httpAddr = flag.String("addr", "localhost:8080", "HTTP服务器监听地址")
)

func main() {
	flag.Parse()

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
		Name:    "ai-waf-http",
		Version: "v1.0.0",
	}, nil)

	// 注册所有工具（与stdio版本相同）
	registerTools(server, client)

	// 添加中间件（参考官方 examples/server/middleware/main.go）
	// 注意：中间件按照添加顺序执行，后添加的先执行（洋葱模型）
	server.AddReceivingMiddleware(createLoggingMiddleware())    // 日志中间件
	server.AddReceivingMiddleware(createTrackingMiddleware(client)) // 追踪中间件

	// 创建 StreamableHTTPHandler
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	// 启动HTTP服务器
	log.Println("================================")
	log.Println("AI-Waf MCP Server (HTTP) 启动成功")
	log.Printf("后端URL: %s\n", backendURL)
	log.Printf("监听地址: http://%s\n", *httpAddr)
	log.Println("已注册31个MCP工具")
	log.Println("================================")

	if err := http.ListenAndServe(*httpAddr, handler); err != nil {
		log.Fatalf("HTTP服务器启动失败: %v", err)
	}
}

// registerTools 注册所有MCP工具
func registerTools(server *mcp.Server, client *tools.APIClient) {
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
}

// createLoggingMiddleware 创建日志中间件（参考官方 examples/http/logging_middleware.go）
func createLoggingMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			sessionID := req.GetSession().ID()

			// 记录请求
			log.Printf("[REQUEST] Session: %s | Method: %s", sessionID, method)

			// 详细记录工具调用
			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				log.Printf("[TOOL CALL] Name: %s | Args: %v", ctr.Params.Name, ctr.Params.Arguments)
			}

			// 执行实际方法
			result, err := next(ctx, method, req)
			duration := time.Since(start)

			// 记录响应
			if err != nil {
				log.Printf("[RESPONSE] Session: %s | Method: %s | Status: ERROR | Duration: %v | Error: %v",
					sessionID, method, duration, err)
			} else {
				log.Printf("[RESPONSE] Session: %s | Method: %s | Status: OK | Duration: %v",
					sessionID, method, duration)
				
				// 记录工具结果详情
				if ctr, ok := result.(*mcp.CallToolResult); ok {
					log.Printf("[TOOL RESULT] IsError: %v | ContentCount: %d", ctr.IsError, len(ctr.Content))
				}
			}

			return result, err
		}
	}
}

// createTrackingMiddleware 创建工具调用追踪中间件（记录到后端数据库）
func createTrackingMiddleware(client *tools.APIClient) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// 执行方法
			start := time.Now()
			result, err := next(ctx, method, req)
			duration := time.Since(start)

			// 仅记录工具调用
			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				// 异步记录到后端数据库
				go func() {
					recordData := map[string]interface{}{
						"toolName":  ctr.Params.Name,
						"arguments": ctr.Params.Arguments,
						"duration":  duration.Milliseconds(),
						"success":   err == nil,
						"timestamp": time.Now().Format(time.RFC3339),
					}
					if err != nil {
						recordData["error"] = err.Error()
					}
					
					// 调用后端记录接口
					_, recordErr := client.Post("/api/v1/mcp/tool-calls/record", recordData)
					if recordErr != nil {
						log.Printf("[TRACKING] Warning: Failed to record tool call: %v", recordErr)
					} else {
						log.Printf("[TRACKING] Recorded: %s", ctr.Params.Name)
					}
				}()
			}

			return result, err
		}
	}
}

// 使用示例:
// export WAF_BACKEND_URL=http://localhost:2333
// export WAF_API_TOKEN=your-token-here
// cd cmd/server-http && go run main.go -addr localhost:8080
