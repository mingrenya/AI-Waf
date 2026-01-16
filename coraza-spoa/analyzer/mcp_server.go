package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MCPServer MCP(Model Context Protocol)服务器实例
type MCPServer struct {
	db        *mongo.Database
	logger    Logger
	server    *mcp.Server
	detector  *AttackPatternDetector
	generator *RuleGenerator
}

// NewMCPServer 创建MCP服务器实例
func NewMCPServer(db *mongo.Database, zlogger zerolog.Logger, logger Logger) *MCPServer {
	// 创建基础 MCP Server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ai-waf-analyzer",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Instructions: "AI-WAF Security Analyzer - 提供 WAF 日志分析、攻击模式检测和规则生成能力",
	})

	// 创建 zeroLogAdapter (不再需要)
	// adapter := &zeroLogAdapter{logger: zlogger}

	mcpServer := &MCPServer{
		db:        db,
		logger:    logger,
		server:    server,
		detector:  NewAttackPatternDetector(db, zlogger),
		generator: NewRuleGenerator(db, logger),
	}

	// 注册所有 MCP 工具
	mcpServer.registerTools()

	return mcpServer
}

// Run 运行 MCP Server (阻塞,直到客户端断开连接)
func (s *MCPServer) Run(ctx context.Context) error {
	transport := &mcp.StdioTransport{}
	return s.server.Run(ctx, transport)
}

// registerTools 注册所有 MCP 工具
func (s *MCPServer) registerTools() {
	// 1. 攻击模式分析工具
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "analyze_attack_patterns",
		Description: "分析 WAF 日志并使用机器学习检测攻击模式",
	}, s.analyzeAttackPatterns)

	// 2. 规则生成工具
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "generate_waf_rules",
		Description: "基于攻击模式生成 ModSecurity 或微规则",
	}, s.generateWAFRules)

	// 3. WAF 统计工具
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_waf_statistics",
		Description: "获取 WAF 系统统计信息",
	}, s.getWAFStatistics)

	// 4. 攻击模式详情工具
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_attack_pattern",
		Description: "获取指定攻击模式的详细信息",
	}, s.getAttackPattern)

	// 5. 规则列表工具
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "list_generated_rules",
		Description: "列出 AI 生成的防护规则",
	}, s.listGeneratedRules)

	// 6. 最近攻击工具
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_recent_attacks",
		Description: "获取最近的攻击日志记录",
	}, s.getRecentAttacks)
}

// ===== Tool Handlers =====

// AnalyzeAttackPatternsInput 分析攻击模式的输入参数
type AnalyzeAttackPatternsInput struct {
	TimeWindowHours     int     `json:"time_window_hours" jsonschema_description:"分析时间窗口(小时)" default:"24"`
	MinSamples          int     `json:"min_samples" jsonschema_description:"最小样本数" default:"100"`
	ConfidenceThreshold float64 `json:"confidence_threshold" jsonschema_description:"置信度阈值(0-1)" default:"0.7"`
}

// AnalyzeAttackPatternsOutput 分析结果
type AnalyzeAttackPatternsOutput struct {
	Patterns []*model.AttackPattern `json:"patterns" jsonschema_description:"检测到的攻击模式"`
	Summary  string                 `json:"summary" jsonschema_description:"分析摘要"`
}

func (s *MCPServer) analyzeAttackPatterns(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input AnalyzeAttackPatternsInput,
) (*mcp.CallToolResult, AnalyzeAttackPatternsOutput, error) {
	// 设置默认值
	if input.TimeWindowHours == 0 {
		input.TimeWindowHours = 24
	}
	if input.MinSamples == 0 {
		input.MinSamples = 100
	}
	if input.ConfidenceThreshold == 0 {
		input.ConfidenceThreshold = 0.7
	}

	// 调用检测器
	patterns, err := s.detector.DetectPatterns()
	if err != nil {
		return nil, AnalyzeAttackPatternsOutput{}, fmt.Errorf("分析失败: %w", err)
	}

	summary := fmt.Sprintf("在过去 %d 小时内检测到 %d 个攻击模式", input.TimeWindowHours, len(patterns))

	return nil, AnalyzeAttackPatternsOutput{
		Patterns: patterns,
		Summary:  summary,
	}, nil
}

// GenerateWAFRulesInput 生成规则的输入参数
type GenerateWAFRulesInput struct {
	PatternID string `json:"pattern_id" jsonschema_description:"攻击模式ID" required:"true"`
	RuleType  string `json:"rule_type" jsonschema_description:"规则类型(modsecurity/micro_rule)" default:"modsecurity"`
}

// GenerateWAFRulesOutput 生成规则的输出
type GenerateWAFRulesOutput struct {
	RuleID      string `json:"rule_id" jsonschema_description:"生成的规则ID"`
	RuleContent string `json:"rule_content" jsonschema_description:"规则内容"`
	Message     string `json:"message" jsonschema_description:"操作消息"`
}

