package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// AlertChannel 告警渠道配置
type AlertChannel struct {
	ID        bson.ObjectID          `bson:"_id,omitempty" json:"id"`
	Name      string                 `bson:"name" json:"name"`
	Type      string                 `bson:"type" json:"type"` // webhook, slack, discord, dingtalk, wecom
	Config    map[string]interface{} `bson:"config" json:"config"`
	Enabled   bool                   `bson:"enabled" json:"enabled"`
	CreatedAt time.Time              `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time              `bson:"updated_at" json:"updatedAt"`
}

// GetCollectionName 获取集合名称
func (AlertChannel) GetCollectionName() string {
	return "alert_channels"
}

// AlertCondition 告警条件
type AlertCondition struct {
	Metric    string      `bson:"metric" json:"metric"`       // qps, block_rate, error_rate, attack_count, etc.
	Operator  string      `bson:"operator" json:"operator"`   // >, <, >=, <=, ==, !=
	Threshold interface{} `bson:"threshold" json:"threshold"` // 阈值
	Duration  int         `bson:"duration" json:"duration"`   // 持续时间（分钟）
}

// AlertRule 告警规则
type AlertRule struct {
	ID          bson.ObjectID    `bson:"_id,omitempty" json:"id"`
	Name        string           `bson:"name" json:"name"`
	Description string           `bson:"description" json:"description"`
	Conditions  []AlertCondition `bson:"conditions" json:"conditions"`
	Logic       string           `bson:"logic" json:"logic"`               // AND, OR - 条件组合逻辑
	Channels    []string         `bson:"channels" json:"channels"`         // Channel IDs
	Template    string           `bson:"template" json:"template"`         // 消息模板
	Cooldown    int              `bson:"cooldown" json:"cooldown"`         // 冷却时间（分钟）
	Severity    string           `bson:"severity" json:"severity"`         // low, medium, high, critical
	Enabled     bool             `bson:"enabled" json:"enabled"`
	CreatedAt   time.Time        `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time        `bson:"updated_at" json:"updatedAt"`
	CreatedBy   string           `bson:"created_by,omitempty" json:"createdBy,omitempty"`
}

// GetCollectionName 获取集合名称
func (AlertRule) GetCollectionName() string {
	return "alert_rules"
}

// AlertHistory 告警历史记录
type AlertHistory struct {
	ID              bson.ObjectID          `bson:"_id,omitempty" json:"id"`
	RuleID          string                 `bson:"rule_id" json:"ruleId"`
	RuleName        string                 `bson:"rule_name" json:"ruleName"`
	Severity        string                 `bson:"severity" json:"severity"`
	Message         string                 `bson:"message" json:"message"`
	Details         map[string]interface{} `bson:"details" json:"details"`
	Channels        []string               `bson:"channels" json:"channels"`
	Status          string                 `bson:"status" json:"status"` // pending, sent, failed, acknowledged
	ErrorMessage    string                 `bson:"error_message,omitempty" json:"errorMessage,omitempty"`
	TriggeredAt     time.Time              `bson:"triggered_at" json:"triggeredAt"`
	SentAt          *time.Time             `bson:"sent_at,omitempty" json:"sentAt,omitempty"`
	AcknowledgedAt  *time.Time             `bson:"acknowledged_at,omitempty" json:"acknowledgedAt,omitempty"`
	AcknowledgedBy  string                 `bson:"acknowledged_by,omitempty" json:"acknowledgedBy,omitempty"`
}

// GetCollectionName 获取集合名称
func (AlertHistory) GetCollectionName() string {
	return "alert_history"
}

// AlertTemplate 告警模板
type AlertTemplate struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Content     string        `bson:"content" json:"content"`
	Variables   []string      `bson:"variables" json:"variables"` // 支持的变量列表
	IsBuiltIn   bool          `bson:"is_built_in" json:"isBuiltIn"`
	CreatedAt   time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updatedAt"`
	CreatedBy   string        `bson:"created_by,omitempty" json:"createdBy,omitempty"`
}

// GetCollectionName 获取集合名称
func (AlertTemplate) GetCollectionName() string {
	return "alert_templates"
}

// AlertChannelType 告警渠道类型常量
const (
	AlertChannelTypeWebhook  = "webhook"
	AlertChannelTypeSlack    = "slack"
	AlertChannelTypeDiscord  = "discord"
	AlertChannelTypeDingTalk = "dingtalk"
	AlertChannelTypeWeCom    = "wecom"
)

// AlertSeverity 告警严重级别常量
const (
	AlertSeverityLow      = "low"
	AlertSeverityMedium   = "medium"
	AlertSeverityHigh     = "high"
	AlertSeverityCritical = "critical"
)

// AlertStatus 告警状态常量
const (
	AlertStatusPending      = "pending"
	AlertStatusSent         = "sent"
	AlertStatusFailed       = "failed"
	AlertStatusAcknowledged = "acknowledged"
)

// AlertMetric 可用的告警指标常量
const (
	AlertMetricQPS         = "qps"
	AlertMetricBlockRate   = "block_rate"
	AlertMetricErrorRate   = "error_rate"
	AlertMetricAttackCount = "attack_count"
	AlertMetricTraffic     = "traffic"
	AlertMetric4xxRate     = "error_4xx_rate"
	AlertMetric5xxRate     = "error_5xx_rate"
)

// AlertOperator 告警条件运算符常量
const (
	AlertOperatorGreaterThan      = ">"
	AlertOperatorLessThan         = "<"
	AlertOperatorGreaterThanEqual = ">="
	AlertOperatorLessThanEqual    = "<="
	AlertOperatorEqual            = "=="
	AlertOperatorNotEqual         = "!="
)
