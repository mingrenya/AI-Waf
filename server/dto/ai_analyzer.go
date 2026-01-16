package dto

import (
	"time"
	"github.com/mingrenya/AI-Waf/pkg/model"
)

// ===== 攻击模式相关 =====

// AttackPatternListRequest 攻击模式列表请求
type AttackPatternListRequest struct {
	Page         int        `form:"page" binding:"omitempty,min=1"`
	Size         int        `form:"size" binding:"omitempty,min=1,max=100"`
	PatternType  string     `form:"patternType"`  // sql_injection, xss等
	AttackType   string     `form:"attackType"`   // 攻击类型(与patternType相同)
	Severity     string     `form:"severity"`     // low, medium, high, critical
	Status       string     `form:"status"`       // active, archived
	StartTime    *time.Time `form:"startTime"`    // 开始时间
	EndTime      *time.Time `form:"endTime"`      // 结束时间
}

// AttackPatternResponse 攻击模式响应
type AttackPatternResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	PatternType     string    `json:"patternType"`
	Confidence      float64   `json:"confidence"`
	Severity        string    `json:"severity"`
	URLPattern      string    `json:"urlPattern"`
	PathPattern     string    `json:"pathPattern"`
	IPPattern       string    `json:"ipPattern"`
	PayloadRegex    string    `json:"payloadRegex"`
	SampleCount     int       `json:"sampleCount"`
	Frequency       float64   `json:"frequency"`
	FirstSeen       time.Time `json:"firstSeen"`
	LastSeen        time.Time `json:"lastSeen"`
	GeneratedRuleIDs []string `json:"generatedRuleIds"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// AttackPatternListResponse 攻击模式列表响应
type AttackPatternListResponse struct {
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	List  []model.AttackPattern `json:"list"` // 直接返回model以避免转换
}

// AttackPatternStatsResponse 攻击模式统计响应
type AttackPatternStatsResponse struct {
	TotalPatterns   int64            `json:"totalPatterns"`
	ActivePatterns  int64            `json:"activePatterns"`
	ByType          map[string]int   `json:"byType"`
	BySeverity      map[string]int   `json:"bySeverity"`
	RecentDetections int             `json:"recentDetections"` // 最近24小时
}

// ===== 生成规则相关 =====

// GeneratedRuleListRequest 生成规则列表请求
type GeneratedRuleListRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Size       int    `form:"size" binding:"omitempty,min=1,max=100"`
	RuleType   string `form:"ruleType"`   // modsecurity, micro_rule
	Status     string `form:"status"`     // pending, approved, deployed, rejected
	PatternID  string `form:"patternId"`
}

// GeneratedRuleResponse 生成规则响应
type GeneratedRuleResponse struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	RuleType           string    `json:"ruleType"`
	SecLangDirective   string    `json:"secLangDirective,omitempty"`
	MicroRuleCondition interface{} `json:"microRuleCondition,omitempty"`
	PatternID          string    `json:"patternId"`
	PatternName        string    `json:"patternName"`
	Confidence         float64   `json:"confidence"`
	Severity           string    `json:"severity"`
	Action             string    `json:"action"`
	Status             string    `json:"status"`
	ReviewRequired     bool      `json:"reviewRequired"`
	ReviewedBy         string    `json:"reviewedBy,omitempty"`
	ReviewedAt         time.Time `json:"reviewedAt,omitempty"`
	ReviewComment      string    `json:"reviewComment,omitempty"`
	DeployedAt         time.Time `json:"deployedAt,omitempty"`
	DeployedRuleID     string    `json:"deployedRuleId,omitempty"`
	MatchCount         int64     `json:"matchCount"`
	BlockCount         int64     `json:"blockCount"`
	FalsePositive      int64     `json:"falsePositive"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// GeneratedRuleListResponse 生成规则列表响应
type GeneratedRuleListResponse struct {
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
	List  []model.GeneratedRule `json:"list"` // 直接返回model以避免转换
}

// ReviewRuleRequest 审核规则请求
type ReviewRuleRequest struct {
	RuleID  string `json:"ruleId" binding:"required"`
	Action  string `json:"action" binding:"required,oneof=approve reject"`  // approve, reject
	Comment string `json:"comment" binding:"required"`
}

// DeployRuleRequest 部署规则请求
type DeployRuleRequest struct {
	RuleIDs []string `json:"ruleIds" binding:"required,min=1"`
}

