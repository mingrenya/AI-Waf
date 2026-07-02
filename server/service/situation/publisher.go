package situation

import (
	"context"
	"encoding/json"

	"github.com/mingrenya/AI-Waf/server/config"
	ws "github.com/mingrenya/AI-Waf/server/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Publisher struct {
	redis  *redis.Client
	logger zerolog.Logger
}

func NewPublisher(redis *redis.Client) *Publisher {
	return &Publisher{
		redis:  redis,
		logger: config.GetServiceLogger("situation-publisher"),
	}
}

// PublishAlert 发布攻击链告警
func (p *Publisher) PublishAlert(chain interface{}) {
	data, _ := json.Marshal(chain)
	p.publishRedis(context.Background(), "situation:alert", data)
	ws.GetHub().BroadcastJSON("situation:alert", chain)
}

// PublishUpdate 发布态势数据更新
func (p *Publisher) PublishUpdate(update interface{}) {
	data, _ := json.Marshal(update)
	p.publishRedis(context.Background(), "situation:update", data)
	ws.GetHub().BroadcastJSON("situation:update", update)
}

// PublishAttack 发布实时攻击事件
func (p *Publisher) PublishAttack(event interface{}) {
	data, _ := json.Marshal(event)
	p.publishRedis(context.Background(), "situation:attack", data)
	ws.GetHub().BroadcastJSON("situation:attack", event)
}

func (p *Publisher) publishRedis(ctx context.Context, channel string, data []byte) {
	if p.redis == nil {
		return
	}
	if err := p.redis.Publish(ctx, channel, string(data)).Err(); err != nil {
		p.logger.Warn().Err(err).Str("channel", channel).Msg("Redis publish failed")
	}
}
