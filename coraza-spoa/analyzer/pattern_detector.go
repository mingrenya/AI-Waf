package analyzer

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AttackPatternDetector 攻击模式检测器 - 基于统计学习和聚类算法
type AttackPatternDetector struct {
	db     *mongo.Database
	logger zerolog.Logger
	
	// 特征提取器
	featureExtractor *FeatureExtractor
	
	// 配置
	minSamples       int     // 最小样本数
	anomalyThreshold float64 // 异常阈值
	timeWindowHours  int     // 时间窗口(小时)
}

// NewAttackPatternDetector 创建攻击模式检测器
func NewAttackPatternDetector(db *mongo.Database, logger zerolog.Logger) *AttackPatternDetector {
	return &AttackPatternDetector{
		db:               db,
		logger:           logger.With().Str("component", "pattern-detector").Logger(),
		featureExtractor: NewFeatureExtractor(),
		minSamples:       100,
		anomalyThreshold: 2.0,
		timeWindowHours:  24,
	}
}

// DetectPatterns 检测攻击模式
func (pd *AttackPatternDetector) DetectPatterns() ([]*model.AttackPattern, error) {
	pd.logger.Info().Msg("开始检测攻击模式")
	
	// 1. 从MongoDB获取最近的WAF日志
	logs, err := pd.fetchRecentLogs()
	if err != nil {
		return nil, fmt.Errorf("获取日志失败: %w", err)
	}
	
	if len(logs) < pd.minSamples {
		pd.logger.Warn().Int("count", len(logs)).Int("minSamples", pd.minSamples).
			Msg("样本数量不足,跳过分析")
		return nil, nil
	}
	
	pd.logger.Info().Int("count", len(logs)).Msg("获取到WAF日志")
	
	// 2. 提取特征
	features := make([]*AttackFeature, 0, len(logs))
	for _, log := range logs {
		feature, err := pd.featureExtractor.ExtractFeatures(log)
		if err != nil {
			pd.logger.Warn().Err(err).Str("requestId", log.RequestID).Msg("特征提取失败")
			continue
		}
		features = append(features, feature)
	}
	
	if len(features) == 0 {
		return nil, nil
	}
	
	pd.logger.Info().Int("count", len(features)).Msg("特征提取完成")
	
	// 3. 聚合相似特征
	aggregated := pd.featureExtractor.AggregateFeatures(features)
	pd.logger.Info().Int("count", len(aggregated)).Msg("特征聚合完成")
	
	// 4. 检测异常模式
	patterns := pd.detectAnomalies(aggregated)
	pd.logger.Info().Int("count", len(patterns)).Msg("异常检测完成")
	
	// 5. 保存检测到的模式
	for _, pattern := range patterns {
		if err := pd.savePattern(pattern); err != nil {
			pd.logger.Error().Err(err).Str("patternName", pattern.Name).Msg("保存模式失败")
		}
	}
	
	return patterns, nil
}

// fetchRecentLogs 获取最近的WAF日志
func (pd *AttackPatternDetector) fetchRecentLogs() ([]*model.WAFLog, error) {
	collection := pd.db.Collection("waf_log")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 查询时间范围
	startTime := time.Now().Add(-time.Duration(pd.timeWindowHours) * time.Hour)
	
	filter := bson.M{
		"createdAt": bson.M{"$gte": startTime},
		"ruleId":    bson.M{"$gt": 0}, // 只获取触发规则的日志
	}
	
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(10000) // 限制最多10000条
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var logs []*model.WAFLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	
	return logs, nil
}

// detectAnomalies 检测异常模式
func (pd *AttackPatternDetector) detectAnomalies(features []*AttackFeature) []*model.AttackPattern {
	patterns := make([]*model.AttackPattern, 0)
	
	// 按模式类型分组
	typeGroups := make(map[string][]*AttackFeature)
	for _, f := range features {
		typeGroups[f.PayloadType] = append(typeGroups[f.PayloadType], f)
	}
	
	// 对每种类型进行异常检测
	for payloadType, group := range typeGroups {
		if len(group) < pd.minSamples/10 { // 每种类型至少要有10个样本
			continue
		}
		
		// 按频率检测异常
		frequencyPatterns := pd.detectFrequencyAnomalies(group, payloadType)
		patterns = append(patterns, frequencyPatterns...)
		
		// 按IP模式检测异常
		ipPatterns := pd.detectIPPatterns(group, payloadType)
		patterns = append(patterns, ipPatterns...)
		
		// 按URL模式检测异常
		urlPatterns := pd.detectURLPatterns(group, payloadType)
		patterns = append(patterns, urlPatterns...)
	}
	
	return patterns
}

