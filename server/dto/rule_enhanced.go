package dto

import (
	"encoding/json"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
)

// ============== 规则模板相关 ==============

// RuleTemplateListRequest 规则模板列表请求
type RuleTemplateListRequest struct {
	Category string `form:"category"` // 分类过滤
	Severity string `form:"severity"` // 严重等级过滤
}

// RuleTemplateResponse 规则模板响应
type RuleTemplateResponse struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Severity    string          `json:"severity"`
	RuleType    string          `json:"rule_type"`
	Priority    int             `json:"priority"`
	Tags        []string        `json:"tags"`
	Condition   json.RawMessage `json:"condition"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// RuleTemplateListResponse 规则模板列表响应
type RuleTemplateListResponse struct {
	Total int64                  `json:"total"`
	Items []RuleTemplateResponse `json:"items"`
}

// CreateRuleFromTemplateRequest 从模板创建规则请求
type CreateRuleFromTemplateRequest struct {
	TemplateID string `json:"template_id" binding:"required"`
	CustomName string `json:"custom_name"` // 可选，自定义规则名称
}

// ============== 规则有效性评分相关 ==============

// RuleEffectivenessScoreResponse 规则有效性评分响应
type RuleEffectivenessScoreResponse struct {
	ID                string  `json:"id,omitempty"`
	RuleID            string  `json:"rule_id"`
	RuleName          string  `json:"rule_name"`
	Score             float64 `json:"score"`
	MatchCount        int64   `json:"match_count"`
	BlockCount        int64   `json:"block_count"`
	FalsePositive     int64   `json:"false_positive"`
	TruePositive      int64   `json:"true_positive"`
	FalsePositiveRate float64 `json:"false_positive_rate"`
	TruePositiveRate  float64 `json:"true_positive_rate"`
	BlockRate         float64 `json:"block_rate"`
	AvgMatchTime      float64 `json:"avg_match_time"`
	PerformanceImpact string  `json:"performance_impact"`
	Recommendation    string  `json:"recommendation"`
	LastEvaluated     string  `json:"last_evaluated"`
	EvaluationPeriod  string  `json:"evaluation_period"`
}

// RuleEffectivenessScoreListResponse 规则有效性评分列表响应
type RuleEffectivenessScoreListResponse struct {
	Total int64                            `json:"total"`
	Items []RuleEffectivenessScoreResponse `json:"items"`
}

// CalculateScoreRequest 计算评分请求
type CalculateScoreRequest struct {
	RuleID string `json:"rule_id" binding:"required"`
	Period string `json:"period" binding:"required,oneof=24h 7d 30d"` // 评估周期
}

// BatchCalculateScoresRequest 批量计算评分请求
type BatchCalculateScoresRequest struct {
	Period string `json:"period" binding:"required,oneof=24h 7d 30d"`
}

// ListScoresRequest 获取评分列表请求
type ListScoresRequest struct {
	SortBy string `form:"sortBy"` // 排序字段: score, match_count, block_rate等
	Order  int    `form:"order"`  // 排序方向: 1升序, -1降序
}

// ============== 保护配置文件相关 ==============

// ProtectionProfileResponse 保护配置文件响应
type ProtectionProfileResponse struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Level       string   `json:"level"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	TemplateIDs []string `json:"template_ids"`
	IsDefault   bool     `json:"is_default"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// ProtectionProfileListResponse 保护配置文件列表响应
type ProtectionProfileListResponse struct {
	Total int64                       `json:"total"`
	Items []ProtectionProfileResponse `json:"items"`
}

// ApplyProfileRequest 应用配置文件请求
type ApplyProfileRequest struct {
	ProfileID string `json:"profile_id" binding:"required"`
}

// ApplyProfileResponse 应用配置文件响应
type ApplyProfileResponse struct {
	CreatedCount int      `json:"created_count"`
	Message      string   `json:"message"`
	RuleNames    []string `json:"rule_names,omitempty"` // 创建的规则名称列表
}

// ============== 辅助函数 ==============

// ConvertRuleTemplateToResponse 转换规则模板为响应格式
func ConvertRuleTemplateToResponse(template *model.RuleTemplate) RuleTemplateResponse {
	// 将 bson.Raw 转换为 JSON，需要先解码再编码
	var conditionData interface{}
	var conditionJSON json.RawMessage
	if len(template.Condition) > 0 {
		if err := bson.Unmarshal(template.Condition, &conditionData); err == nil {
			if jsonBytes, err := json.Marshal(conditionData); err == nil {
				conditionJSON = jsonBytes
			}
		}
	}

	return RuleTemplateResponse{
		ID:          template.ID.Hex(),
		Name:        template.Name,
		Category:    template.Category,
		Description: template.Description,
		Severity:    template.Severity,
		RuleType:    string(template.RuleType),
		Priority:    template.Priority,
		Tags:        template.Tags,
		Condition:   conditionJSON,
		CreatedAt:   template.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   template.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ConvertRuleEffectivenessScoreToResponse 转换规则有效性评分为响应格式
func ConvertRuleEffectivenessScoreToResponse(score *model.RuleEffectivenessScore) RuleEffectivenessScoreResponse {
	return RuleEffectivenessScoreResponse{
		ID:                score.ID.Hex(),
		RuleID:            score.RuleID.Hex(),
		RuleName:          score.RuleName,
		Score:             score.Score,
		MatchCount:        score.MatchCount,
		BlockCount:        score.BlockCount,
		FalsePositive:     score.FalsePositive,
		TruePositive:      score.TruePositive,
		FalsePositiveRate: score.FalsePositiveRate,
		TruePositiveRate:  score.TruePositiveRate,
		BlockRate:         score.BlockRate,
		AvgMatchTime:      score.AvgMatchTime,
		PerformanceImpact: score.PerformanceImpact,
		Recommendation:    score.Recommendation,
		LastEvaluated:     score.LastEvaluated.Format("2006-01-02 15:04:05"),
		EvaluationPeriod:  score.EvaluationPeriod,
	}
}

// ConvertProtectionProfileToResponse 转换保护配置文件为响应格式
func ConvertProtectionProfileToResponse(profile *model.ProtectionProfile) ProtectionProfileResponse {
	templateIDs := make([]string, len(profile.TemplateIDs))
	for i, id := range profile.TemplateIDs {
		templateIDs[i] = id.Hex()
	}

	return ProtectionProfileResponse{
		ID:          profile.ID.Hex(),
		Name:        profile.Name,
		Level:       profile.Level,
		Description: profile.Description,
		Categories:  profile.Categories,
		TemplateIDs: templateIDs,
		IsDefault:   profile.IsDefault,
		CreatedAt:   profile.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   profile.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
