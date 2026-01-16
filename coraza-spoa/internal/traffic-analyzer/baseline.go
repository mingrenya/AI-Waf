package trafficanalyzer

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BaselineCalculator 基线计算器
type BaselineCalculator struct {
	db     *mongo.Database
	logger zerolog.Logger

	// 缓存当前基线
	currentBaselines map[string]*model.BaselineValue
}

// NewBaselineCalculator 创建基线计算器
func NewBaselineCalculator(db *mongo.Database, logger zerolog.Logger) *BaselineCalculator {
	return &BaselineCalculator{
		db:               db,
		logger:           logger.With().Str("component", "baseline").Logger(),
		currentBaselines: make(map[string]*model.BaselineValue),
	}
}

// Calculate 计算基线值
func (bc *BaselineCalculator) Calculate(typ string, patterns []model.TrafficPattern, config *model.AdaptiveThrottlingConfig) *model.BaselineValue {
	if len(patterns) == 0 {
		bc.logger.Warn().Str("type", typ).Msg("No patterns available for baseline calculation")
		return nil
	}

	// 过滤相同类型的模式
	var values []float64
	for _, p := range patterns {
		if p.Type == typ {
			var rate float64
			switch typ {
			case "visit", "attack", "error":
				// 所有类型都使用RequestRate字段
				rate = p.Metrics.RequestRate
			}
			if rate > 0 {
				values = append(values, rate)
			}
		}
	}

	if int64(len(values)) < config.LearningMode.MinSamples {
		bc.logger.Warn().
			Str("type", typ).
			Int("samples", len(values)).
			Int64("required", config.LearningMode.MinSamples).
			Msg("Insufficient samples for baseline calculation")
		return nil
	}

	// 根据配置的计算方法计算基线
	var baselineValue float64
	var stdDev float64

	switch config.Baseline.CalculationMethod {
	case "mean":
		baselineValue, stdDev = bc.calculateMean(values)
	case "median":
		baselineValue = bc.calculateMedian(values)
		stdDev = bc.calculateStdDev(values, baselineValue)
	case "percentile":
		baselineValue = bc.calculatePercentile(values, int(config.Baseline.Percentile))
		stdDev = bc.calculateStdDev(values, baselineValue)
	default:
		baselineValue = bc.calculateMedian(values)
		stdDev = bc.calculateStdDev(values, baselineValue)
	}

	// 计算置信度 (基于样本数量)
	confidence := math.Min(float64(len(values))/float64(config.LearningMode.MinSamples*10), 1.0)

	baseline := &model.BaselineValue{
		Type:            typ,
		Value:           baselineValue,
		ConfidenceLevel: confidence,
		SampleSize:      int64(len(values)),
		CalculatedAt:    time.Now(),
		UpdatedAt:       time.Now(),
	}

	bc.logger.Info().
		Str("type", typ).
		Float64("value", baselineValue).
		Float64("stdDev", stdDev).
		Float64("confidence", confidence).
		Int("samples", len(values)).
		Msg("Baseline calculated")

	return baseline
}

// calculateMean 计算均值和标准差
func (bc *BaselineCalculator) calculateMean(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// 计算标准差
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	stdDev = math.Sqrt(variance / float64(len(values)))

	return mean, stdDev
}

// calculateMedian 计算中位数
func (bc *BaselineCalculator) calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// calculatePercentile 计算百分位数
func (bc *BaselineCalculator) calculatePercentile(values []float64, percentile int) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := int(float64(len(sorted)) * float64(percentile) / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// calculateStdDev 计算标准差
func (bc *BaselineCalculator) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}

	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	return math.Sqrt(variance / float64(len(values)))
}

// Save 保存基线到数据库
func (bc *BaselineCalculator) Save(baseline *model.BaselineValue) error {
	if baseline == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"type": baseline.Type}
	update := bson.M{"$set": baseline}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := bc.db.Collection("baseline_values").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		bc.logger.Error().Err(err).Str("type", baseline.Type).Msg("Failed to save baseline")
		return err
	}

	// 更新缓存
	bc.currentBaselines[baseline.Type] = baseline

	return nil
}

// GetCurrent 获取当前基线值
func (bc *BaselineCalculator) GetCurrent() map[string]*model.BaselineValue {
	// 如果缓存为空，从数据库加载
	if len(bc.currentBaselines) == 0 {
		bc.loadFromDatabase()
	}

	return bc.currentBaselines
}

// loadFromDatabase 从数据库加载基线
func (bc *BaselineCalculator) loadFromDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := bc.db.Collection("baseline_values").Find(ctx, bson.M{})
	if err != nil {
		bc.logger.Error().Err(err).Msg("Failed to load baselines")
		return
	}
	defer cursor.Close(ctx)

	var baselines []model.BaselineValue
	if err := cursor.All(ctx, &baselines); err != nil {
		bc.logger.Error().Err(err).Msg("Failed to decode baselines")
		return
	}

	for i := range baselines {
		bc.currentBaselines[baselines[i].Type] = &baselines[i]
	}

	bc.logger.Info().Int("count", len(baselines)).Msg("Baselines loaded from database")
}
