package alert

import (
	"context"
	"time"

	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/rs/zerolog"
)

// Start 启动告警检查定时任务
func Start(alertService service.AlertService, logger zerolog.Logger) (func(), error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建定时器，每分钟检查一次
	ticker := time.NewTicker(1 * time.Minute)

	// 启动后台 goroutine
	go func() {
		logger.Info().Msg("Alert checker started")

		for {
			select {
			case <-ticker.C:
				if err := alertService.CheckAndTriggerAlerts(ctx); err != nil {
					logger.Error().Err(err).Msg("Failed to check and trigger alerts")
				}
			case <-ctx.Done():
				logger.Info().Msg("Alert checker stopped")
				return
			}
		}
	}()

	// 返回清理函数
	cleanup := func() {
		logger.Info().Msg("Stopping alert checker...")
		ticker.Stop()
		cancel()
	}

	return cleanup, nil
}