func (s *MCPServer) generateWAFRules(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GenerateWAFRulesInput,
) (*mcp.CallToolResult, GenerateWAFRulesOutput, error) {
	if input.PatternID == "" {
		return nil, GenerateWAFRulesOutput{}, fmt.Errorf("pattern_id 是必需的")
	}

	if input.RuleType == "" {
		input.RuleType = "modsecurity"
	}

	// 转换 PatternID 为 ObjectID
	patternOID, err := bson.ObjectIDFromHex(input.PatternID)
	if err != nil {
		return nil, GenerateWAFRulesOutput{}, fmt.Errorf("无效的 pattern_id: %w", err)
	}

	// 获取模式
	var pattern model.AttackPattern
	err = s.db.Collection("attack_patterns").FindOne(ctx, bson.M{"_id": patternOID}).Decode(&pattern)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, GenerateWAFRulesOutput{}, fmt.Errorf("未找到攻击模式: %s", input.PatternID)
		}
		return nil, GenerateWAFRulesOutput{}, fmt.Errorf("查询模式失败: %w", err)
	}

	// 生成规则
	rules, err := s.generator.GenerateRules([]*model.AttackPattern{&pattern}, 0.0)
	if err != nil || len(rules) == 0 {
		return nil, GenerateWAFRulesOutput{}, fmt.Errorf("生成规则失败: %w", err)
	}

	rule := rules[0]

	return nil, GenerateWAFRulesOutput{
		RuleID:      rule.ID.Hex(),
		RuleContent: rule.SecLangDirective,
		Message:     fmt.Sprintf("成功生成 %s 规则", input.RuleType),
	}, nil
}

// GetWAFStatisticsInput 获取统计的输入参数
type GetWAFStatisticsInput struct {
	TimeRange string `json:"time_range" jsonschema_description:"时间范围(1h/24h/7d/30d)" default:"24h"`
}

// GetWAFStatisticsOutput 统计输出
type GetWAFStatisticsOutput struct {
	TotalRequests  int64             `json:"total_requests" jsonschema_description:"总请求数"`
	BlockedCount   int64             `json:"blocked_count" jsonschema_description:"拦截数"`
	AttackTypes    map[string]int64  `json:"attack_types" jsonschema_description:"攻击类型统计"`
	TopAttackIPs   []string          `json:"top_attack_ips" jsonschema_description:"攻击IP Top 10"`
	TimeRange      string            `json:"time_range" jsonschema_description:"时间范围"`
}

func (s *MCPServer) getWAFStatistics(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetWAFStatisticsInput,
) (*mcp.CallToolResult, GetWAFStatisticsOutput, error) {
	if input.TimeRange == "" {
		input.TimeRange = "24h"
	}

	// 解析时间范围
	var duration time.Duration
	switch input.TimeRange {
	case "1h":
		duration = 1 * time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		return nil, GetWAFStatisticsOutput{}, fmt.Errorf("无效的时间范围: %s", input.TimeRange)
	}

	startTime := time.Now().Add(-duration)

	// 查询总请求数
	totalCount, err := s.db.Collection("waf_log").CountDocuments(ctx, bson.M{
		"timestamp": bson.M{"$gte": startTime},
	})
	if err != nil {
		return nil, GetWAFStatisticsOutput{}, fmt.Errorf("查询失败: %w", err)
	}

	// 查询拦截数
	blockedCount, err := s.db.Collection("waf_log").CountDocuments(ctx, bson.M{
		"timestamp": bson.M{"$gte": startTime},
		"action":    "blocked",
	})
	if err != nil {
		return nil, GetWAFStatisticsOutput{}, fmt.Errorf("查询拦截数失败: %w", err)
	}

	// 攻击类型统计
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"timestamp": bson.M{"$gte": startTime},
			"action":    "blocked",
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$attack_type",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := s.db.Collection("waf_log").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, GetWAFStatisticsOutput{}, fmt.Errorf("统计攻击类型失败: %w", err)
	}
	defer cursor.Close(ctx)

	attackTypes := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			attackTypes[result.ID] = result.Count
		}
	}

	// Top 攻击 IP
	topIPPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"timestamp": bson.M{"$gte": startTime},
			"action":    "blocked",
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$source_ip",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: 10}},
	}

	ipCursor, err := s.db.Collection("waf_log").Aggregate(ctx, topIPPipeline)
	if err != nil {
		return nil, GetWAFStatisticsOutput{}, fmt.Errorf("统计攻击IP失败: %w", err)
	}
	defer ipCursor.Close(ctx)

	var topIPs []string
	for ipCursor.Next(ctx) {
		var result struct {
			ID string `bson:"_id"`
		}
		if err := ipCursor.Decode(&result); err == nil {
			topIPs = append(topIPs, result.ID)
		}
	}

	return nil, GetWAFStatisticsOutput{
		TotalRequests: totalCount,
		BlockedCount:  blockedCount,
		AttackTypes:   attackTypes,
		TopAttackIPs:  topIPs,
		TimeRange:     input.TimeRange,
	}, nil
}

