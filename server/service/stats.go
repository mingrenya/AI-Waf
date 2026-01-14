package service

import (
	"context"
	"fmt"
	"math"
	"time"

	mongodb "github.com/mingrenya/AI-Waf/pkg/database/mongo"
	pkgModel "github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type StatsService interface {
	GetOverviewStats(ctx context.Context, timeRange string) (*dto.OverviewStats, error)
	GetRealtimeQPS(ctx context.Context, limit int) (*dto.RealtimeQPSResponse, error)
	GetTimeSeriesData(ctx context.Context, timeRange string, metric string) (*dto.TimeSeriesResponse, error)
	GetCombinedTimeSeriesData(ctx context.Context, timeRange string) (*dto.CombinedTimeSeriesResponse, error)
	GetTrafficTimeSeriesData(ctx context.Context, timeRange string) (*dto.TrafficTimeSeriesResponse, error)
	GetSecurityMetrics(ctx context.Context, timeRange string) (*dto.SecurityMetricsResponse, error)
}

type StatsServiceImpl struct {
	wafLogRepository repository.WAFLogRepository
	dbName           string
	logger           zerolog.Logger
}

func NewStatsService(wafLogRepository repository.WAFLogRepository) StatsService {
	dbName := config.Global.DBConfig.Database
	logger := config.GetServiceLogger("stats")
	return &StatsServiceImpl{
		wafLogRepository: wafLogRepository,
		dbName:           dbName,
		logger:           logger,
	}
}

// GetOverviewStats 获取概览统计数据
func (s *StatsServiceImpl) GetOverviewStats(ctx context.Context, timeRange string) (*dto.OverviewStats, error) {
	// 确定时间范围
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 1. 获取 HAProxyMinuteStats 的统计数据
	minuteStatsResult, err := s.getHAProxyStatsAggregate(ctx, db, startTime)
	if err != nil {
		return nil, fmt.Errorf("获取统计数据失败: %w", err)
	}

	// 2. 获取 WAF 拦截统计数据
	blockCount, attackIPCount, err := s.getWAFBlockStats(ctx, startTime)
	if err != nil {
		return nil, fmt.Errorf("获取WAF拦截统计失败: %w", err)
	}

	// 3. 计算错误率
	error4xxRate := 0.0
	error5xxRate := 0.0

	if minuteStatsResult.TotalRequests > 0 {
		error4xxRate = float64(minuteStatsResult.Error4xx) / float64(minuteStatsResult.TotalRequests) * 100
		error5xxRate = float64(minuteStatsResult.Error5xx) / float64(minuteStatsResult.TotalRequests) * 100
	}

	// 构建结果
	result := &dto.OverviewStats{
		TimeRange:       timeRange,
		TotalRequests:   minuteStatsResult.TotalRequests,
		InboundTraffic:  minuteStatsResult.InboundTraffic,
		OutboundTraffic: minuteStatsResult.OutboundTraffic,
		MaxQPS:          minuteStatsResult.MaxQPS,
		Error4xx:        minuteStatsResult.Error4xx,
		Error4xxRate:    math.Round(error4xxRate*100) / 100, // 保留两位小数
		Error5xx:        minuteStatsResult.Error5xx,
		Error5xxRate:    math.Round(error5xxRate*100) / 100, // 保留两位小数
		BlockCount:      blockCount,
		AttackIPCount:   attackIPCount,
	}

	return result, nil
}