// GeneratedRuleStatsResponse 生成规则统计响应
type GeneratedRuleStatsResponse struct {
	TotalRules     int64          `json:"totalRules"`
	PendingRules   int64          `json:"pendingRules"`
	ApprovedRules  int64          `json:"approvedRules"`
	DeployedRules  int64          `json:"deployedRules"`
	RejectedRules  int64          `json:"rejectedRules"`
	ByType         map[string]int `json:"byType"`
}

// ===== AI分析器配置相关 =====

// AIAnalyzerConfigRequest AI分析器配置请求
type AIAnalyzerConfigRequest struct {
	Enabled bool `json:"enabled"`
	
	PatternDetection struct {
		Enabled          bool    `json:"enabled"`
		MinSamples       int     `json:"minSamples" binding:"omitempty,min=10,max=10000"`
		AnomalyThreshold float64 `json:"anomalyThreshold" binding:"omitempty,min=0.5,max=10"`
		ClusteringMethod string  `json:"clusteringMethod" binding:"omitempty,oneof=kmeans dbscan"`
		TimeWindow       int     `json:"timeWindow" binding:"omitempty,min=1,max=168"` // 1-168小时
	} `json:"patternDetection"`
	
	RuleGeneration struct {
		Enabled              bool    `json:"enabled"`
		ConfidenceThreshold  float64 `json:"confidenceThreshold" binding:"omitempty,min=0,max=1"`
		AutoDeploy           bool    `json:"autoDeploy"`
		ReviewRequired       bool    `json:"reviewRequired"`
		DefaultAction        string  `json:"defaultAction" binding:"omitempty,oneof=block log"`
	} `json:"ruleGeneration"`
	
	AnalysisInterval int `json:"analysisInterval" binding:"omitempty,min=5,max=1440"` // 5-1440分钟
}

// AIAnalyzerConfigResponse AI分析器配置响应
type AIAnalyzerConfigResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	
	PatternDetection struct {
		Enabled          bool    `json:"enabled"`
		MinSamples       int     `json:"minSamples"`
		AnomalyThreshold float64 `json:"anomalyThreshold"`
		ClusteringMethod string  `json:"clusteringMethod"`
		TimeWindow       int     `json:"timeWindow"`
	} `json:"patternDetection"`
	
	RuleGeneration struct {
		Enabled              bool    `json:"enabled"`
		ConfidenceThreshold  float64 `json:"confidenceThreshold"`
		AutoDeploy           bool    `json:"autoDeploy"`
		ReviewRequired       bool    `json:"reviewRequired"`
		DefaultAction        string  `json:"defaultAction"`
	} `json:"ruleGeneration"`
	
	AnalysisInterval int       `json:"analysisInterval"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ===== MCP对话相关 =====

// MCPChatRequest MCP聊天请求
type MCPChatRequest struct {
	PatternID string `json:"patternId,omitempty"`
	RuleID    string `json:"ruleId,omitempty"`
	Message   string `json:"message" binding:"required"`
}

// MCPChatResponse MCP聊天响应
type MCPChatResponse struct {
	Response  string    `json:"response"`
	SessionID string    `json:"sessionId"`
	Timestamp time.Time `json:"timestamp"`
}

// MCPConversationResponse MCP对话历史响应
type MCPConversationResponse struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	PatternID string    `json:"patternId,omitempty"`
	RuleID    string    `json:"ruleId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// AnalyzePatternRequest 分析模式请求
type AnalyzePatternRequest struct {
	PatternID string `json:"patternId" binding:"required"`
}

// OptimizeRuleRequest 优化规则请求
type OptimizeRuleRequest struct {
	RuleID string `json:"ruleId" binding:"required"`
}

// ===== AI分析统计 =====

// AIAnalysisStatsResponse AI分析统计响应
type AIAnalysisStatsResponse struct {
	Enabled           bool                       `json:"enabled"`
	LastAnalysisTime  time.Time                  `json:"lastAnalysisTime"`
	PatternStats      *AttackPatternStatsResponse `json:"patternStats"`
	RuleStats         *GeneratedRuleStatsResponse `json:"ruleStats"`
}

// TriggerAnalysisRequest 触发分析请求
type TriggerAnalysisRequest struct {
	Force bool `json:"force"` // 是否强制立即分析
}
