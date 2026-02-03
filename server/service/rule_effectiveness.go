package service

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RuleEffectivenessService 规则有效性评分服务接口
type RuleEffectivenessService interface {
	// 计算规则评分
	CalculateScore(ctx context.Context, ruleID string, period string) (*model.RuleEffectivenessScore, error)
	// 获取规则评分
	GetScore(ctx context.Context, ruleID string) (*model.RuleEffectivenessScore, error)
	// 获取所有评分
	ListScores(ctx context.Context, sortBy string, order int) ([]model.RuleEffectivenessScore, error)
	// 批量计算评分
	BatchCalculateScores(ctx context.Context, period string) error
}

type ruleEffectivenessServiceImpl struct {
	scoreCollection *mongo.Collection
	ruleCollection  *mongo.Collection
	logCollection   *mongo.Collection
	logger          zerolog.Logger
}

// NewRuleEffectivenessService 创建规则有效性评分服务
func NewRuleEffectivenessService(db *mongo.Database) RuleEffectivenessService {
	return &ruleEffectivenessServiceImpl{
		scoreCollection: db.Collection("rule_effectiveness_score"),
		ruleCollection:  db.Collection("micro_rule"),
		logCollection:   db.Collection("waf_log"),
		logger:          zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
}

// CalculateScore 计算规则评分
func (s *ruleEffectivenessServiceImpl) CalculateScore(ctx context.Context, ruleID string, period string) (*model.RuleEffectivenessScore, error) {
	objID, err := bson.ObjectIDFromHex(ruleID)
	if err != nil {
		return nil, fmt.Errorf("无效的规则ID: %w", err)
	}

	// 获取规则信息
	var rule model.MicroRule
	err = s.ruleCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("规则不存在")
		}
		return nil, fmt.Errorf("查询规则失败: %w", err)
	}

	// 计算时间范围
	endTime := time.Now()
	var startTime time.Time
	switch period {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
		period = "24h"
	}

	// 统计规则匹配次数
	matchCount, err := s.logCollection.CountDocuments(ctx, bson.M{
		"timestamp": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		// 假设日志中有rule_id字段关联到规则
		"rule_name": rule.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("统计匹配次数失败: %w", err)
	}

	// 统计阻止次数
	blockCount, err := s.logCollection.CountDocuments(ctx, bson.M{
		"timestamp": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"rule_name": rule.Name,
		"action":    "blocked",
	})
	if err != nil {
		return nil, fmt.Errorf("统计阻止次数失败: %w", err)
	}

	// 计算平均匹配时间（假设日志中有match_time字段）
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"timestamp": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"rule_name": rule.Name,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":     nil,
			"avgTime": bson.M{"$avg": "$match_time"},
			"maxTime": bson.M{"$max": "$match_time"},
			"minTime": bson.M{"$min": "$match_time"},
		}}},
	}

	cursor, err := s.logCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("统计匹配时间失败: %w", err)
	}
	defer cursor.Close(ctx)

	var avgMatchTime float64 = 0
	if cursor.Next(ctx) {
		var result struct {
			AvgTime float64 `bson:"avgTime"`
		}
		if err := cursor.Decode(&result); err == nil {
			avgMatchTime = result.AvgTime
		}
	}

	// 计算指标
	var blockRate float64
	if matchCount > 0 {
		blockRate = float64(blockCount) / float64(matchCount) * 100
	}

	// 假设误报和真阳的数据（实际应该从人工标注或其他来源获取）
	// 这里简化处理，根据阻止率估算
	falsePositive := int64(0)
	truePositive := blockCount

	var falsePositiveRate float64
	var truePositiveRate float64
	if matchCount > 0 {
		falsePositiveRate = float64(falsePositive) / float64(matchCount) * 100
		truePositiveRate = float64(truePositive) / float64(matchCount) * 100
	}

	// 评估性能影响
	var perfImpact string
	switch {
	case avgMatchTime < 1:
		perfImpact = "low"
	case avgMatchTime < 5:
		perfImpact = "medium"
	default:
		perfImpact = "high"
	}

	// 计算综合评分 (0-100)
	// 评分算法：考虑真阳率、误报率、阻止率和性能
	score := s.calculateCompositeScore(truePositiveRate, falsePositiveRate, blockRate, perfImpact)

	// 生成优化建议
	recommendation := s.generateRecommendation(score, truePositiveRate, falsePositiveRate, blockRate, perfImpact, matchCount)

	effectivenessScore := &model.RuleEffectivenessScore{
		RuleID:            objID,
		RuleName:          rule.Name,
		Score:             score,
		MatchCount:        matchCount,
		BlockCount:        blockCount,
		FalsePositive:     falsePositive,
		TruePositive:      truePositive,
		FalsePositiveRate: falsePositiveRate,
		TruePositiveRate:  truePositiveRate,
		BlockRate:         blockRate,
		AvgMatchTime:      avgMatchTime,
		PerformanceImpact: perfImpact,
		Recommendation:    recommendation,
		LastEvaluated:     time.Now(),
		EvaluationPeriod:  period,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// 保存或更新评分
	filter := bson.M{"rule_id": objID}
	update := bson.M{"$set": effectivenessScore}
	opts := options.UpdateOne().SetUpsert(true)

	_, err = s.scoreCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("保存评分失败: %w", err)
	}

	return effectivenessScore, nil
}

