package trafficanalyzer

import (
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/rs/zerolog"
)

// AnomalyDetector 异常检测器
type AnomalyDetector struct {
	logger zerolog.Logger
}

// Anomaly 异常信息
type Anomaly struct {
	Type           string    // "visit", "attack", "error"
	CurrentValue   float64   // 当前值
	BaselineValue  float64   // 基线值
	AnomalyScore   float64   // 异常分数 (当前值/基线值)
	Severity       string    // "low", "medium", "high", "critical"
	DetectedAt     time.Time
	Reason         string
}

// NewAnomalyDetector 创建异常检测器
func NewAnomalyDetector(logger zerolog.Logger) *AnomalyDetector {
	return &AnomalyDetector{
		logger: logger.With().Str("component", "anomaly-detector").Logger(),
	}
}

// Detect 检测异常
func (ad *AnomalyDetector) Detect(
	currentMetrics map[string]float64,
	baselines map[string]*model.BaselineValue,
	config *model.AdaptiveThrottlingConfig,
) []Anomaly {
	var anomalies []Anomaly

	// 检查各类型的流量
	types := []string{"visit", "attack", "error"}
	
	for _, typ := range types {
		baseline, exists := baselines[typ]
		if !exists || baseline == nil {
			continue
		}

		current, exists := currentMetrics[typ]
		if !exists {
			continue
		}

		// 计算异常分数
		anomalyScore := current / baseline.Value
		
		// 判断是否超过阈值
		if anomalyScore > config.AutoAdjustment.AnomalyThreshold {
			anomaly := Anomaly{
				Type:          typ,
				CurrentValue:  current,
				BaselineValue: baseline.Value,
				AnomalyScore:  anomalyScore,
				DetectedAt:    time.Now(),
				Reason:        ad.generateReason(typ, anomalyScore),
			}

			// 确定严重程度
			anomaly.Severity = ad.calculateSeverity(anomalyScore, config.AutoAdjustment.AnomalyThreshold)

			anomalies = append(anomalies, anomaly)

			ad.logger.Warn().
				Str("type", typ).
				Float64("current", current).
				Float64("baseline", baseline.Value).
				Float64("score", anomalyScore).
				Str("severity", anomaly.Severity).
				Msg("Anomaly detected")
		}
	}

	return anomalies
}

// generateReason 生成异常原因描述
func (ad *AnomalyDetector) generateReason(typ string, score float64) string {
	var typeDesc string
	switch typ {
	case "visit":
		typeDesc = "访问流量"
	case "attack":
		typeDesc = "攻击流量"
	case "error":
		typeDesc = "错误流量"
	default:
		typeDesc = "流量"
	}

	if score < 2.0 {
		return typeDesc + "略高于正常水平"
	} else if score < 3.0 {
		return typeDesc + "明显高于正常水平"
	} else if score < 5.0 {
		return typeDesc + "显著高于正常水平"
	} else {
		return typeDesc + "极度异常"
	}
}

// calculateSeverity 计算严重程度
func (ad *AnomalyDetector) calculateSeverity(score float64, threshold float64) string {
	ratio := score / threshold

	if ratio < 1.5 {
		return "low"
	} else if ratio < 2.5 {
		return "medium"
	} else if ratio < 4.0 {
		return "high"
	} else {
		return "critical"
	}
}

// DetectWithStdDev 使用标准差进行异常检测
func (ad *AnomalyDetector) DetectWithStdDev(
	currentValue float64,
	baseline *model.BaselineValue,
	sigmaMultiplier float64,
) bool {
	if baseline == nil {
		return false
	}

	// 使用倍数阈值规则（因为BaselineValue不包含StandardDev）
	threshold := baseline.Value * sigmaMultiplier
	return currentValue > threshold
}

// IsLearningComplete 判断学习是否完成
func (ad *AnomalyDetector) IsLearningComplete(
	baseline *model.BaselineValue,
	config *model.AdaptiveThrottlingConfig,
) bool {
	if baseline == nil {
		return false
	}

	// 检查样本数量
	if baseline.SampleSize < int64(config.LearningMode.MinSamples) {
		return false
	}

	// 检查置信度
	if baseline.ConfidenceLevel < 0.95 {
		return false
	}

	// 检查学习时长
	learningDuration := time.Since(baseline.CalculatedAt)
	requiredDuration := time.Duration(config.LearningMode.LearningDuration) * time.Second
	if learningDuration < requiredDuration {
		return false
	}

	return true
}