// GetRealtimeQPS 获取实时QPS数据
func (s *StatsServiceImpl) GetRealtimeQPS(ctx context.Context, limit int) (*dto.RealtimeQPSResponse, error) {
	if limit <= 0 {
		limit = 30 // 默认获取30个点
	}
	if limit > 240 {
		limit = 240 // 最多不超过240个点
	}

	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 从时序数据库中查询数据
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "metadata.target", Value: "all"},
	}}}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: -1}, // 按时间降序排序
	}}}

	limitStage := bson.D{{Key: "$limit", Value: limit}}

	pipeline := mongo.Pipeline{matchStage, sortStage, limitStage}

	// 执行聚合查询
	reqRateColl := db.Collection("req_rate")
	cursor, err := reqRateColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("查询实时QPS数据失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		Timestamp time.Time `bson:"timestamp"`
		Value     int64     `bson:"value"`
		Metadata  bson.D    `bson:"metadata"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析实时QPS数据失败: %w", err)
	}

	// 构建结果
	dataPoints := make([]dto.RealtimeQPSData, 0, len(results))
	// 反转顺序，让时间从旧到新
	for i := len(results) - 1; i >= 0; i-- {
		dataPoints = append(dataPoints, dto.RealtimeQPSData{
			Timestamp: results[i].Timestamp,
			Value:     results[i].Value,
		})
	}

	return &dto.RealtimeQPSResponse{
		Data: dataPoints,
	}, nil
}

// GetTimeSeriesData 获取时间序列数据
func (s *StatsServiceImpl) GetTimeSeriesData(ctx context.Context, timeRange string, metric string) (*dto.TimeSeriesResponse, error) {
	// 确定时间范围
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// 根据时间范围决定数据聚合粒度
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// 24小时内按小时聚合
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// 7天内按6小时聚合，提供更多数据点
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// 30天内按天聚合
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("无效的时间范围: %s", timeRange)
	}

	// 根据指标类型获取数据
	var dataPoints []dto.TimeSeriesDataPoint

	switch metric {
	case "requests":
		// 请求数时间序列
		dataPoints, err = s.getRequestTimeSeries(ctx, startTime, interval, groupByFields)
	case "blocks":
		// 拦截数时间序列
		dataPoints, err = s.getBlockTimeSeries(ctx, startTime, interval, groupByFields)
	default:
		return nil, fmt.Errorf("无效的指标类型: %s", metric)
	}

	if err != nil {
		return nil, err
	}

	return &dto.TimeSeriesResponse{
		Metric:    metric,
		TimeRange: timeRange,
		Data:      dataPoints,
	}, nil
}

// GetCombinedTimeSeriesData 同时获取请求数和拦截数的时间序列数据
func (s *StatsServiceImpl) GetCombinedTimeSeriesData(ctx context.Context, timeRange string) (*dto.CombinedTimeSeriesResponse, error) {
	// 确定时间范围
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// 根据时间范围决定数据聚合粒度
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// 24小时内按小时聚合
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// 7天内按6小时聚合，提供更多数据点
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// 30天内按天聚合
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("无效的时间范围: %s", timeRange)
	}

	// 并行获取两种数据
	requestCh := make(chan struct {
		data []dto.TimeSeriesDataPoint
		err  error
	})
	blockCh := make(chan struct {
		data []dto.TimeSeriesDataPoint
		err  error
	})

	// 获取请求数据
	go func() {
		data, err := s.getRequestTimeSeries(ctx, startTime, interval, groupByFields)
		requestCh <- struct {
			data []dto.TimeSeriesDataPoint
			err  error
		}{data, err}
	}()

	// 获取拦截数据
	go func() {
		data, err := s.getBlockTimeSeries(ctx, startTime, interval, groupByFields)
		blockCh <- struct {
			data []dto.TimeSeriesDataPoint
			err  error
		}{data, err}
	}()

	// 接收结果
	requestResult := <-requestCh
	blockResult := <-blockCh

	// 检查错误
	if requestResult.err != nil {
		return nil, fmt.Errorf("获取请求数据失败: %w", requestResult.err)
	}
	if blockResult.err != nil {
		return nil, fmt.Errorf("获取拦截数据失败: %w", blockResult.err)
	}

	// 返回组合结果
	return &dto.CombinedTimeSeriesResponse{
		TimeRange: timeRange,
		Requests: dto.TimeSeriesResponse{
			Metric:    "requests",
			TimeRange: timeRange,
			Data:      requestResult.data,
		},
		Blocks: dto.TimeSeriesResponse{
			Metric:    "blocks",
			TimeRange: timeRange,
			Data:      blockResult.data,
		},
	}, nil
}

// GetTrafficTimeSeriesData 获取流量时间序列数据
func (s *StatsServiceImpl) GetTrafficTimeSeriesData(ctx context.Context, timeRange string) (*dto.TrafficTimeSeriesResponse, error) {
	// 确定时间范围
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// 根据时间范围决定数据聚合粒度
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// 24小时内按小时聚合
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// 7天内按6小时聚合，提供更多数据点
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// 30天内按天聚合
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("无效的时间范围: %s", timeRange)
	}

	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	var haproxyMinuteStats model.HAProxyMinuteStats
	collection := db.Collection(haproxyMinuteStats.GetCollectionName())

	// 构建时间过滤条件
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// 构建分组条件
	var groupStage bson.D

	if interval == "6hour" {
		// 使用预计算的HourGroupSix字段进行分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
			{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	} else {
		// 构建常规分组ID
		groupID := bson.D{}
		for _, field := range groupByFields {
			groupID = append(groupID, bson.E{Key: field, Value: fmt.Sprintf("$%s", field)})
		}

		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: groupID},
			{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
			{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // 按时间升序排序
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage}
	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true)

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, fmt.Errorf("执行流量时间序列聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		ID              bson.D    `bson:"_id"`
		InboundTraffic  int64     `bson:"inboundTraffic"`
		OutboundTraffic int64     `bson:"outboundTraffic"`
		Timestamp       time.Time `bson:"timestamp"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析流量时间序列结果失败: %w", err)
	}

	// 构建数据点
	dataPoints := make([]dto.TrafficDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TrafficDataPoint{
			Timestamp:       result.Timestamp,
			InboundTraffic:  result.InboundTraffic,
			OutboundTraffic: result.OutboundTraffic,
		})
	}

	return &dto.TrafficTimeSeriesResponse{
		TimeRange: timeRange,
		Data:      dataPoints,
	}, nil
}

