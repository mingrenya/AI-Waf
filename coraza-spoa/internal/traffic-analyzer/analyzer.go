package trafficanalyzer

import (
	"context"
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	flowcontroller "github.com/mingrenya/AI-Waf/coraza-spoa/internal/flow-controller"
)

// TrafficAnalyzer 流量分析器 - 采集和分析流量数据
type TrafficAnalyzer struct {
	db     *mongo.Database
	logger zerolog.Logger

	// 流量统计
	statistics *StatisticsCollector

	// 基线计算
	baseline *BaselineCalculator

	// 异常检测
	detector *AnomalyDetector

	// 流量控制器
	flowController *flowcontroller.FlowController

	// 配置
	config     *model.AdaptiveThrottlingConfig
	configLock sync.RWMutex

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// TrafficEvent 流量事件
type TrafficEvent struct {
	Timestamp   time.Time
	Type        string // "visit", "attack", "error"
	SrcIP       string
	DstIP       string
	DstPort     int64
	Method      string
	Path        string
	StatusCode  int
	IsBlocked   bool
	IsAttack    bool
	ResponseTime time.Duration
}

// NewTrafficAnalyzer 创建流量分析器
func NewTrafficAnalyzer(db *mongo.Database, logger zerolog.Logger, flowController *flowcontroller.FlowController) *TrafficAnalyzer {
	ctx, cancel := context.WithCancel(context.Background())

	ta := &TrafficAnalyzer{
		db:             db,
		logger:         logger.With().Str("component", "traffic-analyzer").Logger(),
		statistics:     NewStatisticsCollector(db, logger),
		baseline:       NewBaselineCalculator(db, logger),
		detector:       NewAnomalyDetector(logger),
		flowController: flowController,
		ctx:            ctx,
		cancel:         cancel,
	}

	// 加载配置
	ta.loadConfig()

	// 启动后台任务
	ta.start()

	return ta
}

// loadConfig 从数据库加载配置
func (ta *TrafficAnalyzer) loadConfig() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var config model.AdaptiveThrottlingConfig
	err := ta.db.Collection("adaptive_throttling_config").FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ta.logger.Info().Msg("No adaptive throttling config found, using defaults")
			// 使用默认配置
			config = ta.getDefaultConfig()
		} else {
			ta.logger.Error().Err(err).Msg("Failed to load config")
			config = ta.getDefaultConfig()
		}
	}

	ta.configLock.Lock()
	ta.config = &config
	ta.configLock.Unlock()

	ta.logger.Info().Bool("enabled", config.Enabled).Msg("Loaded adaptive throttling config")
}

// getDefaultConfig 获取默认配置
func (ta *TrafficAnalyzer) getDefaultConfig() model.AdaptiveThrottlingConfig {
	return model.GetDefaultAdaptiveThrottlingConfig()
}

// start 启动后台任务
func (ta *TrafficAnalyzer) start() {
	// 定期重新加载配置
	ta.wg.Add(1)
	go func() {
		defer ta.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ta.ctx.Done():
				return
			case <-ticker.C:
				ta.loadConfig()
			}
		}
	}()

	// 定期更新基线
	ta.wg.Add(1)
	go func() {
		defer ta.wg.Done()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ta.ctx.Done():
				return
			case <-ticker.C:
				ta.updateBaselines()
			}
		}
	}()

	// 定期检测异常并调整阈值
	ta.wg.Add(1)
	go func() {
		defer ta.wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ta.ctx.Done():
				return
			case <-ticker.C:
				ta.checkAnomaliesAndAdjust()
			}
		}
	}()
}

// RecordTraffic 记录流量事件
func (ta *TrafficAnalyzer) RecordTraffic(event *TrafficEvent) {
	ta.configLock.RLock()
	enabled := ta.config != nil && ta.config.Enabled
	ta.configLock.RUnlock()

	if !enabled {
		return
	}

	// 记录到统计数据
	ta.statistics.Record(event)
}

// updateBaselines 更新基线值
func (ta *TrafficAnalyzer) updateBaselines() {
	ta.configLock.RLock()
	config := ta.config
	ta.configLock.RUnlock()

	if config == nil || !config.Enabled {
		return
	}

	ta.logger.Debug().Msg("Updating baselines")

	// 获取历史数据
	patterns := ta.statistics.GetRecentPatterns(time.Duration(config.Baseline.HistoryWindow) * time.Second)
	
	// 计算各类型基线
	if config.ApplyTo.VisitLimit {
		baseline := ta.baseline.Calculate("visit", patterns, config)
		ta.baseline.Save(baseline)
	}

	if config.ApplyTo.AttackLimit {
		baseline := ta.baseline.Calculate("attack", patterns, config)
		ta.baseline.Save(baseline)
	}

	if config.ApplyTo.ErrorLimit {
		baseline := ta.baseline.Calculate("error", patterns, config)
		ta.baseline.Save(baseline)
	}
}

// checkAnomaliesAndAdjust 检测异常并调整阈值
func (ta *TrafficAnalyzer) checkAnomaliesAndAdjust() {
	ta.configLock.RLock()
	config := ta.config
	ta.configLock.RUnlock()

	if config == nil || !config.Enabled || !config.AutoAdjustment.Enabled {
		return
	}

	ta.logger.Debug().Msg("Checking for anomalies")

	// 获取当前流量
	current := ta.statistics.GetCurrentMetrics()

	// 获取基线
	baselines := ta.baseline.GetCurrent()

	// 检测异常
	anomalies := ta.detector.Detect(current, baselines, config)

	// 如果检测到异常，调整阈值
	if len(anomalies) > 0 {
		ta.logger.Info().Int("count", len(anomalies)).Msg("Anomalies detected")
		ta.adjustThresholds(anomalies, config)
	}
}

