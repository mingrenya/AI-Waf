package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// AttackPattern 攻击模式
// @Description AI检测到的攻击模式
type AttackPattern struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name         string        `json:"name" bson:"name"`                                     // 模式名称
	Description  string        `json:"description" bson:"description"`                       // 模式描述
	PatternType  string        `json:"patternType" bson:"patternType"`                       // 模式类型: sql_injection, xss, path_traversal等
	Confidence   float64       `json:"confidence" bson:"confidence"`                         // 置信度 0-1
	Severity     string        `json:"severity" bson:"severity"`                             // 严重程度: low, medium, high, critical
	
	// 模式特征
	URLPattern   string        `json:"urlPattern" bson:"urlPattern"`                         // URL模式
	PathPattern  string        `json:"pathPattern" bson:"pathPattern"`                       // 路径模式
	IPPattern    string        `json:"ipPattern" bson:"ipPattern"`                           // IP模式(CIDR)
	PayloadRegex string        `json:"payloadRegex" bson:"payloadRegex"`                     // 载荷正则表达式
	
	// 统计信息
	SampleCount  int           `json:"sampleCount" bson:"sampleCount"`                       // 样本数量
	Frequency    float64       `json:"frequency" bson:"frequency"`                           // 频率(次/秒)
	FirstSeen    time.Time     `json:"firstSeen" bson:"firstSeen"`                           // 首次发现时间
	LastSeen     time.Time     `json:"lastSeen" bson:"lastSeen"`                             // 最后发现时间
	
	// 关联规则
	GeneratedRuleIDs []string  `json:"generatedRuleIds" bson:"generatedRuleIds"`             // 已生成的规则ID列表
	
	// 元信息
	Status       string        `json:"status" bson:"status"`                                 // 状态: active, archived
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt" bson:"updatedAt"`
}

func (a *AttackPattern) GetCollectionName() string {
	return "attack_patterns"
}

// GeneratedRule AI生成的防护规则
// @Description 基于攻击模式生成的ModSecurity规则
type GeneratedRule struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string        `json:"name" bson:"name"`                                   // 规则名称
	Description     string        `json:"description" bson:"description"`                     // 规则描述
	RuleType        string        `json:"ruleType" bson:"ruleType"`                           // 规则类型: modsecurity, micro_rule
	
	// ModSecurity规则
	SecLangDirective string       `json:"secLangDirective" bson:"secLangDirective"`           // SecLang指令
	
	// MicroRule规则
	MicroRuleCondition bson.Raw   `json:"microRuleCondition,omitempty" bson:"microRuleCondition,omitempty"` // MicroRule条件
	
	// 关联信息
	PatternID       string        `json:"patternId" bson:"patternId"`                         // 关联的攻击模式ID
	PatternName     string        `json:"patternName" bson:"patternName"`                     // 模式名称
	
	// 规则配置
	Confidence      float64       `json:"confidence" bson:"confidence"`                       // 置信度
	Severity        string        `json:"severity" bson:"severity"`                           // 严重程度
	Action          string        `json:"action" bson:"action"`                               // 动作: block, log
	
	// 部署状态
	Status          string        `json:"status" bson:"status"`                               // 状态: pending, approved, deployed, rejected
	ReviewRequired  bool          `json:"reviewRequired" bson:"reviewRequired"`               // 是否需要审核
	ReviewedBy      string        `json:"reviewedBy,omitempty" bson:"reviewedBy,omitempty"`   // 审核人
	ReviewedAt      time.Time     `json:"reviewedAt,omitempty" bson:"reviewedAt,omitempty"`   // 审核时间
	ReviewComment   string        `json:"reviewComment,omitempty" bson:"reviewComment,omitempty"` // 审核意见
	
	// 部署信息
	DeployedAt      time.Time     `json:"deployedAt,omitempty" bson:"deployedAt,omitempty"`   // 部署时间
	DeployedRuleID  string        `json:"deployedRuleId,omitempty" bson:"deployedRuleId,omitempty"` // 部署后的规则ID
	
	// 效果统计
	MatchCount      int64         `json:"matchCount" bson:"matchCount"`                       // 匹配次数
	BlockCount      int64         `json:"blockCount" bson:"blockCount"`                       // 拦截次数
	FalsePositive   int64         `json:"falsePositive" bson:"falsePositive"`                 // 误报次数
	
	// 元信息
	CreatedAt       time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt" bson:"updatedAt"`
}

func (g *GeneratedRule) GetCollectionName() string {
	return "generated_rules"
}

// AIAnalyzerConfig AI分析器配置
// @Description AI安全分析器的配置信息
type AIAnalyzerConfig struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string        `json:"name" bson:"name"`
	
	// 全局开关
	Enabled  bool          `json:"enabled" bson:"enabled"`
	
	// 模式检测配置
	PatternDetection struct {
		Enabled          bool    `json:"enabled" bson:"enabled"`
		MinSamples       int     `json:"minSamples" bson:"minSamples"`                     // 最小样本数
		AnomalyThreshold float64 `json:"anomalyThreshold" bson:"anomalyThreshold"`         // 异常阈值
		ClusteringMethod string  `json:"clusteringMethod" bson:"clusteringMethod"`         // 聚类方法
		TimeWindow       int     `json:"timeWindow" bson:"timeWindow"`                     // 时间窗口(小时)
	} `bson:"patternDetection" json:"patternDetection"`
	
	// 规则生成配置
	RuleGeneration struct {
		Enabled              bool    `json:"enabled" bson:"enabled"`
		ConfidenceThreshold  float64 `json:"confidenceThreshold" bson:"confidenceThreshold"` // 置信度阈值
		AutoDeploy           bool    `json:"autoDeploy" bson:"autoDeploy"`                   // 是否自动部署
		ReviewRequired       bool    `json:"reviewRequired" bson:"reviewRequired"`           // 是否需要审核
		DefaultAction        string  `json:"defaultAction" bson:"defaultAction"`             // 默认动作
	} `bson:"ruleGeneration" json:"ruleGeneration"`
	
	// 分析周期
	AnalysisInterval int       `json:"analysisInterval" bson:"analysisInterval"`           // 分析间隔(分钟)
	
	CreatedAt        time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt" bson:"updatedAt"`
}

func (a *AIAnalyzerConfig) GetCollectionName() string {
	return "ai_analyzer_config"
}

// MCPConversation MCP对话记录
// @Description MCP与LLM的对话历史
type MCPConversation struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SessionID    string        `json:"sessionId" bson:"sessionId"`
	Role         string        `json:"role" bson:"role"`                               // 角色: user, assistant, system
	Content      string        `json:"content" bson:"content"`                         // 对话内容
	
	// 关联信息
	PatternID    string        `json:"patternId,omitempty" bson:"patternId,omitempty"`
	RuleID       string        `json:"ruleId,omitempty" bson:"ruleId,omitempty"`
	
	// 元信息
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt"`
}

func (m *MCPConversation) GetCollectionName() string {
	return "mcp_conversations"
}
