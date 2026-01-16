package service

import (
	"context"
	"errors"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrAdaptiveThrottlingConfigNotFound = errors.New("自适应限流配置不存在")
)

// AdaptiveThrottlingService 自适应限流服务接口
type AdaptiveThrottlingService interface {
	// 配置管理
	GetConfig(ctx context.Context) (*model.AdaptiveThrottlingConfig, error)
	CreateConfig(ctx context.Context, req *dto.AdaptiveThrottlingConfigRequest) (*model.AdaptiveThrottlingConfig, error)
	UpdateConfig(ctx context.Context, req *dto.AdaptiveThrottlingConfigRequest) (*model.AdaptiveThrottlingConfig, error)
	DeleteConfig(ctx context.Context) error

	// 流量模式查询
	GetTrafficPatterns(ctx context.Context, query *dto.TrafficPatternQuery) (*dto.TrafficPatternResponse, error)

	// 基线值查询
	GetBaselines(ctx context.Context, query *dto.BaselineQuery) (*dto.BaselineResponse, error)

	// 调整日志查询
	GetAdjustmentLogs(ctx context.Context, query *dto.AdjustmentLogQuery) (*dto.AdjustmentLogResponse, error)

	// 统计信息
	GetStats(ctx context.Context) (*dto.AdaptiveThrottlingStatsDTO, error)

	// 操作
	RecalculateBaseline(ctx context.Context, typ string) error
	ResetLearning(ctx context.Context) error
}

// AdaptiveThrottlingServiceImpl 自适应限流服务实现
type AdaptiveThrottlingServiceImpl struct {
	repo   repository.AdaptiveThrottlingRepository
	logger zerolog.Logger
}

// NewAdaptiveThrottlingService 创建自适应限流服务
func NewAdaptiveThrottlingService(repo repository.AdaptiveThrottlingRepository) AdaptiveThrottlingService {
	logger := config.GetServiceLogger("adaptive_throttling")
	return &AdaptiveThrottlingServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// GetConfig 获取配置
func (s *AdaptiveThrottlingServiceImpl) GetConfig(ctx context.Context) (*model.AdaptiveThrottlingConfig, error) {
	cfg, err := s.repo.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAdaptiveThrottlingConfigNotFound
		}
		s.logger.Error().Err(err).Msg("获取自适应限流配置失败")
		return nil, err
	}
	return cfg, nil
}

// CreateConfig 创建配置
func (s *AdaptiveThrottlingServiceImpl) CreateConfig(ctx context.Context, req *dto.AdaptiveThrottlingConfigRequest) (*model.AdaptiveThrottlingConfig, error) {
	// 检查是否已存在配置
	existing, err := s.repo.GetConfig(ctx)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		s.logger.Error().Err(err).Msg("检查配置是否存在失败")
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("配置已存在，请使用更新接口")
	}

	// 转换DTO为模型
	cfg := s.dtoToModel(req)

	// 创建配置
	if err := s.repo.CreateConfig(ctx, cfg); err != nil {
		s.logger.Error().Err(err).Msg("创建自适应限流配置失败")
		return nil, err
	}

	return cfg, nil
}

// UpdateConfig 更新配置
func (s *AdaptiveThrottlingServiceImpl) UpdateConfig(ctx context.Context, req *dto.AdaptiveThrottlingConfigRequest) (*model.AdaptiveThrottlingConfig, error) {
	// 获取现有配置
	existing, err := s.repo.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAdaptiveThrottlingConfigNotFound
		}
		s.logger.Error().Err(err).Msg("获取现有配置失败")
		return nil, err
	}

	// 转换DTO为模型并保留ID
	cfg := s.dtoToModel(req)
	cfg.ID = existing.ID
	cfg.CreatedAt = existing.CreatedAt

	// 更新配置
	if err := s.repo.UpdateConfig(ctx, cfg); err != nil {
		s.logger.Error().Err(err).Msg("更新自适应限流配置失败")
		return nil, err
	}

	return cfg, nil
}

// DeleteConfig 删除配置
func (s *AdaptiveThrottlingServiceImpl) DeleteConfig(ctx context.Context) error {
	if err := s.repo.DeleteConfig(ctx); err != nil {
		s.logger.Error().Err(err).Msg("删除自适应限流配置失败")
		return err
	}
	return nil
}