// calculateCompositeScore 计算综合评分
func (s *ruleEffectivenessServiceImpl) calculateCompositeScore(
	truePositiveRate, falsePositiveRate, blockRate float64,
	perfImpact string,
) float64 {
	// 权重分配
	const (
		truePositiveWeight  = 0.4 // 真阳率权重40%
		falsePositiveWeight = 0.3 // 误报率权重30% (负面)
		blockRateWeight     = 0.2 // 阻止率权重20%
		perfWeight          = 0.1 // 性能权重10%
	)

	// 性能得分
	var perfScore float64
	switch perfImpact {
	case "low":
		perfScore = 100
	case "medium":
		perfScore = 70
	case "high":
		perfScore = 40
	}

	// 综合评分 = 真阳率*权重 - 误报率*权重 + 阻止率*权重 + 性能得分*权重
	score := truePositiveRate*truePositiveWeight -
		falsePositiveRate*falsePositiveWeight +
		blockRate*blockRateWeight +
		perfScore*perfWeight

	// 确保评分在0-100范围内
	score = math.Max(0, math.Min(100, score))

	return math.Round(score*10) / 10 // 保留一位小数
}

// generateRecommendation 生成优化建议
func (s *ruleEffectivenessServiceImpl) generateRecommendation(
	score, truePositiveRate, falsePositiveRate, blockRate float64,
	perfImpact string, matchCount int64,
) string {
	if score >= 85 {
		return "规则效果优秀，保持当前配置"
	}

	recommendations := []string{}

	if truePositiveRate < 50 {
		recommendations = append(recommendations, "真阳率较低，建议检查规则条件是否过于宽松")
	}

	if falsePositiveRate > 10 {
		recommendations = append(recommendations, "误报率较高，建议优化规则条件以减少误报")
	}

	if blockRate < 30 {
		recommendations = append(recommendations, "阻止率较低，考虑提高规则优先级或调整匹配条件")
	}

	if perfImpact == "high" {
		recommendations = append(recommendations, "性能影响较大，建议优化正则表达式或匹配逻辑")
	}

	if matchCount == 0 {
		recommendations = append(recommendations, "规则未触发，考虑禁用或删除该规则")
	}

	if len(recommendations) == 0 {
		return "规则效果良好，可考虑根据实际情况微调"
	}

	result := "建议："
	for i, r := range recommendations {
		if i > 0 {
			result += "；"
		}
		result += r
	}

	return result
}

// GetScore 获取规则评分
func (s *ruleEffectivenessServiceImpl) GetScore(ctx context.Context, ruleID string) (*model.RuleEffectivenessScore, error) {
	objID, err := bson.ObjectIDFromHex(ruleID)
	if err != nil {
		return nil, fmt.Errorf("无效的规则ID: %w", err)
	}

	var score model.RuleEffectivenessScore
	err = s.scoreCollection.FindOne(ctx, bson.M{"rule_id": objID}).Decode(&score)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("评分不存在")
		}
		return nil, fmt.Errorf("查询评分失败: %w", err)
	}

	return &score, nil
}

// ListScores 获取所有评分
func (s *ruleEffectivenessServiceImpl) ListScores(ctx context.Context, sortBy string, order int) ([]model.RuleEffectivenessScore, error) {
	if sortBy == "" {
		sortBy = "score"
	}
	if order == 0 {
		order = -1 // 默认降序
	}

	opts := options.Find().SetSort(bson.D{{Key: sortBy, Value: order}})
	cursor, err := s.scoreCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("查询评分失败: %w", err)
	}
	defer cursor.Close(ctx)

	var scores []model.RuleEffectivenessScore
	if err = cursor.All(ctx, &scores); err != nil {
		return nil, fmt.Errorf("解析评分失败: %w", err)
	}

	return scores, nil
}

// BatchCalculateScores 批量计算评分
func (s *ruleEffectivenessServiceImpl) BatchCalculateScores(ctx context.Context, period string) error {
	// 获取所有启用的规则
	cursor, err := s.ruleCollection.Find(ctx, bson.M{"status": model.RuleEnabled})
	if err != nil {
		return fmt.Errorf("查询规则失败: %w", err)
	}
	defer cursor.Close(ctx)

	var rules []model.MicroRule
	if err = cursor.All(ctx, &rules); err != nil {
		return fmt.Errorf("解析规则失败: %w", err)
	}

	// 逐个计算评分
	for _, rule := range rules {
		_, err := s.CalculateScore(ctx, rule.ID.Hex(), period)
		if err != nil {
			// 记录错误但继续处理其他规则
			s.logger.Error().Err(err).Str("rule_name", rule.Name).Msg("计算规则评分失败")
			continue
		}
	}

	return nil
}
