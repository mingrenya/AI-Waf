// server/service/ai_engine.go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/coraza-spoa/analyzer"
	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AIEngine AI分析引擎服务
type AIEngine struct {
	detector  *analyzer.AttackPatternDetector
	generator *analyzer.RuleGenerator
	db        *mongo.Database
	logger    zerolog.Logger
}

// NewAIEngine 创建AI分析引擎
func NewAIEngine(db *mongo.Database) *AIEngine {
	logger := config.GetServiceLogger("ai_engine")
	simpleLogger := &SimpleLogger{logger: logger}
	
	return &AIEngine{
		detector:  analyzer.NewAttackPatternDetector(db, logger),
		generator: analyzer.NewRuleGenerator(db, simpleLogger),
		db:        db,
		logger:    logger,
	}
}

// SimpleLogger 简单的日志实现
type SimpleLogger struct {
	logger zerolog.Logger
}

func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *SimpleLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// RunAttackPatternDetection 运行攻击模式检测
func (e *AIEngine) RunAttackPatternDetection(ctx context.Context) ([]*model.AttackPattern, error) {
	e.logger.Info().Msg("Starting attack pattern detection...")
	
	patterns, err := e.detector.DetectPatterns()
	if err != nil {
		e.logger.Error().Err(err).Msg("Failed to detect patterns")
		return nil, fmt.Errorf("检测攻击模式失败: %w", err)
	}
	
	e.logger.Info().Int("count", len(patterns)).Msg("Attack patterns detected")
	return patterns, nil
}

// GenerateRulesForPattern 为特定模式生成规则
func (e *AIEngine) GenerateRulesForPattern(ctx context.Context, pattern *model.AttackPattern, threshold float64) ([]*model.GeneratedRule, error) {
	e.logger.Info().
		Str("pattern_id", pattern.ID.Hex()).
		Str("pattern_type", pattern.PatternType).
		Msg("Generating rules for pattern")
	
	rules, err := e.generator.GenerateRules([]*model.AttackPattern{pattern}, threshold)
	if err != nil {
		e.logger.Error().Err(err).Msg("Failed to generate rules")
		return nil, fmt.Errorf("生成规则失败: %w", err)
	}
	
	e.logger.Info().Int("count", len(rules)).Msg("Rules generated")
	return rules, nil
}

// GenerateRulesForPatterns 为多个模式批量生成规则
func (e *AIEngine) GenerateRulesForPatterns(ctx context.Context, patterns []*model.AttackPattern, threshold float64) ([]*model.GeneratedRule, error) {
	e.logger.Info().
		Int("pattern_count", len(patterns)).
		Float64("threshold", threshold).
		Msg("Batch generating rules")
	
	rules, err := e.generator.GenerateRules(patterns, threshold)
	if err != nil {
		e.logger.Error().Err(err).Msg("Failed to batch generate rules")
		return nil, fmt.Errorf("批量生成规则失败: %w", err)
	}
	
	e.logger.Info().Int("rule_count", len(rules)).Msg("Batch rules generated")
	return rules, nil
}

// GetDetectionStats 获取检测统计信息
func (e *AIEngine) GetDetectionStats(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 统计攻击模式
	patternCount, err := e.db.Collection("attack_patterns").CountDocuments(ctx, map[string]interface{}{
		"created_at": map[string]interface{}{
			"$gte": startTime,
			"$lte": endTime,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("统计攻击模式失败: %w", err)
	}
	stats["pattern_count"] = patternCount
	
	// 统计生成的规则
	ruleCount, err := e.db.Collection("generated_rules").CountDocuments(ctx, map[string]interface{}{
		"created_at": map[string]interface{}{
			"$gte": startTime,
			"$lte": endTime,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("统计生成规则失败: %w", err)
	}
	stats["rule_count"] = ruleCount
	
	// 统计已部署的规则
	deployedCount, err := e.db.Collection("generated_rules").CountDocuments(ctx, map[string]interface{}{
		"status": "deployed",
		"created_at": map[string]interface{}{
			"$gte": startTime,
			"$lte": endTime,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("统计已部署规则失败: %w", err)
	}
	stats["deployed_rule_count"] = deployedCount
	
	return stats, nil
}

// GetDB 获取数据库实例（用于定时任务）
func (e *AIEngine) GetDB() *mongo.Database {
	return e.db
}
