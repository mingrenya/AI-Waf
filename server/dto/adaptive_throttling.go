package dto

import "time"

// AdaptiveThrottlingConfigRequest 自适应限流配置请求
type AdaptiveThrottlingConfigRequest struct {
	Enabled        bool                      `json:"enabled"`                           // 是否启用
	LearningMode   LearningModeConfigDTO     `json:"learningMode" binding:"required"`   // 学习模式配置
	Baseline       BaselineConfigDTO         `json:"baseline" binding:"required"`       // 基线配置
	AutoAdjustment AutoAdjustmentConfigDTO   `json:"autoAdjustment" binding:"required"` // 自动调整配置
	ApplyTo        ApplyToConfigDTO          `json:"applyTo" binding:"required"`        // 应用范围配置
}

// LearningModeConfigDTO 学习模式配置DTO
type LearningModeConfigDTO struct {
	Enabled          bool  `json:"enabled"`                                      // 是否启用学习模式
	LearningDuration int64 `json:"learningDuration" binding:"required,min=3600"` // 学习周期(秒)
	SampleInterval   int64 `json:"sampleInterval" binding:"required,min=10"`     // 采样间隔(秒)
	MinSamples       int   `json:"minSamples" binding:"required,min=10"`         // 最小样本数
}

// BaselineConfigDTO 基线配置DTO
type BaselineConfigDTO struct {
	CalculationMethod string `json:"calculationMethod" binding:"required,oneof=mean median percentile"` // 计算方法
	Percentile        int    `json:"percentile" binding:"required,min=0,max=100"`                       // 百分位数
	UpdateInterval    int64  `json:"updateInterval" binding:"required,min=60"`                          // 更新间隔(秒)
	HistoryWindow     int64  `json:"historyWindow" binding:"required,min=86400"`                        // 历史窗口(秒)
}

// AutoAdjustmentConfigDTO 自动调整配置DTO
type AutoAdjustmentConfigDTO struct {
	Enabled              bool    `json:"enabled"`                                    // 是否启用自动调整
	AnomalyThreshold     float64 `json:"anomalyThreshold" binding:"required,min=1"` // 异常阈值倍数
	MinThreshold         int     `json:"minThreshold" binding:"required,min=1"`     // 最小阈值
	MaxThreshold         int     `json:"maxThreshold" binding:"required,min=100"`   // 最大阈值
	AdjustmentFactor     float64 `json:"adjustmentFactor" binding:"required,min=1"` // 调整因子
	CooldownPeriod       int64   `json:"cooldownPeriod" binding:"required,min=60"`  // 冷却期(秒)
	GradualAdjustment    bool    `json:"gradualAdjustment"`                          // 是否渐进式调整
	AdjustmentStepRatio  float64 `json:"adjustmentStepRatio" binding:"required,min=0.01"` // 调整步长比例
}

// ApplyToConfigDTO 应用范围配置DTO
type ApplyToConfigDTO struct {
	VisitLimit  bool `json:"visitLimit"`  // 应用到访问限流
	AttackLimit bool `json:"attackLimit"` // 应用到攻击限流
	ErrorLimit  bool `json:"errorLimit"`  // 应用到错误限流
}

// TrafficPatternQuery 流量模式查询参数
type TrafficPatternQuery struct {
	Type      string    `form:"type" binding:"omitempty,oneof=visit attack error"` // 类型筛选
	StartTime time.Time `form:"startTime" binding:"omitempty"`                      // 开始时间
	EndTime   time.Time `form:"endTime" binding:"omitempty"`                        // 结束时间
	Page      int       `form:"page" binding:"omitempty,min=1"`                     // 页码
	PageSize  int       `form:"pageSize" binding:"omitempty,min=1,max=100"`         // 每页数量
}

// BaselineQuery 基线查询参数
type BaselineQuery struct {
	Type string `form:"type" binding:"omitempty,oneof=visit attack error"` // 类型筛选
}

// AdjustmentLogQuery 调整日志查询参数
type AdjustmentLogQuery struct {
	Type      string    `form:"type" binding:"omitempty,oneof=visit attack error"` // 类型筛选
	StartTime time.Time `form:"startTime" binding:"omitempty"`                      // 开始时间
	EndTime   time.Time `form:"endTime" binding:"omitempty"`                        // 结束时间
	Page      int       `form:"page" binding:"omitempty,min=1"`                     // 页码
	PageSize  int       `form:"pageSize" binding:"omitempty,min=1,max=100"`         // 每页数量
}

// TrafficPatternResponse 流量模式响应
type TrafficPatternResponse struct {
	Results      []TrafficPatternDTO `json:"results"`      // 流量模式列表
	TotalCount   int                 `json:"totalCount"`   // 总数
	CurrentPage  int                 `json:"currentPage"`  // 当前页
	PageSize     int                 `json:"pageSize"`     // 每页数量
	TotalPages   int                 `json:"totalPages"`   // 总页数
}

