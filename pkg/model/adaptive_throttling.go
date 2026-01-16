package model

import "time"

// AdaptiveThrottlingConfig 自适应限流配置
//
//	@Description	基于流量模式的自适应限流配置
type AdaptiveThrottlingConfig struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Enabled   bool      `bson:"enabled" json:"enabled" example:"true" description:"是否启用自适应限流"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`

	// 学习模式配置
	LearningMode struct {
		Enabled          bool  `bson:"enabled" json:"enabled" example:"true" description:"是否启用学习模式"`
		LearningDuration int64 `bson:"learningDuration" json:"learningDuration" example:"86400" description:"学习周期（秒），默认24小时"`
		SampleInterval   int64 `bson:"sampleInterval" json:"sampleInterval" example:"60" description:"采样间隔（秒）"`
		MinSamples       int64 `bson:"minSamples" json:"minSamples" example:"100" description:"最小样本数"`
	} `bson:"learningMode" json:"learningMode" description:"学习模式配置"`

	// 基线配置
	Baseline struct {
		CalculationMethod string  `bson:"calculationMethod" json:"calculationMethod" example:"percentile" description:"基线计算方法: mean(均值), median(中位数), percentile(百分位数)"`
		Percentile        float64 `bson:"percentile" json:"percentile" example:"95" description:"百分位数值(0-100)"`
		UpdateInterval    int64   `bson:"updateInterval" json:"updateInterval" example:"3600" description:"基线更新间隔（秒）"`
		HistoryWindow     int64   `bson:"historyWindow" json:"historyWindow" example:"604800" description:"历史数据窗口（秒），默认7天"`
	} `bson:"baseline" json:"baseline" description:"基线计算配置"`

	// 自动调整策略
	AutoAdjustment struct {
		Enabled             bool    `bson:"enabled" json:"enabled" example:"true" description:"是否启用自动调整"`
		AnomalyThreshold    float64 `bson:"anomalyThreshold" json:"anomalyThreshold" example:"2.0" description:"异常检测阈值（倍数）"`
		MinThreshold        int64   `bson:"minThreshold" json:"minThreshold" example:"10" description:"最小限流阈值"`
		MaxThreshold        int64   `bson:"maxThreshold" json:"maxThreshold" example:"10000" description:"最大限流阈值"`
		AdjustmentFactor    float64 `bson:"adjustmentFactor" json:"adjustmentFactor" example:"1.5" description:"调整因子"`
		CooldownPeriod      int64   `bson:"cooldownPeriod" json:"cooldownPeriod" example:"300" description:"冷却期（秒）"`
		GradualAdjustment   bool    `bson:"gradualAdjustment" json:"gradualAdjustment" example:"true" description:"是否渐进式调整"`
		AdjustmentStepRatio float64 `bson:"adjustmentStepRatio" json:"adjustmentStepRatio" example:"0.1" description:"每次调整步长比例"`
	} `bson:"autoAdjustment" json:"autoAdjustment" description:"自动调整策略"`

	// 应用范围
	ApplyTo struct {
		VisitLimit  bool `bson:"visitLimit" json:"visitLimit" example:"true" description:"应用到访问限流"`
		AttackLimit bool `bson:"attackLimit" json:"attackLimit" example:"true" description:"应用到攻击限流"`
		ErrorLimit  bool `bson:"errorLimit" json:"errorLimit" example:"false" description:"应用到错误限流"`
	} `bson:"applyTo" json:"applyTo" description:"应用范围"`
}

// TrafficPattern 流量模式记录
//
//	@Description	用于存储历史流量模式数据
type TrafficPattern struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp" description:"时间戳"`
	Type      string    `bson:"type" json:"type" example:"visit" description:"类型: visit, attack, error"`

	// 流量指标
	Metrics struct {
		RequestRate  float64 `bson:"requestRate" json:"requestRate" description:"请求速率（每秒）"`
		UniqueIPs    int64   `bson:"uniqueIPs" json:"uniqueIPs" description:"唯一IP数"`
		BlockedCount int64   `bson:"blockedCount" json:"blockedCount" description:"被阻止的请求数"`
		PassedCount  int64   `bson:"passedCount" json:"passedCount" description:"通过的请求数"`
	} `bson:"metrics" json:"metrics" description:"流量指标"`

	// 统计数据
	Statistics struct {
		Mean   float64 `bson:"mean" json:"mean" description:"均值"`
		Median float64 `bson:"median" json:"median" description:"中位数"`
		StdDev float64 `bson:"stdDev" json:"stdDev" description:"标准差"`
		P95    float64 `bson:"p95" json:"p95" description:"95百分位数"`
		P99    float64 `bson:"p99" json:"p99" description:"99百分位数"`
		Min    float64 `bson:"min" json:"min" description:"最小值"`
		Max    float64 `bson:"max" json:"max" description:"最大值"`
	} `bson:"statistics" json:"statistics" description:"统计数据"`
}

