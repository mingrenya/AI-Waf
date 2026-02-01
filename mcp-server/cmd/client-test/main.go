// AI-Waf MCP Client 测试工具
// 用于测试MCP Server的工具调用功能
// 使用方式: go run client-test.go -server http://localhost:8080
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	serverURL = flag.String("server", "http://localhost:8088", "MCP Server地址")
	toolName  = flag.String("tool", "", "要测试的工具名称（空则列出所有工具）")
	args      = flag.String("args", "{}", "工具参数（JSON格式）")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "AI-Waf MCP Client 测试工具\n\n")
		fmt.Fprintf(os.Stderr, "使用方式:\n")
		fmt.Fprintf(os.Stderr, "  1. 列出所有工具:\n")
		fmt.Fprintf(os.Stderr, "     go run client-test.go -server http://localhost:8088\n\n")
		fmt.Fprintf(os.Stderr, "  2. 调用特定工具:\n")
		fmt.Fprintf(os.Stderr, "     go run client-test.go -server http://localhost:8088 -tool list_attack_logs -args '{\"limit\":10}'\n\n")
		fmt.Fprintf(os.Stderr, "参数:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()

	// 创建MCP客户端
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "ai-waf-test-client",
		Version: "v1.0.0",
	}, nil)

	// 连接到MCP Server
	log.Printf("正在连接到 MCP Server: %s", *serverURL)
	transport := &mcp.StreamableClientTransport{
		Endpoint: *serverURL,
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer session.Close()

	log.Printf("✅ 已连接到 MCP Server (会话ID: %s)\n", session.ID())

	// 如果没有指定工具名称，列出所有工具
	if *toolName == "" {
		listAllTools(ctx, session)
		return
	}

	// 调用指定的工具
	callTool(ctx, session, *toolName, *args)
}

// listAllTools 列出所有可用的工具
func listAllTools(ctx context.Context, session *mcp.ClientSession) {
	log.Println("\n正在获取工具列表...")

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("获取工具列表失败: %v", err)
	}

	fmt.Println("\n================================")
	fmt.Printf("可用工具数量: %d\n", len(result.Tools))
	fmt.Println("================================\n")

	// 按类别分组显示
	categories := map[string][]string{
		"日志查询":   {"list_attack_logs", "get_log_stats"},
		"规则管理":   {"list_micro_rules", "create_micro_rule", "update_micro_rule", "delete_micro_rule"},
		"IP封禁管理": {"list_blocked_ips", "get_blocked_ip_stats"},
		"站点管理":   {"list_sites", "get_site_details"},
		"AI分析器":  {"list_attack_patterns", "list_generated_rules", "trigger_ai_analysis", "review_rule", "deploy_rule"},
		"配置管理":   {"get_waf_config", "update_waf_config", "get_stats_overview"},
		"批量操作":   {"batch_block_ips", "batch_unblock_ips", "batch_create_rules", "batch_delete_rules"},
		"实时监控":   {"get_realtime_qps", "get_time_series_data", "get_security_metrics", "get_system_health"},
		"高级AI分析": {"analyze_attack_patterns", "generate_rule_from_pattern", "evaluate_rule_effectiveness", "optimize_rule", "compare_rules"},
	}

	// 创建工具映射
	toolMap := make(map[string]*mcp.Tool)
	for _, tool := range result.Tools {
		toolMap[tool.Name] = tool
	}

	// 按类别显示
	for category, toolNames := range categories {
		fmt.Printf("📦 %s:\n", category)
		for _, name := range toolNames {
			if tool, ok := toolMap[name]; ok {
				fmt.Printf("   • %s\n", tool.Name)
				fmt.Printf("     %s\n", tool.Description)
			}
		}
		fmt.Println()
	}

	fmt.Println("测试示例:")
	fmt.Printf("  go run client-test.go -server %s -tool list_attack_logs -args '{\"limit\":5}'\n", *serverURL)
	fmt.Printf("  go run client-test.go -server %s -tool get_stats_overview -args '{}'\n", *serverURL)
	fmt.Printf("  go run client-test.go -server %s -tool list_sites -args '{}'\n", *serverURL)
}

// callTool 调用指定的工具
func callTool(ctx context.Context, session *mcp.ClientSession, toolName, argsJSON string) {
	log.Printf("\n正在调用工具: %s", toolName)
	log.Printf("参数: %s\n", argsJSON)

	// 解析参数
	var arguments map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &arguments); err != nil {
		log.Fatalf("参数解析失败: %v", err)
	}

	// 为特定工具添加默认参数
	if toolName == "list_attack_logs" {
		if _, ok := arguments["hours"]; !ok {
			arguments["hours"] = 24 // 默认值为 24 小时
		}
	}

	// 调用工具
	start := time.Now()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("工具调用失败: %v", err)
	}

	fmt.Println("\n================================")
	fmt.Printf("工具调用成功 (耗时: %v)\n", duration)
	fmt.Println("================================\n")

	// 显示结果
	if result.IsError {
		fmt.Println("❌ 工具执行失败")
		for i, content := range result.Content {
			fmt.Printf("\n错误 %d:\n", i+1)
			printContent(content)
		}
	} else {
		fmt.Println("✅ 工具执行成功")
		for i, content := range result.Content {
			fmt.Printf("\n结果 %d:\n", i+1)
			printContent(content)
		}
	}
}

// printContent 格式化打印内容
func printContent(content mcp.Content) {
	switch c := content.(type) {
	case *mcp.TextContent:
		fmt.Printf("类型: 文本\n")
		fmt.Printf("内容:\n%s\n", c.Text)

	case *mcp.ImageContent:
		fmt.Printf("类型: 图片\n")
		fmt.Printf("数据: %s\n", c.Data)
		fmt.Printf("MIME类型: %s\n", c.MIMEType)

	case *mcp.EmbeddedResource:
		fmt.Printf("类型: 嵌入式资源\n")
		if c.Resource.URI != "" {
			fmt.Printf("URI: %s\n", c.Resource.URI)
		}
		if c.Resource.Text != "" {
			fmt.Printf("文本: %s\n", c.Resource.Text)
		}
		if len(c.Resource.Blob) > 0 {
			blobLen := len(c.Resource.Blob)
			if blobLen > 100 {
				blobLen = 100
			}
			fmt.Printf("Blob: %x... (长度: %d 字节)\n", c.Resource.Blob[:blobLen], len(c.Resource.Blob))
		}

	default:
		// 尝试JSON序列化
		data, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			fmt.Printf("未知内容类型: %T\n", content)
		} else {
			fmt.Printf("JSON:\n%s\n", string(data))
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 测试示例命令:
//
// 1. 列出所有工具:
//    go run client-test.go -server http://localhost:8080
//
// 2. 测试日志查询:
//    go run client-test.go -server http://localhost:8080 \
//      -tool list_attack_logs \
//      -args '{"limit":5,"severity":"high"}'
//
// 3. 测试统计概览:
//    go run client-test.go -server http://localhost:8080 \
//      -tool get_stats_overview \
//      -args '{}'
//
// 4. 测试站点列表:
//    go run client-test.go -server http://localhost:8080 \
//      -tool list_sites \
//      -args '{"page":1,"pageSize":10}'
//
// 5. 测试规则列表:
//    go run client-test.go -server http://localhost:8080 \
//      -tool list_micro_rules \
//      -args '{"enabled":true}'
//
// 6. 测试AI分析:
//    go run client-test.go -server http://localhost:8080 \
//      -tool trigger_ai_analysis \
//      -args '{"timeRange":"24h","analysisType":"attack_pattern"}'
