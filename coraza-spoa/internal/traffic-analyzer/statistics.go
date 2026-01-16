package trafficanalyzer

import (
	"context"
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// StatisticsCollector 统计数据收集器
type StatisticsCollector struct {
	db     *mongo.Database
	logger zerolog.Logger

	// 内存中的统计数据 (最近1小时)
	recentEvents []TrafficEvent
	eventsMutex  sync.RWMutex

	// 聚合统计
	visitCount  int64
	attackCount int64
	errorCount  int64
	statsMutex  sync.RWMutex
}

// NewStatisticsCollector 创建统计收集器
func NewStatisticsCollector(db *mongo.Database, logger zerolog.Logger) *StatisticsCollector {
	return &StatisticsCollector{
		db:           db,
		logger:       logger.With().Str("component", "statistics").Logger(),
		recentEvents: make([]TrafficEvent, 0, 10000),
	}
}

// Record 记录流量事件
func (sc *StatisticsCollector) Record(event *TrafficEvent) {
	// 更新内存统计
	sc.statsMutex.Lock()
	switch event.Type {
	case "visit":
		sc.visitCount++
	case "attack":
		sc.attackCount++
	case "error":
		sc.errorCount++
	}
	sc.statsMutex.Unlock()

	// 保存到内存(用于实时分析)
	sc.eventsMutex.Lock()
	sc.recentEvents = append(sc.recentEvents, *event)
	
	// 保持最近1小时的数据
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	// 清理旧数据
	for len(sc.recentEvents) > 0 && sc.recentEvents[0].Timestamp.Before(oneHourAgo) {
		sc.recentEvents = sc.recentEvents[1:]
	}
	sc.eventsMutex.Unlock()

	// 异步保存到数据库
	go sc.persistToDatabase(event)
}

// persistToDatabase 持久化到数据库
func (sc *StatisticsCollector) persistToDatabase(event *TrafficEvent) {
	// 每分钟聚合一次数据
	timestamp := event.Timestamp.Truncate(time.Minute)

	// 构建流量模式记录
	pattern := model.TrafficPattern{
		Timestamp: timestamp,
		Type:      event.Type,
	}
	// 设置Metrics字段
	pattern.Metrics.RequestRate = 1.0 / 60.0 // 每分钟的请求率
	pattern.Metrics.UniqueIPs = 1
	if event.IsBlocked {
		pattern.Metrics.BlockedCount = 1
	} else {
		pattern.Metrics.PassedCount = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用 upsert 更新或插入
	filter := bson.M{
		"timestamp": timestamp,
		"type":      event.Type,
	}

	update := bson.M{
		"$inc": bson.M{
			"metrics.requestRate": 1.0 / 60.0,
		},
		"$setOnInsert": pattern,
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := sc.db.Collection("traffic_patterns").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		sc.logger.Error().Err(err).Msg("Failed to persist traffic pattern")
	}
}

// GetRecentPatterns 获取最近的流量模式
func (sc *StatisticsCollector) GetRecentPatterns(duration time.Duration) []model.TrafficPattern {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now().Add(-duration)
	filter := bson.M{
		"timestamp": bson.M{"$gte": startTime},
	}

	cursor, err := sc.db.Collection("traffic_patterns").Find(ctx, filter)
	if err != nil {
		sc.logger.Error().Err(err).Msg("Failed to query traffic patterns")
		return nil
	}
	defer cursor.Close(ctx)

	var patterns []model.TrafficPattern
	if err := cursor.All(ctx, &patterns); err != nil {
		sc.logger.Error().Err(err).Msg("Failed to decode traffic patterns")
		return nil
	}

	return patterns
}

// GetCurrentMetrics 获取当前指标 (最近5分钟)
func (sc *StatisticsCollector) GetCurrentMetrics() map[string]float64 {
	sc.eventsMutex.RLock()
	defer sc.eventsMutex.RUnlock()

	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)

	var visitCount, attackCount, errorCount int
	var totalResponseTime time.Duration

	for i := len(sc.recentEvents) - 1; i >= 0; i-- {
		event := sc.recentEvents[i]
		if event.Timestamp.Before(fiveMinutesAgo) {
			break
		}

		switch event.Type {
		case "visit":
			visitCount++
		case "attack":
			attackCount++
		case "error":
			errorCount++
		}
		totalResponseTime += event.ResponseTime
	}

	total := visitCount + attackCount + errorCount
	avgResponseTime := float64(0)
	if total > 0 {
		avgResponseTime = float64(totalResponseTime.Milliseconds()) / float64(total)
	}

	return map[string]float64{
		"visit":        float64(visitCount) / 300.0,  // 每秒请求数
		"attack":       float64(attackCount) / 300.0,
		"error":        float64(errorCount) / 300.0,
		"responseTime": avgResponseTime,
	}
}

// GetStats 获取统计摘要
func (sc *StatisticsCollector) GetStats() map[string]interface{} {
	sc.statsMutex.RLock()
	defer sc.statsMutex.RUnlock()

	sc.eventsMutex.RLock()
	recentCount := len(sc.recentEvents)
	sc.eventsMutex.RUnlock()

	return map[string]interface{}{
		"totalVisits":  sc.visitCount,
		"totalAttacks": sc.attackCount,
		"totalErrors":  sc.errorCount,
		"recentEvents": recentCount,
	}
}
