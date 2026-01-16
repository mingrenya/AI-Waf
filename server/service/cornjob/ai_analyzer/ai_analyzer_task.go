// server/service/cornjob/ai_analyzer/ai_analyzer_task.go
package ai_analyzer

import (
	"context"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AIAnalyzerTask AI分析定时任务
type AIAnalyzerTask struct {
	engine *service.AIEngine
	cron   *cron.Cron
	logger zerolog.Logger
}

// NewAIAnalyzerTask 创建AI分析定时任务
func NewAIAnalyzerTask(db *mongo.Database) *AIAnalyzerTask {
	logger := config.GetLogger().With().Str("component", "ai-analyzer-task").Logger()
	
	return &AIAnalyzerTask{
		engine: service.NewAIEngine(db),
		cron:   cron.New(),
		logger: logger,
	}
}

// Start 启动定时任务
func (t *AIAnalyzerTask) Start() error {
	t.logger.Info().Msg("Starting AI Analyzer cron tasks")
	
	// 每小时执行一次攻击模式检测
	_, err := t.cron.AddFunc("0 * * * *", func() {
		t.logger.Info().Msg("Running hourly attack pattern detection")
		if err := t.detectAndGenerateRules(); err != nil {
			t.logger.Error().Err(err).Msg("Failed to run attack pattern detection")
		}
	})
	if err != nil {
		return err
	}
	
	// 每天凌晨2点清理旧数据
	_, err = t.cron.AddFunc("0 2 * * *", func() {
		t.logger.Info().Msg("Running daily cleanup")
		if err := t.cleanup(); err != nil {
			t.logger.Error().Err(err).Msg("Failed to run cleanup")
		}
	})
	if err != nil {
		return err
	}
	
	t.cron.Start()
	t.logger.Info().Msg("AI Analyzer cron tasks started successfully")
	
	return nil
}

// Stop 停止定时任务
func (t *AIAnalyzerTask) Stop() {
	t.logger.Info().Msg("Stopping AI Analyzer cron tasks")
	t.cron.Stop()
}

// detectAndGenerateRules 检测攻击模式并生成规则
func (t *AIAnalyzerTask) detectAndGenerateRules() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	// 1. 运行攻击模式检测
	patterns, err := t.engine.RunAttackPatternDetection(ctx)
	if err != nil {
		return err
	}
	
	if len(patterns) == 0 {
		t.logger.Info().Msg("No new attack patterns detected")
		return nil
	}
	
	// 2. 为检测到的高危模式自动生成规则
	var highRiskPatterns []*model.AttackPattern
	for i := range patterns {
		if patterns[i].Severity == "high" || patterns[i].Severity == "critical" {
			highRiskPatterns = append(highRiskPatterns, patterns[i])
		}
	}
	
	if len(highRiskPatterns) > 0 {
		t.logger.Info().
			Int("high_risk_count", len(highRiskPatterns)).
			Msg("Generating rules for high-risk patterns")
		
		_, err := t.engine.GenerateRulesForPatterns(ctx, highRiskPatterns, 0.7)
		if err != nil {
			t.logger.Error().Err(err).Msg("Failed to generate rules for high-risk patterns")
			return err
		}
	}
	
	return nil
}

// cleanup 清理旧数据
func (t *AIAnalyzerTask) cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	// 删除30天前的已拒绝规则
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	db := t.engine.GetDB()
	result, err := db.Collection("generated_rules").DeleteMany(ctx, map[string]interface{}{
		"status": "rejected",
		"created_at": map[string]interface{}{
			"$lt": thirtyDaysAgo,
		},
	})
	
	if err != nil {
		return err
	}
	
	t.logger.Info().
		Int64("deleted_count", result.DeletedCount).
		Msg("Cleaned up rejected rules older than 30 days")
	
	return nil
}

// RunNow 立即运行一次检测（用于手动触发）
func (t *AIAnalyzerTask) RunNow() error {
	t.logger.Info().Msg("Manually triggering attack pattern detection")
	return t.detectAndGenerateRules()
}
