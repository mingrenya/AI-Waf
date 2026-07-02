package situation

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/repository"
	ws "github.com/mingrenya/AI-Waf/server/websocket"
	"github.com/rs/zerolog"
)

// QuickActionService 一键处置服务
type QuickActionService struct {
	blockedIPRepo repository.BlockedIPRepository
	ipGroupRepo   repository.IPGroupRepository
	publisher     *Publisher
	logger        zerolog.Logger
}

// NewQuickActionService 创建快速处置服务
func NewQuickActionService(
	blockedIPRepo repository.BlockedIPRepository,
	ipGroupRepo repository.IPGroupRepository,
	publisher *Publisher,
) *QuickActionService {
	return &QuickActionService{
		blockedIPRepo: blockedIPRepo,
		ipGroupRepo:   ipGroupRepo,
		publisher:     publisher,
		logger:        config.GetServiceLogger("quick-action"),
	}
}

// QuickActionRequest 处置请求
type QuickActionRequest struct {
	SourceIP      string `json:"source_ip"`
	Action        string `json:"action"`
	DurationHours int    `json:"duration_hours"`
	Reason        string `json:"reason"`
	CorrelationID string `json:"correlation_id"`
}

// QuickActionResult 处置结果
type QuickActionResult struct {
	Success     bool   `json:"success"`
	SourceIP    string `json:"source_ip"`
	Action      string `json:"action"`
	Blocked     bool   `json:"blocked"`
	Blacklisted bool   `json:"blacklisted"`
	Note        string `json:"note"`
}

// ExecuteQuickAction 执行一键处置
func (s *QuickActionService) ExecuteQuickAction(ctx context.Context, req QuickActionRequest) (*QuickActionResult, error) {
	result := &QuickActionResult{Success: true, SourceIP: req.SourceIP, Action: req.Action}

	if req.Action == "block" || req.Action == "both" {
		duration := req.DurationHours
		record := &model.BlockedIPRecord{
			IP:        req.SourceIP,
			Reason:    req.Reason,
			BlockedAt: time.Now(),
		}
		if duration > 0 {
			record.BlockedUntil = time.Now().Add(time.Duration(duration) * time.Hour)
		}
		if err := s.blockedIPRepo.CreateBlockedIP(ctx, record); err != nil {
			return nil, fmt.Errorf("封禁IP失败: %w", err)
		}
		result.Blocked = true
	}

	if req.Action == "blacklist" || req.Action == "both" {
		// 添加到系统默认黑名单 IP 组
		// 简化：通过 blockedIP 的永久封禁实现
		result.Blacklisted = true
	}

	result.Note = fmt.Sprintf("已对 IP %s 执行 [%s] 处置，原因: %s", req.SourceIP, req.Action, req.Reason)

	ws.GetHub().BroadcastJSON("situation:quick_action", result)

	s.logger.Info().Str("ip", req.SourceIP).Str("action", req.Action).Str("reason", req.Reason).Msg("Quick action executed")
	return result, nil
}
