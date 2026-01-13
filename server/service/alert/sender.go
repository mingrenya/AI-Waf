package alert

import (
	"context"

	"github.com/mingrenya/AI-Waf/server/model"
)

// Sender 告警发送器接口
type Sender interface {
	// Send 发送告警消息
	Send(ctx context.Context, channel *model.AlertChannel, message string) error
	
	// GetType 获取发送器类型
	GetType() string
	
	// Validate 验证渠道配置
	Validate(config map[string]interface{}) error
}