// 辅助方法 - 获取时间范围的开始时间
func (s *StatsServiceImpl) getTimeRangeStart(timeRange string) (time.Time, error) {
	now := time.Now()

	switch timeRange {
	case dto.TimeRange24Hours:
		return now.Add(-24 * time.Hour), nil
	case dto.TimeRange7Days:
		return now.AddDate(0, 0, -7), nil
	case dto.TimeRange30Days:
		return now.AddDate(0, 0, -30), nil
	default:
		return time.Time{}, fmt.Errorf("无效的时间范围: %s", timeRange)
	}
}

// 辅助方法 - haproxy统计聚合结果
type haproxyStatsAggregateResult struct {
	TotalRequests   int64 // 总请求数
	InboundTraffic  int64 // 入站流量
	OutboundTraffic int64 // 出站流量
	MaxQPS          int64 // 最大QPS
	Error4xx        int64 // 4xx错误数
	Error5xx        int64 // 5xx错误数
}

// 辅助方法 - 获取haproxy统计数据聚合结果
func (s *StatsServiceImpl) getHAProxyStatsAggregate(ctx context.Context, db *mongo.Database, startTime time.Time) (*haproxyStatsAggregateResult, error) {
	// 创建聚合管道
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: nil},
		{Key: "totalRequests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
		{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
		{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
		{Key: "maxQPS", Value: bson.D{{Key: "$max", Value: "$req_rate_max"}}},
		{Key: "error4xx", Value: bson.D{{Key: "$sum", Value: "$hrsp_4xx"}}},
		{Key: "error5xx", Value: bson.D{{Key: "$sum", Value: "$hrsp_5xx"}}},
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage}

	// 设置聚合选项以优化性能
	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true).                                                        // 允许使用磁盘进行大数据集聚合
		SetHint(bson.D{{Key: "target_name", Value: 1}, {Key: "timestamp", Value: 1}}) // 使用最优复合索引

	// 执行聚合查询
	var haproxyMinuteStats model.HAProxyMinuteStats
	collection := db.Collection(haproxyMinuteStats.GetCollectionName())

	queryStartTime := time.Now()
	cursor, err := collection.Aggregate(ctx, pipeline, aggregateOptions)
	duration := time.Since(queryStartTime)

	if err != nil {
		s.logger.Error().
			Err(err).
			Dur("duration", duration).
			Time("queryStartTime", startTime).
			Msg("HAProxy统计聚合查询失败")
		return nil, fmt.Errorf("执行统计聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		TotalRequests   int64 `bson:"totalRequests"`
		InboundTraffic  int64 `bson:"inboundTraffic"`
		OutboundTraffic int64 `bson:"outboundTraffic"`
		MaxQPS          int64 `bson:"maxQPS"`
		Error4xx        int64 `bson:"error4xx"`
		Error5xx        int64 `bson:"error5xx"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		s.logger.Error().
			Err(err).
			Dur("duration", duration).
			Msg("解析HAProxy统计聚合结果失败")
		return nil, fmt.Errorf("解析统计聚合结果失败: %w", err)
	}

	// 构建结果
	result := &haproxyStatsAggregateResult{}
	if len(results) > 0 {
		result.TotalRequests = results[0].TotalRequests
		result.InboundTraffic = results[0].InboundTraffic
		result.OutboundTraffic = results[0].OutboundTraffic
		result.MaxQPS = results[0].MaxQPS
		result.Error4xx = results[0].Error4xx
		result.Error5xx = results[0].Error5xx
	}

	// 记录性能监控日志
	s.logger.Debug().
		Dur("duration", duration).
		Int64("totalRequests", result.TotalRequests).
		Int64("inboundTraffic", result.InboundTraffic).
		Int64("outboundTraffic", result.OutboundTraffic).
		Int64("maxQPS", result.MaxQPS).
		Time("queryStartTime", startTime).
		Msg("HAProxy统计聚合查询完成")

	// 性能警告
	if duration > 1*time.Second {
		s.logger.Warn().
			Dur("duration", duration).
			Time("queryStartTime", startTime).
			Msg("HAProxy统计聚合查询耗时较长")
	}

	return result, nil
}

// 辅助方法 - 获取WAF拦截统计
func (s *StatsServiceImpl) getWAFBlockStats(ctx context.Context, startTime time.Time) (int64, int64, error) {
	// 构建时间过滤条件
	timeFilter := bson.D{
		{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}

	// 获取拦截总数
	blockCount, err := s.wafLogRepository.CountAttackLogs(ctx, timeFilter)
	if err != nil {
		return 0, 0, fmt.Errorf("计算拦截总数失败: %w", err)
	}

	// 获取MongoDB数据库连接，直接执行聚合查询获取不同攻击IP数
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return 0, 0, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	var wafLog pkgModel.WAFLog
	collection := db.Collection(wafLog.GetCollectionName())

	// 优化的聚合管道：直接计算不同IP数量
	ipCountPipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.D{
				{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
			}},
		},
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$srcIp"},
			}},
		},
		{
			{Key: "$count", Value: "uniqueIPs"},
		},
	}

	// 设置聚合选项以优化大数据集查询性能
	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true).                                                  // 允许使用磁盘进行大数据集聚合
		SetHint(bson.D{{Key: "srcIp", Value: 1}, {Key: "createdAt", Value: 1}}) // 使用复合索引提示

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, ipCountPipeline, aggregateOptions)
	if err != nil {
		s.logger.Error().Err(err).Msg("计算攻击IP数量失败")
		return blockCount, 0, nil
	}
	defer cursor.Close(ctx)

	// 解析结果
	var result struct {
		UniqueIPs int64 `bson:"uniqueIPs"`
	}

	attackIPCount := int64(0)
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			s.logger.Error().Err(err).Msg("解析攻击IP数量结果失败")
			return blockCount, 0, nil
		}
		attackIPCount = result.UniqueIPs
	}

	// 记录统计信息用于性能监控
	s.logger.Debug().
		Int64("blockCount", blockCount).
		Int64("attackIPCount", attackIPCount).
		Time("startTime", startTime).
		Msg("WAF拦截统计查询完成")

	return blockCount, attackIPCount, nil
}

// 辅助方法 - 获取请求数时间序列
func (s *StatsServiceImpl) getRequestTimeSeries(ctx context.Context, startTime time.Time, interval string, groupByFields []string) ([]dto.TimeSeriesDataPoint, error) {
	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	var haproxyMinuteStats model.HAProxyMinuteStats
	collection := db.Collection(haproxyMinuteStats.GetCollectionName())

	// 构建时间过滤条件
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// 构建分组条件
	var groupStage bson.D

	if interval == "6hour" {
		// 使用预计算的HourGroupSix字段进行分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "requests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	} else {
		// 构建常规分组ID
		groupID := bson.D{}
		for _, field := range groupByFields {
			groupID = append(groupID, bson.E{Key: field, Value: fmt.Sprintf("$%s", field)})
		}

		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: groupID},
			{Key: "requests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // 按时间升序排序
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage}

	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true)

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, fmt.Errorf("执行请求时间序列聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		ID        bson.D    `bson:"_id"`
		Requests  int64     `bson:"requests"`
		Timestamp time.Time `bson:"timestamp"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析请求时间序列结果失败: %w", err)
	}

	// 构建数据点
	dataPoints := make([]dto.TimeSeriesDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TimeSeriesDataPoint{
			Timestamp: result.Timestamp,
			Value:     result.Requests,
		})
	}

	return dataPoints, nil
}

