package dto

import (
	"time"

	"github.com/mingrenya/AI-Waf/server/model"
)

// AlertChannel DTO

// CreateAlertChannelRequest 创建告警渠道请求
type CreateAlertChannelRequest struct {
	Name    string                 `json:"name" binding:"required"`
	Type    string                 `json:"type" binding:"required,oneof=webhook slack discord dingtalk wecom"`
	Config  map[string]interface{} `json:"config" binding:"required"`
	Enabled bool                   `json:"enabled"`
}

// UpdateAlertChannelRequest 更新告警渠道请求
type UpdateAlertChannelRequest struct {
	Name    string                 `json:"name"`
	Config  map[string]interface{} `json:"config"`
	Enabled *bool                  `json:"enabled"`
}

// AlertChannelResponse 告警渠道响应
type AlertChannelResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

// TestAlertChannelRequest 测试告警渠道请求
type TestAlertChannelRequest struct {
	Message string `json:"message" binding:"required"`
}

// AlertRule DTO

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description"`
	Conditions  []model.AlertCondition   `json:"conditions" binding:"required,min=1,dive"`
	Logic       string                   `json:"logic" binding:"required,oneof=AND OR"`
	Channels    []string                 `json:"channels" binding:"required,min=1"`
	Template    string                   `json:"template" binding:"required"`
	Cooldown    int                      `json:"cooldown" binding:"required,min=1"`
	Severity    string                   `json:"severity" binding:"required,oneof=low medium high critical"`
	Enabled     bool                     `json:"enabled"`
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Conditions  []model.AlertCondition   `json:"conditions"`
	Logic       string                   `json:"logic" binding:"omitempty,oneof=AND OR"`
	Channels    []string                 `json:"channels"`
	Template    string                   `json:"template"`
	Cooldown    *int                     `json:"cooldown"`
	Severity    string                   `json:"severity" binding:"omitempty,oneof=low medium high critical"`
	Enabled     *bool                    `json:"enabled"`
}

// AlertRuleResponse 告警规则响应
type AlertRuleResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Conditions  []model.AlertCondition   `json:"conditions"`
	Logic       string                   `json:"logic"`
	Channels    []string                 `json:"channels"`
	Template    string                   `json:"template"`
	Cooldown    int                      `json:"cooldown"`
	Severity    string                   `json:"severity"`
	Enabled     bool                     `json:"enabled"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   time.Time                `json:"updatedAt"`
	CreatedBy   string                   `json:"createdBy,omitempty"`
}

// AlertHistory DTO

// AlertHistoryResponse 告警历史响应
type AlertHistoryResponse struct {
	ID              string                 `json:"id"`
	RuleID          string                 `json:"ruleId"`
	RuleName        string                 `json:"ruleName"`
	Severity        string                 `json:"severity"`
	Message         string                 `json:"message"`
	Details         map[string]interface{} `json:"details"`
	Channels        []string               `json:"channels"`
	Status          string                 `json:"status"`
	ErrorMessage    string                 `json:"errorMessage,omitempty"`
	TriggeredAt     time.Time              `json:"triggeredAt"`
	SentAt          *time.Time             `json:"sentAt,omitempty"`
	AcknowledgedAt  *time.Time             `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy  string                 `json:"acknowledgedBy,omitempty"`
}

// GetAlertHistoryRequest 获取告警历史请求
type GetAlertHistoryRequest struct {
	RuleID    string    `form:"ruleId"`
	Severity  string    `form:"severity" binding:"omitempty,oneof=low medium high critical"`
	Status    string    `form:"status" binding:"omitempty,oneof=pending sent failed acknowledged"`
	StartTime time.Time `form:"startTime"`
	EndTime   time.Time `form:"endTime"`
	Page      int       `form:"page" binding:"min=1"`
	PageSize  int       `form:"pageSize" binding:"min=1,max=100"`
}

// AcknowledgeAlertRequest 确认告警请求
type AcknowledgeAlertRequest struct {
	Comment string `json:"comment"`
}

// AlertTemplate DTO

// CreateAlertTemplateRequest 创建告警模板请求
type CreateAlertTemplateRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Content     string   `json:"content" binding:"required"`
	Variables   []string `json:"variables"`
}

// UpdateAlertTemplateRequest 更新告警模板请求
type UpdateAlertTemplateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Variables   []string `json:"variables"`
}

// AlertTemplateResponse 告警模板响应
type AlertTemplateResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Variables   []string  `json:"variables"`
	IsBuiltIn   bool      `json:"isBuiltIn"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   string    `json:"createdBy,omitempty"`
}

// Alert Statistics DTO

// AlertStatisticsResponse 告警统计响应
type AlertStatisticsResponse struct {
	TotalAlerts       int64                     `json:"totalAlerts"`
	AlertsBySeverity  map[string]int64          `json:"alertsBySeverity"`
	AlertsByStatus    map[string]int64          `json:"alertsByStatus"`
	TopAlertRules     []TopAlertRule            `json:"topAlertRules"`
	RecentAlerts      []AlertHistoryResponse    `json:"recentAlerts"`
}

// TopAlertRule Top 告警规则
type TopAlertRule struct {
	RuleID    string `json:"ruleId"`
	RuleName  string `json:"ruleName"`
	Count     int64  `json:"count"`
	Severity  string `json:"severity"`
}

// ChannelConfig 渠道配置接口 - 用于不同渠道的配置验证

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	URL     string            `json:"url" binding:"required,url"`
	Method  string            `json:"method" binding:"required,oneof=GET POST PUT"`
	Headers map[string]string `json:"headers"`
	Timeout int               `json:"timeout" binding:"min=1,max=300"` // 超时时间（秒）
}

// SlackConfig Slack 配置
type SlackConfig struct {
	WebhookURL string `json:"webhookUrl" binding:"required,url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"iconEmoji"`
}

// DiscordConfig Discord 配置
type DiscordConfig struct {
	WebhookURL string `json:"webhookUrl" binding:"required,url"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatarUrl"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `json:"webhookUrl" binding:"required,url"`
	Secret     string `json:"secret"`
	AtMobiles  []string `json:"atMobiles"`
	IsAtAll    bool   `json:"isAtAll"`
}

// WeComConfig 企业微信配置
type WeComConfig struct {
	WebhookURL string `json:"webhookUrl" binding:"required,url"`
	MentionedList []string `json:"mentionedList"`
	MentionedMobileList []string `json:"mentionedMobileList"`
}
