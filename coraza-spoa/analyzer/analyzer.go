package analyzer

import (
	"context"
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// zeroLogAdapter 适配zerolog到Logger接口
type zeroLogAdapter struct {
	logger zerolog.Logger
}

func (z *zeroLogAdapter) Infof(format string, args ...interface{}) {
	z.logger.Info().Msgf(format, args...)
}

func (z *zeroLogAdapter) Warnf(format string, args ...interface{}) {
	z.logger.Warn().Msgf(format, args...)
}

func (z *zeroLogAdapter) Errorf(format string, args ...interface{}) {
	z.logger.Error().Msgf(format, args...)
}

func (z *zeroLogAdapter) Debugf(format string, args ...interface{}) {
	z.logger.Debug().Msgf(format, args...)
}

// AISecurityAnalyzer AI安全分析器 - 基于机器学习的攻击模式检测和规则生成
type AISecurityAnalyzer struct {
	db     *mongo.Database
	logger zerolog.Logger

	// 攻击模式检测器
	patternDetector *AttackPatternDetector

	// 规则生成器
	ruleGenerator *RuleGenerator

	// 特征提取器
	featureExtractor *FeatureExtractor

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 配置
	config     *AIAnalyzerConfig
	configLock sync.RWMutex
}

// AIAnalyzerConfig AI分析器配置
type AIAnalyzerConfig struct {
	Enabled bool // 是否启用AI分析

	// 模式检测配置
	PatternDetection struct {
		Enabled          bool    // 是否启用模式检测
		MinSamples       int     // 最小样本数
		AnomalyThreshold float64 // 异常阈值
		ClusteringMethod string  // 聚类方法: "kmeans", "dbscan"
	}

	// 规则生成配置
	RuleGeneration struct {
		Enabled           bool    // 是否启用规则生成
		ConfidenceThreshold float64 // 置信度阈值
		AutoDeploy        bool    // 是否自动部署
		ReviewRequired    bool    // 是否需要人工审核
	}

	// 分析周期
	AnalysisInterval time.Duration // 分析间隔
}

// NewAISecurityAnalyzer 创建AI安全分析器
func NewAISecurityAnalyzer(db *mongo.Database, logger zerolog.Logger) *AISecurityAnalyzer {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建适配的Logger
	loggerAdapter := &zeroLogAdapter{logger: logger}

	analyzer := &AISecurityAnalyzer{
		db:               db,
		logger:           logger.With().Str("component", "ai-analyzer").Logger(),
		patternDetector:  NewAttackPatternDetector(db, logger),
		ruleGenerator:    NewRuleGenerator(db, loggerAdapter),
		featureExtractor: NewFeatureExtractor(),
		ctx:              ctx,
		cancel:           cancel,
		config:           getDefaultConfig(),
	}

	// 启动后台分析任务
	analyzer.start()

	return analyzer
}

// start 启动后台任务
func (a *AISecurityAnalyzer) start() {
	a.logger.Info().Msg("启动AI安全分析器")

	// 定期分析任务
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(a.config.AnalysisInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				a.performAnalysis()
			case <-a.ctx.Done():
				return
			}
		}
	}()
}

// performAnalysis 执行分析
func (a *AISecurityAnalyzer) performAnalysis() {
	a.configLock.RLock()
	config := a.config
	a.configLock.RUnlock()

	if !config.Enabled {
		return
	}

	a.logger.Debug().Msg("开始AI安全分析")

	// 1. 检测攻击模式
	if config.PatternDetection.Enabled {
		patterns, err := a.patternDetector.DetectPatterns()
		if err != nil {
			a.logger.Error().Err(err).Msg("攻击模式检测失败")
			return
		}

		a.logger.Info().Int("count", len(patterns)).Msg("检测到攻击模式")

		// 2. 生成防护规则
		if config.RuleGeneration.Enabled && len(patterns) > 0 {
			rules, err := a.ruleGenerator.GenerateRules(patterns, config.RuleGeneration.ConfidenceThreshold)
			if err != nil {
				a.logger.Error().Err(err).Msg("规则生成失败")
				return
			}

			a.logger.Info().Int("count", len(rules)).Msg("生成防护规则")

			// 3. 保存规则（等待审核或自动部署）
			for _, rule := range rules {
				if err := a.saveGeneratedRule(rule); err != nil {
					a.logger.Error().Err(err).Str("ruleId", rule.ID.Hex()).Msg("保存规则失败")
				}
			}
		}
	}
}

// saveGeneratedRule 保存生成的规则
func (a *AISecurityAnalyzer) saveGeneratedRule(rule *model.GeneratedRule) error {
	return a.ruleGenerator.SaveGeneratedRule(rule)
}

// Stop 停止分析器
func (a *AISecurityAnalyzer) Stop() {
	a.logger.Info().Msg("停止AI安全分析器")
	a.cancel()
	a.wg.Wait()
}

// GetStats 获取统计信息
func (a *AISecurityAnalyzer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":          a.config.Enabled,
		"patternsDetected": 0, // TODO: 实现实际统计
		"rulesGenerated":   0,
		"rulesDeployed":    0,
	}
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *AIAnalyzerConfig {
	config := &AIAnalyzerConfig{
		Enabled:          false,
		AnalysisInterval: 1 * time.Hour, // 每小时分析一次
	}

	config.PatternDetection.Enabled = true
	config.PatternDetection.MinSamples = 100
	config.PatternDetection.AnomalyThreshold = 2.0
	config.PatternDetection.ClusteringMethod = "kmeans"

	config.RuleGeneration.Enabled = true
	config.RuleGeneration.ConfidenceThreshold = 0.8
	config.RuleGeneration.AutoDeploy = false
	config.RuleGeneration.ReviewRequired = true

	return config
}