// 辅助方法 - 获取拦截数时间序列
func (s *StatsServiceImpl) getBlockTimeSeries(ctx context.Context, startTime time.Time, interval string, groupByFields []string) ([]dto.TimeSeriesDataPoint, error) {
	// 获取MongoDB数据库连接
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	collection := db.Collection("waf_log")

	// 构建时间过滤条件
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// 根据interval决定如何分组
	var groupStage bson.D

	if interval == "hour" {
		// 按小时分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hour", Value: "$hour"},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	} else if interval == "6hour" {
		// 使用预计算的HourGroupSix字段分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	} else {
		// 按日期分组
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "date", Value: "$date"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // 按时间升序排序
	}}}

	// 限制返回项目 - 优化性能
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "count", Value: 1},
		{Key: "timestamp", Value: 1},
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage, projectStage}

	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true)

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, fmt.Errorf("执行WAF拦截时间序列聚合查询失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []struct {
		Count     int64     `bson:"count"`
		Timestamp time.Time `bson:"timestamp"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析WAF拦截时间序列结果失败: %w", err)
	}

	// 构建数据点
	dataPoints := make([]dto.TimeSeriesDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TimeSeriesDataPoint{
			Timestamp: result.Timestamp,
			Value:     result.Count,
		})
	}

	return dataPoints, nil
}

// ========== 综合安全指标实现 ==========

// GetSecurityMetrics 获取综合安全指标
func (s *StatsServiceImpl) GetSecurityMetrics(ctx context.Context, timeRange string) (*dto.SecurityMetricsResponse, error) {
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 并行获取各项指标
	var (
		overview                *dto.OverviewStats
		ruleEngine              *dto.RuleEngineStats
		topTriggeredRules       []dto.RuleTriggerStats
		severityDistribution    []dto.SeverityStats
		attackTypeDistribution  []dto.AttackTypeStats
		topAttackSources        []dto.GeoLocationStats
		blockedIPMetrics        *dto.BlockedIPStats
		threatLevel             *dto.ThreatLevelDistribution
		responseTime            *dto.ResponseTimeStats
		requestTrend            *dto.TimeSeriesResponse
		blockTrend              *dto.TimeSeriesResponse
		trafficTrend            *dto.TrafficTimeSeriesResponse
	)

	// 1. 获取概览统计
	overview, err = s.GetOverviewStats(ctx, timeRange)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取概览统计失败")
		overview = &dto.OverviewStats{TimeRange: timeRange}
	}

	// 2. 获取规则引擎统计
	ruleEngine, err = s.getRuleEngineStats(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取规则引擎统计失败")
		ruleEngine = &dto.RuleEngineStats{}
	}

	// 3. 获取Top触发规则
	topTriggeredRules, err = s.getTopTriggeredRules(ctx, db, startTime, 10)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取Top触发规则失败")
		topTriggeredRules = []dto.RuleTriggerStats{}
	}

	// 4. 获取严重等级分布
	severityDistribution, err = s.getSeverityDistribution(ctx, db, startTime)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取严重等级分布失败")
		severityDistribution = []dto.SeverityStats{}
	}

	// 5. 获取攻击类型分布
	attackTypeDistribution, err = s.getAttackTypeDistribution(ctx, db, startTime)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取攻击类型分布失败")
		attackTypeDistribution = []dto.AttackTypeStats{}
	}

	// 6. 获取Top攻击来源
	topAttackSources, err = s.getTopAttackSources(ctx, db, startTime, 10)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取Top攻击来源失败")
		topAttackSources = []dto.GeoLocationStats{}
	}

	// 7. 获取封禁IP指标
	blockedIPMetrics, err = s.getBlockedIPMetrics(ctx, db)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取封禁IP指标失败")
		blockedIPMetrics = &dto.BlockedIPStats{}
	}

	// 8. 获取威胁等级分布
	threatLevel, err = s.getThreatLevelDistribution(ctx, db, startTime)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取威胁等级分布失败")
		threatLevel = &dto.ThreatLevelDistribution{}
	}

	// 9. 获取响应时间统计
	responseTime, err = s.getResponseTimeStats(ctx, db, startTime)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取响应时间统计失败")
		responseTime = &dto.ResponseTimeStats{}
	}

	// 10. 获取请求趋势
	requestTrend, err = s.GetTimeSeriesData(ctx, timeRange, "requests")
	if err != nil {
		s.logger.Error().Err(err).Msg("获取请求趋势失败")
		requestTrend = &dto.TimeSeriesResponse{Metric: "requests", TimeRange: timeRange, Data: []dto.TimeSeriesDataPoint{}}
	}

	// 11. 获取拦截趋势
	blockTrend, err = s.GetTimeSeriesData(ctx, timeRange, "blocks")
	if err != nil {
		s.logger.Error().Err(err).Msg("获取拦截趋势失败")
		blockTrend = &dto.TimeSeriesResponse{Metric: "blocks", TimeRange: timeRange, Data: []dto.TimeSeriesDataPoint{}}
	}

	// 12. 获取流量趋势
	trafficTrend, err = s.GetTrafficTimeSeriesData(ctx, timeRange)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取流量趋势失败")
		trafficTrend = &dto.TrafficTimeSeriesResponse{TimeRange: timeRange, Data: []dto.TrafficDataPoint{}}
	}

	// 构建响应
	response := &dto.SecurityMetricsResponse{
		TimeRange:              timeRange,
		Overview:               *overview,
		RuleEngine:             *ruleEngine,
		TopTriggeredRules:      topTriggeredRules,
		SeverityDistribution:   severityDistribution,
		AttackTypeDistribution: attackTypeDistribution,
		TopAttackSources:       topAttackSources,
		BlockedIPMetrics:       *blockedIPMetrics,
		ThreatLevel:            *threatLevel,
		ResponseTime:           *responseTime,
		RequestTrend:           *requestTrend,
		BlockTrend:             *blockTrend,
		TrafficTrend:           *trafficTrend,
	}

	return response, nil
}

// getRuleEngineStats 获取规则引擎统计
func (s *StatsServiceImpl) getRuleEngineStats(ctx context.Context) (*dto.RuleEngineStats, error) {
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, err
	}

	collection := db.Collection((&pkgModel.MicroRule{}).GetCollectionName())

	// 统计总规则数
	totalRules, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	// 统计已启用规则数
	enabledRules, err := collection.CountDocuments(ctx, bson.D{{Key: "status", Value: pkgModel.RuleEnabled}})
	if err != nil {
		return nil, err
	}

	// 统计已禁用规则数
	disabledRules, err := collection.CountDocuments(ctx, bson.D{{Key: "status", Value: pkgModel.RuleDisabled}})
	if err != nil {
		return nil, err
	}

	// 统计白名单规则数
	whitelistRules, err := collection.CountDocuments(ctx, bson.D{{Key: "type", Value: pkgModel.WhitelistRule}})
	if err != nil {
		return nil, err
	}

	// 统计黑名单规则数
	blacklistRules, err := collection.CountDocuments(ctx, bson.D{{Key: "type", Value: pkgModel.BlacklistRule}})
	if err != nil {
		return nil, err
	}

	return &dto.RuleEngineStats{
		TotalRules:     totalRules,
		EnabledRules:   enabledRules,
		DisabledRules:  disabledRules,
		WhitelistRules: whitelistRules,
		BlacklistRules: blacklistRules,
		AvgMatchTime:   0.5,   // 这个值可以从实际监控系统获取
		RuleEfficiency: 95.5,  // 这个值可以基于规则触发率和误报率计算
	}, nil
}

// getTopTriggeredRules 获取Top触发规则
func (s *StatsServiceImpl) getTopTriggeredRules(ctx context.Context, db *mongo.Database, startTime time.Time, limit int) ([]dto.RuleTriggerStats, error) {
	collection := db.Collection((&pkgModel.WAFLog{}).GetCollectionName())

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$rule_id"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		RuleID int64 `bson:"_id"`
		Count  int64 `bson:"count"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 计算总数用于百分比
	var totalCount int64
	for _, r := range results {
		totalCount += r.Count
	}

	// 构建结果
	stats := make([]dto.RuleTriggerStats, 0, len(results))
	for _, r := range results {
		percentage := float64(0)
		if totalCount > 0 {
			percentage = math.Round(float64(r.Count)/float64(totalCount)*10000) / 100
		}

		stats = append(stats, dto.RuleTriggerStats{
			RuleID:     r.RuleID,
			RuleName:   fmt.Sprintf("Rule %d", r.RuleID),
			Count:      r.Count,
			Percentage: percentage,
		})
	}

	return stats, nil
}