// GetAttackPatternInput 获取攻击模式详情的输入
type GetAttackPatternInput struct {
	PatternID string `json:"pattern_id" jsonschema_description:"攻击模式ID" required:"true"`
}

// GetAttackPatternOutput 攻击模式详情输出
type GetAttackPatternOutput struct {
	Pattern model.AttackPattern `json:"pattern" jsonschema_description:"攻击模式详情"`
}

func (s *MCPServer) getAttackPattern(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetAttackPatternInput,
) (*mcp.CallToolResult, GetAttackPatternOutput, error) {
	if input.PatternID == "" {
		return nil, GetAttackPatternOutput{}, fmt.Errorf("pattern_id 是必需的")
	}

	patternOID, err := bson.ObjectIDFromHex(input.PatternID)
	if err != nil {
		return nil, GetAttackPatternOutput{}, fmt.Errorf("无效的 pattern_id: %w", err)
	}

	var pattern model.AttackPattern
	err = s.db.Collection("attack_patterns").FindOne(ctx, bson.M{"_id": patternOID}).Decode(&pattern)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, GetAttackPatternOutput{}, fmt.Errorf("未找到攻击模式: %s", input.PatternID)
		}
		return nil, GetAttackPatternOutput{}, fmt.Errorf("查询失败: %w", err)
	}

	return nil, GetAttackPatternOutput{Pattern: pattern}, nil
}

// ListGeneratedRulesInput 列出规则的输入
type ListGeneratedRulesInput struct {
	Status string `json:"status" jsonschema_description:"规则状态(pending/approved/deployed/rejected)"`
	Limit  int    `json:"limit" jsonschema_description:"返回数量" default:"20"`
}

// ListGeneratedRulesOutput 规则列表输出
type ListGeneratedRulesOutput struct {
	Rules []model.GeneratedRule `json:"rules" jsonschema_description:"规则列表"`
	Total int                   `json:"total" jsonschema_description:"总数"`
}

func (s *MCPServer) listGeneratedRules(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListGeneratedRulesInput,
) (*mcp.CallToolResult, ListGeneratedRulesOutput, error) {
	if input.Limit == 0 {
		input.Limit = 20
	}

	filter := bson.M{}
	if input.Status != "" {
		filter["status"] = input.Status
	}

	opts := options.Find().SetLimit(int64(input.Limit)).SetSort(bson.M{"created_at": -1})

	cursor, err := s.db.Collection("generated_rules").Find(ctx, filter, opts)
	if err != nil {
		return nil, ListGeneratedRulesOutput{}, fmt.Errorf("查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var rules []model.GeneratedRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, ListGeneratedRulesOutput{}, fmt.Errorf("解析失败: %w", err)
	}

	total, _ := s.db.Collection("generated_rules").CountDocuments(ctx, filter)

	return nil, ListGeneratedRulesOutput{
		Rules: rules,
		Total: int(total),
	}, nil
}

// GetRecentAttacksInput 获取最近攻击的输入
type GetRecentAttacksInput struct {
	Limit    int    `json:"limit" jsonschema_description:"返回数量" default:"50"`
	Severity string `json:"severity" jsonschema_description:"严重级别(low/medium/high/critical)"`
}

// GetRecentAttacksOutput 最近攻击输出
type GetRecentAttacksOutput struct {
	Attacks []model.WAFLog `json:"attacks" jsonschema_description:"攻击日志列表"`
	Total   int            `json:"total" jsonschema_description:"总数"`
}

func (s *MCPServer) getRecentAttacks(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetRecentAttacksInput,
) (*mcp.CallToolResult, GetRecentAttacksOutput, error) {
	if input.Limit == 0 {
		input.Limit = 50
	}

	filter := bson.M{"action": "blocked"}
	if input.Severity != "" {
		filter["severity"] = input.Severity
	}

	opts := options.Find().SetLimit(int64(input.Limit)).SetSort(bson.M{"timestamp": -1})

	cursor, err := s.db.Collection("waf_log").Find(ctx, filter, opts)
	if err != nil {
		return nil, GetRecentAttacksOutput{}, fmt.Errorf("查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	var attacks []model.WAFLog
	if err := cursor.All(ctx, &attacks); err != nil {
		return nil, GetRecentAttacksOutput{}, fmt.Errorf("解析失败: %w", err)
	}

	total, _ := s.db.Collection("waf_log").CountDocuments(ctx, filter)

	return nil, GetRecentAttacksOutput{
		Attacks: attacks,
		Total:   int(total),
	}, nil
}