// TrafficPatternDTO 流量模式DTO
type TrafficPatternDTO struct {
	Type       string                `json:"type"`       // 类型
	Timestamp  time.Time             `json:"timestamp"`  // 时间戳
	Metrics    TrafficMetricsDTO     `json:"metrics"`    // 流量指标
	Statistics TrafficStatisticsDTO  `json:"statistics"` // 统计信息
}

// TrafficMetricsDTO 流量指标DTO
type TrafficMetricsDTO struct {
	RequestCount int     `json:"requestCount"` // 请求数
	AvgLatency   float64 `json:"avgLatency"`   // 平均延迟
	ErrorRate    float64 `json:"errorRate"`    // 错误率
	P95Latency   float64 `json:"p95Latency"`   // P95延迟
	P99Latency   float64 `json:"p99Latency"`   // P99延迟
}

// TrafficStatisticsDTO 流量统计DTO
type TrafficStatisticsDTO struct {
	Mean   float64 `json:"mean"`   // 均值
	Median float64 `json:"median"` // 中位数
	StdDev float64 `json:"stdDev"` // 标准差
	Min    float64 `json:"min"`    // 最小值
	Max    float64 `json:"max"`    // 最大值
}

// BaselineResponse 基线响应
type BaselineResponse struct {
	Results []BaselineValueDTO `json:"results"` // 基线列表
}

// BaselineValueDTO 基线值DTO
type BaselineValueDTO struct {
	Type            string    `json:"type"`            // 类型
	Value           float64   `json:"value"`           // 基线值
	ConfidenceLevel float64   `json:"confidenceLevel"` // 置信度
	SampleSize      int       `json:"sampleSize"`      // 样本数量
	CalculatedAt    time.Time `json:"calculatedAt"`    // 计算时间
	UpdatedAt       time.Time `json:"updatedAt"`       // 更新时间
}

// AdjustmentLogResponse 调整日志响应
type AdjustmentLogResponse struct {
	Results      []ThrottleAdjustmentLogDTO `json:"results"`      // 调整日志列表
	TotalCount   int                        `json:"totalCount"`   // 总数
	CurrentPage  int                        `json:"currentPage"`  // 当前页
	PageSize     int                        `json:"pageSize"`     // 每页数量
	TotalPages   int                        `json:"totalPages"`   // 总页数
}

// ThrottleAdjustmentLogDTO 调整日志DTO
type ThrottleAdjustmentLogDTO struct {
	ID               string    `json:"id"`               // ID
	Type             string    `json:"type"`             // 类型
	Timestamp        time.Time `json:"timestamp"`        // 时间戳
	OldThreshold     int       `json:"oldThreshold"`     // 旧阈值
	NewThreshold     int       `json:"newThreshold"`     // 新阈值
	AdjustmentRatio  float64   `json:"adjustmentRatio"`  // 调整比例
	OldBaseline      float64   `json:"oldBaseline"`      // 旧基线
	NewBaseline      float64   `json:"newBaseline"`      // 新基线
	CurrentTraffic   float64   `json:"currentTraffic"`   // 当前流量
	AnomalyScore     float64   `json:"anomalyScore"`     // 异常分数
	Reason           string    `json:"reason"`           // 原因
	TriggeredBy      string    `json:"triggeredBy"`      // 触发方式
}

// AdaptiveThrottlingStatsDTO 自适应限流统计DTO
type AdaptiveThrottlingStatsDTO struct {
	CurrentBaseline   BaselineStatsDTO   `json:"currentBaseline"`   // 当前基线
	CurrentThreshold  ThresholdStatsDTO  `json:"currentThreshold"`  // 当前阈值
	LearningProgress  float64            `json:"learningProgress"`  // 学习进度
	RecentAdjustments int                `json:"recentAdjustments"` // 近期调整次数
	AnomalyDetected   bool               `json:"anomalyDetected"`   // 是否检测到异常
	LastUpdateTime    time.Time          `json:"lastUpdateTime"`    // 最后更新时间
}

// BaselineStatsDTO 基线统计DTO
type BaselineStatsDTO struct {
	Visit  float64 `json:"visit"`  // 访问基线
	Attack float64 `json:"attack"` // 攻击基线
	Error  float64 `json:"error"`  // 错误基线
}

// ThresholdStatsDTO 阈值统计DTO
type ThresholdStatsDTO struct {
	Visit  int `json:"visit"`  // 访问阈值
	Attack int `json:"attack"` // 攻击阈值
	Error  int `json:"error"`  // 错误阈值
}