// BaselineValue 基线值
//
//	@Description	当前生效的基线值
type BaselineValue struct {
	ID             string    `bson:"_id,omitempty" json:"id"`
	Type           string    `bson:"type" json:"type" example:"visit" description:"类型: visit, attack, error"`
	Value          float64   `bson:"value" json:"value" description:"基线值"`
	CalculatedAt   time.Time `bson:"calculatedAt" json:"calculatedAt" description:"计算时间"`
	SampleSize     int64     `bson:"sampleSize" json:"sampleSize" description:"样本数量"`
	ConfidenceLevel float64  `bson:"confidenceLevel" json:"confidenceLevel" description:"置信度"`
	UpdatedAt      time.Time `bson:"updatedAt" json:"updatedAt" description:"更新时间"`
}

// ThrottleAdjustmentLog 限流调整日志
//
//	@Description	记录每次自动调整的详细信息
type ThrottleAdjustmentLog struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp" description:"调整时间"`
	Type      string    `bson:"type" json:"type" example:"visit" description:"类型: visit, attack, error"`

	// 调整前的值
	OldThreshold int64   `bson:"oldThreshold" json:"oldThreshold" description:"调整前的阈值"`
	OldBaseline  float64 `bson:"oldBaseline" json:"oldBaseline" description:"调整前的基线"`

	// 调整后的值
	NewThreshold int64   `bson:"newThreshold" json:"newThreshold" description:"调整后的阈值"`
	NewBaseline  float64 `bson:"newBaseline" json:"newBaseline" description:"调整后的基线"`

	// 调整原因
	Reason          string  `bson:"reason" json:"reason" description:"调整原因"`
	CurrentTraffic  float64 `bson:"currentTraffic" json:"currentTraffic" description:"当前流量"`
	AnomalyScore    float64 `bson:"anomalyScore" json:"anomalyScore" description:"异常分数"`
	TriggeredBy     string  `bson:"triggeredBy" json:"triggeredBy" example:"auto" description:"触发方式: auto(自动), manual(手动)"`
	AdjustmentRatio float64 `bson:"adjustmentRatio" json:"adjustmentRatio" description:"调整比例"`
}

// GetCollectionName 获取自适应限流配置的集合名称
func (AdaptiveThrottlingConfig) GetCollectionName() string {
	return "adaptive_throttling_config"
}

// GetCollectionName 获取流量模式的集合名称
func (TrafficPattern) GetCollectionName() string {
	return "traffic_patterns"
}

// GetCollectionName 获取基线值的集合名称
func (BaselineValue) GetCollectionName() string {
	return "baseline_values"
}

// GetCollectionName 获取调整日志的集合名称
func (ThrottleAdjustmentLog) GetCollectionName() string {
	return "throttle_adjustment_logs"
}

// GetDefaultAdaptiveThrottlingConfig 返回默认的自适应限流配置
func GetDefaultAdaptiveThrottlingConfig() AdaptiveThrottlingConfig {
	now := time.Now()
	config := AdaptiveThrottlingConfig{
		Enabled:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 学习模式默认配置
	config.LearningMode.Enabled = true
	config.LearningMode.LearningDuration = 86400 // 24小时
	config.LearningMode.SampleInterval = 60      // 1分钟
	config.LearningMode.MinSamples = 100

	// 基线配置
	config.Baseline.CalculationMethod = "percentile"
	config.Baseline.Percentile = 95.0
	config.Baseline.UpdateInterval = 3600  // 1小时
	config.Baseline.HistoryWindow = 604800 // 7天

	// 自动调整策略
	config.AutoAdjustment.Enabled = true
	config.AutoAdjustment.AnomalyThreshold = 2.0
	config.AutoAdjustment.MinThreshold = 10
	config.AutoAdjustment.MaxThreshold = 10000
	config.AutoAdjustment.AdjustmentFactor = 1.5
	config.AutoAdjustment.CooldownPeriod = 300
	config.AutoAdjustment.GradualAdjustment = true
	config.AutoAdjustment.AdjustmentStepRatio = 0.1

	// 应用范围
	config.ApplyTo.VisitLimit = true
	config.ApplyTo.AttackLimit = true
	config.ApplyTo.ErrorLimit = false

	return config
}
