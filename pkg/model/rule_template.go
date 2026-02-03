package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// RuleTemplate 规则模板
// @Description WAF规则模板，包含预定义的 OWASP Top 10 等安全规则模板
type RuleTemplate struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string        `json:"name" bson:"name" example:"SQL注入防护模板"`               // 模板名称
	Category    string        `json:"category" bson:"category" example:"injection"`       // 分类 (OWASP Top 10)
	Description string        `json:"description" bson:"description" example:"防止SQL注入攻击"` // 描述
	Severity    string        `json:"severity" bson:"severity" example:"critical"`        // 严重等级: low, medium, high, critical
	RuleType    RuleType      `json:"rule_type" bson:"rule_type" example:"blacklist"`     // 规则类型
	Priority    int           `json:"priority" bson:"priority" example:"500"`             // 优先级
	Tags        []string      `json:"tags" bson:"tags" example:"owasp,a03,injection"`     // 标签
	Condition   bson.Raw      `json:"condition" bson:"condition" swaggertype:"object"`    // 规则条件模板
	CreatedAt   time.Time     `json:"created_at" bson:"created_at"`                       // 创建时间
	UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at"`                       // 更新时间
}

func (r *RuleTemplate) GetCollectionName() string {
	return "rule_template"
}

// RuleEffectivenessScore 规则有效性评分
// @Description 规则有效性评分，用于评估规则的实际效果
type RuleEffectivenessScore struct {
	ID                bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	RuleID            bson.ObjectID `json:"rule_id" bson:"rule_id"`                         // 关联的规则ID
	RuleName          string        `json:"rule_name" bson:"rule_name"`                     // 规则名称
	Score             float64       `json:"score" bson:"score" example:"85.5"`              // 综合评分 (0-100)
	MatchCount        int64         `json:"match_count" bson:"match_count" example:"1234"`  // 匹配次数
	BlockCount        int64         `json:"block_count" bson:"block_count" example:"987"`   // 阻止次数
	FalsePositive     int64         `json:"false_positive" bson:"false_positive"`           // 误报次数
	TruePositive      int64         `json:"true_positive" bson:"true_positive"`             // 真阳次数
	FalsePositiveRate float64       `json:"false_positive_rate" bson:"false_positive_rate"` // 误报率 (%)
	TruePositiveRate  float64       `json:"true_positive_rate" bson:"true_positive_rate"`   // 真阳率 (%)
	BlockRate         float64       `json:"block_rate" bson:"block_rate"`                   // 阻止率 (%)
	AvgMatchTime      float64       `json:"avg_match_time" bson:"avg_match_time"`           // 平均匹配时间 (ms)
	PerformanceImpact string        `json:"performance_impact" bson:"performance_impact"`   // 性能影响: low, medium, high
	Recommendation    string        `json:"recommendation" bson:"recommendation"`           // 优化建议
	LastEvaluated     time.Time     `json:"last_evaluated" bson:"last_evaluated"`           // 最后评估时间
	EvaluationPeriod  string        `json:"evaluation_period" bson:"evaluation_period"`     // 评估周期: 24h, 7d, 30d
	CreatedAt         time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at" bson:"updated_at"`
}

func (r *RuleEffectivenessScore) GetCollectionName() string {
	return "rule_effectiveness_score"
}

// ProtectionProfile 保护配置文件
// @Description 一键保护配置文件，包含预定义的规则组合
type ProtectionProfile struct {
	ID          bson.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string          `json:"name" bson:"name" example:"标准保护"`                       // 配置文件名称
	Level       string          `json:"level" bson:"level" example:"standard"`                 // 保护级别: basic, standard, strict
	Description string          `json:"description" bson:"description" example:"适合大多数应用的标准保护"` // 描述
	Categories  []string        `json:"categories" bson:"categories"`                          // 包含的分类
	TemplateIDs []bson.ObjectID `json:"template_ids" bson:"template_ids"`                      // 包含的模板ID列表
	IsDefault   bool            `json:"is_default" bson:"is_default"`                          // 是否为默认配置
	CreatedAt   time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" bson:"updated_at"`
}

func (p *ProtectionProfile) GetCollectionName() string {
	return "protection_profile"
}

// OWASP Top 10 2021 分类常量
const (
	CategoryBrokenAccessControl      = "broken_access_control"     // A01:2021 - 失效的访问控制
	CategoryCryptographicFailures    = "cryptographic_failures"    // A02:2021 - 加密机制失效
	CategoryInjection                = "injection"                 // A03:2021 - 注入
	CategoryInsecureDesign           = "insecure_design"           // A04:2021 - 不安全设计
	CategorySecurityMisconfiguration = "security_misconfiguration" // A05:2021 - 安全配置错误
	CategoryVulnerableComponents     = "vulnerable_components"     // A06:2021 - 易受攻击和过时的组件
	CategoryAuthenticationFailures   = "authentication_failures"   // A07:2021 - 识别和身份验证失败
	CategoryDataIntegrityFailures    = "data_integrity_failures"   // A08:2021 - 软件和数据完整性故障
	CategoryLoggingFailures          = "logging_failures"          // A09:2021 - 安全日志和监控失败
	CategorySSRF                     = "ssrf"                      // A10:2021 - 服务器端请求伪造 (SSRF)
)

// 保护级别常量
const (
	ProtectionLevelBasic    = "basic"    // 基础保护
	ProtectionLevelStandard = "standard" // 标准保护
	ProtectionLevelStrict   = "strict"   // 严格保护
)

// 严重等级常量
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)