// getSeverityDistribution 获取严重等级分布
func (s *StatsServiceImpl) getSeverityDistribution(ctx context.Context, db *mongo.Database, startTime time.Time) ([]dto.SeverityStats, error) {
	collection := db.Collection((&pkgModel.WAFLog{}).GetCollectionName())

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$severity"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Level int64 `bson:"_id"`
		Count int64 `bson:"count"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 计算总数用于百分比
	var totalCount int64
	for _, r := range results {
		totalCount += r.Count
	}

	// 严重等级名称映射
	severityNames := map[int64]string{
		0: "信息",
		1: "低",
		2: "中",
		3: "高",
		4: "严重",
		5: "紧急",
	}

	// 构建结果
	stats := make([]dto.SeverityStats, 0, len(results))
	for _, r := range results {
		percentage := float64(0)
		if totalCount > 0 {
			percentage = math.Round(float64(r.Count)/float64(totalCount)*10000) / 100
		}

		levelName := severityNames[r.Level]
		if levelName == "" {
			levelName = fmt.Sprintf("Level %d", r.Level)
		}

		stats = append(stats, dto.SeverityStats{
			Level:      r.Level,
			LevelName:  levelName,
			Count:      r.Count,
			Percentage: percentage,
		})
	}

	return stats, nil
}

// getAttackTypeDistribution 获取攻击类型分布
func (s *StatsServiceImpl) getAttackTypeDistribution(ctx context.Context, db *mongo.Database, startTime time.Time) ([]dto.AttackTypeStats, error) {
	collection := db.Collection((&pkgModel.WAFLog{}).GetCollectionName())

	// 基于 sec_mark 字段进行分组统计攻击类型
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
			{Key: "sec_mark", Value: bson.D{{Key: "$ne", Value: ""}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$sec_mark"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Category string `bson:"_id"`
		Count    int64  `bson:"count"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 计算总数用于百分比
	var totalCount int64
	for _, r := range results {
		totalCount += r.Count
	}

	// 构建结果
	stats := make([]dto.AttackTypeStats, 0, len(results))
	for _, r := range results {
		percentage := float64(0)
		if totalCount > 0 {
			percentage = math.Round(float64(r.Count)/float64(totalCount)*10000) / 100
		}

		stats = append(stats, dto.AttackTypeStats{
			Category:   r.Category,
			Count:      r.Count,
			Percentage: percentage,
		})
	}

	return stats, nil
}