// GetTrafficPatterns 获取流量模式
func (s *AdaptiveThrottlingServiceImpl) GetTrafficPatterns(ctx context.Context, query *dto.TrafficPatternQuery) (*dto.TrafficPatternResponse, error) {
	// 构建过滤条件
	filter := bson.M{}
	if query.Type != "" {
		filter["type"] = query.Type
	}
	if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
		timeFilter := bson.M{}
		if !query.StartTime.IsZero() {
			timeFilter["$gte"] = query.StartTime
		}
		if !query.EndTime.IsZero() {
			timeFilter["$lte"] = query.EndTime
		}
		filter["timestamp"] = timeFilter
	}

	// 分页参数
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	// 查询数据
	patterns, total, err := s.repo.GetTrafficPatterns(ctx, filter, skip, limit)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取流量模式失败")
		return nil, err
	}

	// 转换为DTO
	results := make([]dto.TrafficPatternDTO, len(patterns))
	for i, p := range patterns {
		results[i] = s.trafficPatternToDTO(p)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &dto.TrafficPatternResponse{
		Results:     results,
		TotalCount:  int(total),
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

// GetBaselines 获取基线值
func (s *AdaptiveThrottlingServiceImpl) GetBaselines(ctx context.Context, query *dto.BaselineQuery) (*dto.BaselineResponse, error) {
	// 构建过滤条件
	filter := bson.M{}
	if query.Type != "" {
		filter["type"] = query.Type
	}

	// 查询数据
	baselines, err := s.repo.GetBaselines(ctx, filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取基线值失败")
		return nil, err
	}

	// 转换为DTO
	results := make([]dto.BaselineValueDTO, len(baselines))
	for i, b := range baselines {
		results[i] = s.baselineToDTO(b)
	}

	return &dto.BaselineResponse{
		Results: results,
	}, nil
}

// GetAdjustmentLogs 获取调整日志
func (s *AdaptiveThrottlingServiceImpl) GetAdjustmentLogs(ctx context.Context, query *dto.AdjustmentLogQuery) (*dto.AdjustmentLogResponse, error) {
	// 构建过滤条件
	filter := bson.M{}
	if query.Type != "" {
		filter["type"] = query.Type
	}
	if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
		timeFilter := bson.M{}
		if !query.StartTime.IsZero() {
			timeFilter["$gte"] = query.StartTime
		}
		if !query.EndTime.IsZero() {
			timeFilter["$lte"] = query.EndTime
		}
		filter["timestamp"] = timeFilter
	}

	// 分页参数
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	// 查询数据
	logs, total, err := s.repo.GetAdjustmentLogs(ctx, filter, skip, limit)
	if err != nil {
		s.logger.Error().Err(err).Msg("获取调整日志失败")
		return nil, err
	}

	// 转换为DTO
	results := make([]dto.ThrottleAdjustmentLogDTO, len(logs))
	for i, l := range logs {
		results[i] = s.adjustmentLogToDTO(l)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &dto.AdjustmentLogResponse{
		Results:     results,
		TotalCount:  int(total),
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

// GetStats 获取统计信息
func (s *AdaptiveThrottlingServiceImpl) GetStats(ctx context.Context) (*dto.AdaptiveThrottlingStatsDTO, error) {
	// 获取配置
	cfg, err := s.repo.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAdaptiveThrottlingConfigNotFound
		}
		return nil, err
	}

	// 获取所有基线值
	baselines, err := s.repo.GetBaselines(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// 构建基线统计
	baselineStats := dto.BaselineStatsDTO{}
	thresholdStats := dto.ThresholdStatsDTO{
		Visit:  100, // 默认值，实际应从配置或其他地方获取
		Attack: 50,
		Error:  30,
	}

	for _, b := range baselines {
		switch b.Type {
		case "visit":
			baselineStats.Visit = b.Value
		case "attack":
			baselineStats.Attack = b.Value
		case "error":
			baselineStats.Error = b.Value
		}
	}

	// 计算学习进度
	learningProgress := 100.0
	if cfg.LearningMode.Enabled {
		// 这里应该根据实际学习时长计算进度
		// 简化处理，实际需要从数据库获取学习开始时间
		learningProgress = 75.0
	}

	// 获取最近24小时调整次数
	since := time.Now().Add(-24 * time.Hour)
	recentAdjustments, err := s.repo.GetRecentAdjustmentCount(ctx, since)
	if err != nil {
		s.logger.Warn().Err(err).Msg("获取最近调整次数失败")
		recentAdjustments = 0
	}

	// 检测异常（简化处理）
	anomalyDetected := false

	return &dto.AdaptiveThrottlingStatsDTO{
		CurrentBaseline:   baselineStats,
		CurrentThreshold:  thresholdStats,
		LearningProgress:  learningProgress,
		RecentAdjustments: int(recentAdjustments),
		AnomalyDetected:   anomalyDetected,
		LastUpdateTime:    time.Now(),
	}, nil
}

// RecalculateBaseline 重新计算基线
func (s *AdaptiveThrottlingServiceImpl) RecalculateBaseline(ctx context.Context, typ string) error {
	// TODO: 实现基线重新计算逻辑
	s.logger.Info().Str("type", typ).Msg("重新计算基线")
	return nil
}

// ResetLearning 重置学习
func (s *AdaptiveThrottlingServiceImpl) ResetLearning(ctx context.Context) error {
	// TODO: 实现学习重置逻辑
	s.logger.Info().Msg("重置学习")
	return nil
}

// dtoToModel 将DTO转换为模型
func (s *AdaptiveThrottlingServiceImpl) dtoToModel(req *dto.AdaptiveThrottlingConfigRequest) *model.AdaptiveThrottlingConfig {
	config := &model.AdaptiveThrottlingConfig{
		Enabled: req.Enabled,
	}
	
	// 学习模式
	config.LearningMode.Enabled = req.LearningMode.Enabled
	config.LearningMode.LearningDuration = req.LearningMode.LearningDuration
	config.LearningMode.SampleInterval = req.LearningMode.SampleInterval
	config.LearningMode.MinSamples = int64(req.LearningMode.MinSamples)
	
	// 基线配置
	config.Baseline.CalculationMethod = req.Baseline.CalculationMethod
	config.Baseline.Percentile = float64(req.Baseline.Percentile)
	config.Baseline.UpdateInterval = req.Baseline.UpdateInterval
	config.Baseline.HistoryWindow = req.Baseline.HistoryWindow
	
	// 自动调整
	config.AutoAdjustment.Enabled = req.AutoAdjustment.Enabled
	config.AutoAdjustment.AnomalyThreshold = req.AutoAdjustment.AnomalyThreshold
	config.AutoAdjustment.MinThreshold = int64(req.AutoAdjustment.MinThreshold)
	config.AutoAdjustment.MaxThreshold = int64(req.AutoAdjustment.MaxThreshold)
	config.AutoAdjustment.AdjustmentFactor = req.AutoAdjustment.AdjustmentFactor
	config.AutoAdjustment.CooldownPeriod = req.AutoAdjustment.CooldownPeriod
	config.AutoAdjustment.GradualAdjustment = req.AutoAdjustment.GradualAdjustment
	config.AutoAdjustment.AdjustmentStepRatio = req.AutoAdjustment.AdjustmentStepRatio
	
	// 应用范围
	config.ApplyTo.VisitLimit = req.ApplyTo.VisitLimit
	config.ApplyTo.AttackLimit = req.ApplyTo.AttackLimit
	config.ApplyTo.ErrorLimit = req.ApplyTo.ErrorLimit
	
	return config
}

// trafficPatternToDTO 将流量模式转换为DTO
func (s *AdaptiveThrottlingServiceImpl) trafficPatternToDTO(p *model.TrafficPattern) dto.TrafficPatternDTO {
	return dto.TrafficPatternDTO{
		Type:      p.Type,
		Timestamp: p.Timestamp,
		Metrics: dto.TrafficMetricsDTO{
			RequestCount: int(p.Metrics.RequestRate * 60), // 转换为每分钟请求数
			AvgLatency:   0,                                // 模型中没有这个字段，使用0
			ErrorRate:    0,                                // 模型中没有这个字段，使用0
			P95Latency:   0,                                // 模型中没有这个字段，使用0
			P99Latency:   0,                                // 模型中没有这个字段，使用0
		},
		Statistics: dto.TrafficStatisticsDTO{
			Mean:   p.Statistics.Mean,
			Median: p.Statistics.Median,
			StdDev: p.Statistics.StdDev,
			Min:    p.Statistics.Min,
			Max:    p.Statistics.Max,
		},
	}
}

// baselineToDTO 将基线值转换为DTO
func (s *AdaptiveThrottlingServiceImpl) baselineToDTO(b *model.BaselineValue) dto.BaselineValueDTO {
	return dto.BaselineValueDTO{
		Type:            b.Type,
		Value:           b.Value,
		ConfidenceLevel: b.ConfidenceLevel,
		SampleSize:      int(b.SampleSize),
		CalculatedAt:    b.CalculatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

// adjustmentLogToDTO 将调整日志转换为DTO
func (s *AdaptiveThrottlingServiceImpl) adjustmentLogToDTO(l *model.ThrottleAdjustmentLog) dto.ThrottleAdjustmentLogDTO {
	return dto.ThrottleAdjustmentLogDTO{
		ID:              l.ID,
		Type:            l.Type,
		Timestamp:       l.Timestamp,
		OldThreshold:    int(l.OldThreshold),
		NewThreshold:    int(l.NewThreshold),
		AdjustmentRatio: l.AdjustmentRatio,
		OldBaseline:     l.OldBaseline,
		NewBaseline:     l.NewBaseline,
		CurrentTraffic:  l.CurrentTraffic,
		AnomalyScore:    l.AnomalyScore,
		Reason:          l.Reason,
		TriggeredBy:     l.TriggeredBy,
	}
}