// detectFrequencyAnomalies 检测频率异常
func (pd *AttackPatternDetector) detectFrequencyAnomalies(features []*AttackFeature, payloadType string) []*model.AttackPattern {
	patterns := make([]*model.AttackPattern, 0)
	
	// 计算频率统计
	frequencies := make([]float64, len(features))
	for i, f := range features {
		frequencies[i] = f.Frequency
	}
	
	mean, stdDev := calculateStats(frequencies)
	if stdDev == 0 {
		return patterns
	}
	
	// 检测超过阈值的异常
	for _, f := range features {
		zScore := (f.Frequency - mean) / stdDev
		if zScore > pd.anomalyThreshold {
			pattern := &model.AttackPattern{
				Name:         fmt.Sprintf("高频%s攻击", payloadType),
				Description:  fmt.Sprintf("检测到异常高频的%s攻击模式, 频率为%.2f次/秒(均值%.2f)", payloadType, f.Frequency, mean),
				PatternType:  payloadType,
				Confidence:   math.Min(zScore/10.0, 1.0), // 归一化到0-1
				Severity:     calculateSeverity(f.Severity),
				URLPattern:   f.URLPattern,
				PathPattern:  f.PathPattern,
				IPPattern:    f.IPPattern,
				PayloadRegex: generatePayloadRegex(f.Payload),
				SampleCount:  f.RequestCount,
				Frequency:    f.Frequency,
				FirstSeen:    f.Timestamp.Add(-time.Duration(f.TimeWindowSec) * time.Second),
				LastSeen:     f.Timestamp,
				Status:       "active",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns
}

// detectIPPatterns 检测IP模式
func (pd *AttackPatternDetector) detectIPPatterns(features []*AttackFeature, payloadType string) []*model.AttackPattern {
	patterns := make([]*model.AttackPattern, 0)
	
	// 按IP模式分组并计数
	ipCounts := make(map[string]int)
	ipFeatures := make(map[string]*AttackFeature)
	
	for _, f := range features {
		if f.IPPattern != "" {
			ipCounts[f.IPPattern]++
			if ipFeatures[f.IPPattern] == nil {
				ipFeatures[f.IPPattern] = f
			}
		}
	}
	
	// 检测高频IP段
	threshold := len(features) / 10 // 超过10%的流量来自同一IP段
	for ipPattern, count := range ipCounts {
		if count >= threshold {
			f := ipFeatures[ipPattern]
			pattern := &model.AttackPattern{
				Name:         fmt.Sprintf("IP段%s的%s攻击", ipPattern, payloadType),
				Description:  fmt.Sprintf("检测到来自%s的集中%s攻击, 共%d次", ipPattern, payloadType, count),
				PatternType:  payloadType,
				Confidence:   float64(count) / float64(len(features)),
				Severity:     calculateSeverity(f.Severity),
				URLPattern:   f.URLPattern,
				PathPattern:  f.PathPattern,
				IPPattern:    ipPattern,
				PayloadRegex: generatePayloadRegex(f.Payload),
				SampleCount:  count,
				Frequency:    f.Frequency,
				FirstSeen:    f.Timestamp,
				LastSeen:     f.Timestamp,
				Status:       "active",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns
}

// detectURLPatterns 检测URL模式
func (pd *AttackPatternDetector) detectURLPatterns(features []*AttackFeature, payloadType string) []*model.AttackPattern {
	patterns := make([]*model.AttackPattern, 0)
	
	// 按URL模式分组
	urlCounts := make(map[string]int)
	urlFeatures := make(map[string]*AttackFeature)
	
	for _, f := range features {
		if f.PathPattern != "" {
			urlCounts[f.PathPattern]++
			if urlFeatures[f.PathPattern] == nil {
				urlFeatures[f.PathPattern] = f
			}
		}
	}
	
	// 检测高频URL
	threshold := len(features) / 20 // 超过5%的攻击针对同一路径
	for urlPattern, count := range urlCounts {
		if count >= threshold {
			f := urlFeatures[urlPattern]
			pattern := &model.AttackPattern{
				Name:         fmt.Sprintf("针对%s的%s攻击", urlPattern, payloadType),
				Description:  fmt.Sprintf("检测到针对%s路径的%s攻击模式, 共%d次", urlPattern, payloadType, count),
				PatternType:  payloadType,
				Confidence:   float64(count) / float64(len(features)),
				Severity:     calculateSeverity(f.Severity),
				URLPattern:   f.URLPattern,
				PathPattern:  urlPattern,
				IPPattern:    f.IPPattern,
				PayloadRegex: generatePayloadRegex(f.Payload),
				SampleCount:  count,
				Frequency:    f.Frequency,
				FirstSeen:    f.Timestamp,
				LastSeen:     f.Timestamp,
				Status:       "active",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns
}

// savePattern 保存攻击模式
func (pd *AttackPatternDetector) savePattern(pattern *model.AttackPattern) error {
	collection := pd.db.Collection("attack_patterns")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 检查是否已存在相似模式
	filter := bson.M{
		"patternType":  pattern.PatternType,
		"pathPattern":  pattern.PathPattern,
		"ipPattern":    pattern.IPPattern,
		"status":       "active",
	}
	
	var existing model.AttackPattern
	err := collection.FindOne(ctx, filter).Decode(&existing)
	
	if err == mongo.ErrNoDocuments {
		// 不存在，插入新模式
		_, err := collection.InsertOne(ctx, pattern)
		if err != nil {
			return err
		}
		pd.logger.Info().Str("patternName", pattern.Name).Msg("保存新攻击模式")
	} else if err == nil {
		// 存在，更新统计信息
		update := bson.M{
			"$set": bson.M{
				"lastSeen":    pattern.LastSeen,
				"sampleCount": existing.SampleCount + pattern.SampleCount,
				"updatedAt":   time.Now(),
			},
		}
		_, err := collection.UpdateByID(ctx, existing.ID, update)
		if err != nil {
			return err
		}
		pd.logger.Debug().Str("patternId", existing.ID.Hex()).Msg("更新现有模式")
	} else {
		return err
	}
	
	return nil
}

// calculateStats 计算统计值
func calculateStats(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	// 计算均值
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))
	
	// 计算标准差
	varSum := 0.0
	for _, v := range values {
		diff := v - mean
		varSum += diff * diff
	}
	
	if len(values) > 1 {
		stdDev = math.Sqrt(varSum / float64(len(values)-1))
	}
	
	return mean, stdDev
}

// calculateSeverity 计算严重程度
func calculateSeverity(severity int) string {
	switch {
	case severity >= 4:
		return "critical"
	case severity >= 3:
		return "high"
	case severity >= 2:
		return "medium"
	default:
		return "low"
	}
}

// generatePayloadRegex 生成载荷正则表达式
func generatePayloadRegex(payload string) string {
	// 简化版本：对特殊字符进行转义
	// 实际应用中应该更智能地提取关键模式
	if payload == "" {
		return ""
	}
	
	// 限制长度
	if len(payload) > 200 {
		payload = payload[:200]
	}
	
	// 简单转义特殊字符
	specialChars := []string{".", "*", "+", "?", "[", "]", "(", ")", "{", "}", "|", "^", "$", "\\"}
	for _, char := range specialChars {
		payload = strings.Replace(payload, char, "\\"+char, -1)
	}
	
	return payload
}

// GetPatternStats 获取模式统计
func (pd *AttackPatternDetector) GetPatternStats() (map[string]interface{}, error) {
	collection := pd.db.Collection("attack_patterns")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 统计活跃模式数量
	activeCount, err := collection.CountDocuments(ctx, bson.M{"status": "active"})
	if err != nil {
		return nil, err
	}
	
	// 按类型统计
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": "active"}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$patternType",
			"count": bson.M{"$sum": 1},
		}}},
	}
	
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	typeStats := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			Type  string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		typeStats[result.Type] = result.Count
	}
	
	return map[string]interface{}{
		"activePatterns": activeCount,
		"byType":         typeStats,
	}, nil
}