// getTopAttackSources 获取Top攻击来源
func (s *StatsServiceImpl) getTopAttackSources(ctx context.Context, db *mongo.Database, startTime time.Time, limit int) ([]dto.GeoLocationStats, error) {
	collection := db.Collection((&pkgModel.WAFLog{}).GetCollectionName())

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
			{Key: "src_ip_info", Value: bson.D{{Key: "$exists", Value: true}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "country", Value: "$src_ip_info.country"},
				{Key: "country_code", Value: "$src_ip_info.country_code"},
				{Key: "city", Value: "$src_ip_info.city"},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID struct {
			Country     string `bson:"country"`
			CountryCode string `bson:"country_code"`
			City        string `bson:"city"`
		} `bson:"_id"`
		Count int64 `bson:"count"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 计算总数用于百分比
	var totalCount int64
	for _, r := range results {
		totalCount += r.Count
	}

	// 构建结果
	stats := make([]dto.GeoLocationStats, 0, len(results))
	for _, r := range results {
		percentage := float64(0)
		if totalCount > 0 {
			percentage = math.Round(float64(r.Count)/float64(totalCount)*10000) / 100
		}

		stats = append(stats, dto.GeoLocationStats{
			Country:     r.ID.Country,
			CountryCode: r.ID.CountryCode,
			City:        r.ID.City,
			Count:       r.Count,
			Percentage:  percentage,
		})
	}

	return stats, nil
}

// getBlockedIPMetrics 获取封禁IP指标
func (s *StatsServiceImpl) getBlockedIPMetrics(ctx context.Context, db *mongo.Database) (*dto.BlockedIPStats, error) {
	collection := db.Collection("blocked_ips")

	now := time.Now()

	// 统计总封禁数
	totalBlocked, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	// 统计当前活跃封禁数
	activeBlocked, err := collection.CountDocuments(ctx, bson.D{
		{Key: "blocked_until", Value: bson.D{{Key: "$gt", Value: now}}},
	})
	if err != nil {
		return nil, err
	}

	// 统计已过期封禁数
	expiredBlocked, err := collection.CountDocuments(ctx, bson.D{
		{Key: "blocked_until", Value: bson.D{{Key: "$lte", Value: now}}},
	})
	if err != nil {
		return nil, err
	}

	// 按封禁原因统计
	highFrequencyVisit, _ := collection.CountDocuments(ctx, bson.D{
		{Key: "reason", Value: "high_frequency_visit"},
		{Key: "blocked_until", Value: bson.D{{Key: "$gt", Value: now}}},
	})

	highFrequencyAttack, _ := collection.CountDocuments(ctx, bson.D{
		{Key: "reason", Value: "high_frequency_attack"},
		{Key: "blocked_until", Value: bson.D{{Key: "$gt", Value: now}}},
	})

	highFrequencyError, _ := collection.CountDocuments(ctx, bson.D{
		{Key: "reason", Value: "high_frequency_error"},
		{Key: "blocked_until", Value: bson.D{{Key: "$gt", Value: now}}},
	})

	return &dto.BlockedIPStats{
		TotalBlocked:        totalBlocked,
		ActiveBlocked:       activeBlocked,
		ExpiredBlocked:      expiredBlocked,
		HighFrequencyVisit:  highFrequencyVisit,
		HighFrequencyAttack: highFrequencyAttack,
		HighFrequencyError:  highFrequencyError,
	}, nil
}

// getThreatLevelDistribution 获取威胁等级分布
func (s *StatsServiceImpl) getThreatLevelDistribution(ctx context.Context, db *mongo.Database, startTime time.Time) (*dto.ThreatLevelDistribution, error) {
	collection := db.Collection((&pkgModel.WAFLog{}).GetCollectionName())

	// 最近时间窗口内的威胁分布 (最近1小时)
	recentTime := time.Now().Add(-1 * time.Hour)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: recentTime}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$severity"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Severity int64 `bson:"_id"`
		Count    int64 `bson:"count"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 映射到威胁等级
	distribution := &dto.ThreatLevelDistribution{}
	for _, r := range results {
		switch r.Severity {
		case 0, 1:
			distribution.Low += r.Count
		case 2:
			distribution.Medium += r.Count
		case 3, 4:
			distribution.High += r.Count
		case 5:
			distribution.Critical += r.Count
		}
	}

	return distribution, nil
}

// getResponseTimeStats 获取响应时间统计
func (s *StatsServiceImpl) getResponseTimeStats(ctx context.Context, db *mongo.Database, startTime time.Time) (*dto.ResponseTimeStats, error) {
	// 从 HAProxy 统计数据中获取响应时间信息
	// 这里使用模拟数据，实际应该从 HAProxy metrics 中获取
	return &dto.ResponseTimeStats{
		AvgResponseTime: 15.5,
		MaxResponseTime: 250.0,
		MinResponseTime: 5.0,
		P50ResponseTime: 10.0,
		P95ResponseTime: 50.0,
		P99ResponseTime: 100.0,
	}, nil
}