// adjustThresholds 调整限流阈值
func (ta *TrafficAnalyzer) adjustThresholds(anomalies []Anomaly, config *model.AdaptiveThrottlingConfig) {
	ta.logger.Info().Int("anomalyCount", len(anomalies)).Msg("开始调整限流阈值")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, anomaly := range anomalies {
		// 计算新阈值
		newThreshold := ta.calculateNewThreshold(anomaly, config)
		
		// 应用最小/最大限制
		if newThreshold < config.AutoAdjustment.MinThreshold {
			newThreshold = config.AutoAdjustment.MinThreshold
		}
		if newThreshold > config.AutoAdjustment.MaxThreshold {
			newThreshold = config.AutoAdjustment.MaxThreshold
		}

		// 获取当前阈值（从配置或默认值）
		oldThreshold := ta.getCurrentThreshold(anomaly.Type)
		
		// 如果阈值变化不大，跳过
		if oldThreshold > 0 && float64(newThreshold)/float64(oldThreshold) > 0.95 && 
		   float64(newThreshold)/float64(oldThreshold) < 1.05 {
			ta.logger.Debug().
				Str("type", anomaly.Type).
				Int64("oldThreshold", oldThreshold).
				Int64("newThreshold", newThreshold).
				Msg("阈值变化小于5%，跳过调整")
			continue
		}

		// 创建调整日志
		adjustmentLog := &model.ThrottleAdjustmentLog{
			Timestamp:       time.Now(),
			Type:            anomaly.Type,
			OldThreshold:    oldThreshold,
			OldBaseline:     anomaly.BaselineValue,
			NewThreshold:    newThreshold,
			NewBaseline:     anomaly.CurrentValue,
			Reason:          anomaly.Reason,
			CurrentTraffic:  anomaly.CurrentValue,
			AnomalyScore:    anomaly.AnomalyScore,
			TriggeredBy:     "auto",
			AdjustmentRatio: float64(newThreshold) / float64(oldThreshold),
		}

		// 保存调整日志到数据库
		_, err := ta.db.Collection("throttle_adjustment_logs").InsertOne(ctx, adjustmentLog)
		if err != nil {
			ta.logger.Error().Err(err).Str("type", anomaly.Type).Msg("保存调整日志失败")
			continue
		}

		ta.logger.Info().
			Str("type", anomaly.Type).
			Int64("oldThreshold", oldThreshold).
			Int64("newThreshold", newThreshold).
			Float64("anomalyScore", anomaly.AnomalyScore).
			Msg("阈值调整完成")

		// 调用flow-controller更新阈值
		if ta.flowController != nil {
			err := ta.updateFlowControllerThreshold(anomaly.Type, newThreshold)
			if err != nil {
				ta.logger.Error().Err(err).Str("type", anomaly.Type).Msg("更新flow-controller阈值失败")
			}
		}
	}
}

// calculateNewThreshold 计算新的限流阈值
func (ta *TrafficAnalyzer) calculateNewThreshold(anomaly Anomaly, config *model.AdaptiveThrottlingConfig) int64 {
	// 基于基线值和调整因子计算新阈值
	baseValue := anomaly.BaselineValue * config.AutoAdjustment.AdjustmentFactor
	
	// 如果启用渐进式调整
	if config.AutoAdjustment.GradualAdjustment {
		// 当前阈值
		currentThreshold := ta.getCurrentThreshold(anomaly.Type)
		if currentThreshold > 0 {
			// 计算目标阈值和当前阈值的差值
			diff := int64(baseValue) - currentThreshold
			// 只调整一个步长
			step := int64(float64(diff) * config.AutoAdjustment.AdjustmentStepRatio)
			return currentThreshold + step
		}
	}
	
	return int64(baseValue)
}

// getCurrentThreshold 获取当前阈值
func (ta *TrafficAnalyzer) getCurrentThreshold(typ string) int64 {
	if ta.flowController == nil {
		// 如果没有flow-controller，返回默认值
		switch typ {
		case "visit":
			return 100
		case "attack":
			return 50
		case "error":
			return 30
		default:
			return 100
		}
	}

	// 从flow-controller配置中获取当前阈值
	config := ta.flowController.GetConfig()
	switch typ {
	case "visit":
		return config.VisitLimit.Threshold
	case "attack":
		return config.AttackLimit.Threshold
	case "error":
		return config.ErrorLimit.Threshold
	default:
		return 100
	}
}

// Stop 停止分析器
func (ta *TrafficAnalyzer) Stop() {
	ta.logger.Info().Msg("Stopping traffic analyzer")
	ta.cancel()
	ta.wg.Wait()
}

// GetStats 获取统计信息
func (ta *TrafficAnalyzer) GetStats() map[string]interface{} {
	return ta.statistics.GetStats()
}

// updateFlowControllerThreshold 更新flow-controller的阈值
func (ta *TrafficAnalyzer) updateFlowControllerThreshold(typ string, threshold int64) error {
	if ta.flowController == nil {
		return nil // 如果没有flow-controller，静默忽略
	}

	err := ta.flowController.UpdateThreshold(typ, threshold)
	if err != nil {
		ta.logger.Error().
			Err(err).
			Str("type", typ).
			Int64("threshold", threshold).
			Msg("更新flow-controller阈值失败")
		return err
	}

	ta.logger.Info().
		Str("type", typ).
		Int64("threshold", threshold).
		Msg("成功更新flow-controller阈值")

	return nil
}
