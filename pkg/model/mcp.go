package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// MCPToolCall MCP工具调用记录
type MCPToolCall struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ToolName  string        `bson:"tool_name" json:"toolName"`
	Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
	Duration  int64         `bson:"duration" json:"duration"` // 毫秒
	Success   bool          `bson:"success" json:"success"`
	Error     string        `bson:"error,omitempty" json:"error,omitempty"`
	UserID    string        `bson:"user_id,omitempty" json:"userId,omitempty"`
	RequestID string        `bson:"request_id,omitempty" json:"requestId,omitempty"`
}

// GetCollectionName 返回集合名称
func (m MCPToolCall) GetCollectionName() string {
	return "mcp_tool_calls"
}

// AIRuleSuggestion AI生成的规则建议
type AIRuleSuggestion struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	PatternID       string        `bson:"pattern_id,omitempty" json:"patternId,omitempty"`
	PatternName     string        `bson:"pattern_name" json:"patternName"`
	RuleName        string        `bson:"rule_name" json:"ruleName"`
	RuleType        string        `bson:"rule_type" json:"ruleType"` // micro_rule, modsecurity
	Confidence      float64       `bson:"confidence" json:"confidence"`
	Severity        string        `bson:"severity" json:"severity"` // low, medium, high, critical
	Description     string        `bson:"description" json:"description"`
	Recommendation  string        `bson:"recommendation" json:"recommendation"`
	RuleContent     interface{}   `bson:"rule_content" json:"ruleContent"`
	Status          string        `bson:"status" json:"status"` // pending, approved, rejected, deployed
	CreatedAt       time.Time     `bson:"created_at" json:"createdAt"`
	ReviewedAt      *time.Time    `bson:"reviewed_at,omitempty" json:"reviewedAt,omitempty"`
	DeployedAt      *time.Time    `bson:"deployed_at,omitempty" json:"deployedAt,omitempty"`
	ReviewedBy      string        `bson:"reviewed_by,omitempty" json:"reviewedBy,omitempty"`
	DeployedRuleID  string        `bson:"deployed_rule_id,omitempty" json:"deployedRuleId,omitempty"`
}

// GetCollectionName 返回集合名称
func (m AIRuleSuggestion) GetCollectionName() string {
	return "ai_rule_suggestions"
}
