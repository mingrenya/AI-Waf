package ftw

import (
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Scheduler FTW 定时回归测试调度器
type Scheduler struct {
	service service.FTWTestService
	cron    *cron.Cron
	logger  zerolog.Logger
}

// NewScheduler creates a new FTW cron scheduler
func NewScheduler(db *mongo.Database) *Scheduler {
	return &Scheduler{
		service: service.NewFTWTestService(db),
		cron:    cron.New(),
		logger:  config.GetLogger().With().Str("component", "ftw-cron").Logger(),
	}
}

// Start 启动定时任务：每天凌晨 3 点自动执行回归测试
func (s *Scheduler) Start() error {
	s.logger.Info().Msg("启动 FTW 定时回归测试...")

	// 每天凌晨 3 点执行
	_, err := s.cron.AddFunc("0 3 * * *", func() {
		s.logger.Info().Msg("开始执行每日 FTW 回归测试")
		report, err := s.service.RunTests(nil, "http://localhost:8080", "")
		if err != nil {
			s.logger.Error().Err(err).Msg("每日 FTW 回归测试失败")
			return
		}
		s.logger.Info().
			Int("total", report.TotalTests).
			Int("passed", report.Passed).
			Float64("blockRate", report.BlockRate).
			Msg(fmt.Sprintf("每日 FTW 回归测试完成：拦截率 %.1f%%", report.BlockRate))
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.Info().Msg("FTW 定时回归测试已启动（每天凌晨 3:00）")
	return nil
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	timer := time.NewTimer(5 * time.Second)
	done := make(chan struct{})
	go func() {
		s.cron.Stop()
		close(done)
	}()
	select {
	case <-done:
		s.logger.Info().Msg("FTW 定时任务已停止")
	case <-timer.C:
		s.logger.Warn().Msg("FTW 定时任务停止超时")
	}
}

// RunNow 立即执行一次回归测试
func (s *Scheduler) RunNow() error {
	s.logger.Info().Msg("手动触发 FTW 回归测试")
	report, err := s.service.RunTests(nil, "http://localhost:8080", "")
	if err != nil {
		return err
	}
	s.logger.Info().Msgf("FTW 回归测试完成：%d/%d 通过，拦截率 %.1f%%",
		report.Passed, report.TotalTests, report.BlockRate)
	return nil
}
