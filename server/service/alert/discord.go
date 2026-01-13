package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mingrenya/AI-Waf/server/model"
)

// DiscordSender Discord 告警发送器
type DiscordSender struct {
	client *http.Client
}

// NewDiscordSender 创建 Discord 发送器
func NewDiscordSender() Sender {
	return &DiscordSender{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *DiscordSender) Send(ctx context.Context, channel *model.AlertChannel, message string) error {
	webhookURL, ok := channel.Config["webhookUrl"].(string)
	if !ok {
		return fmt.Errorf("discord webhook URL not configured")
	}

	// 构建 Discord 消息格式
	payload := map[string]interface{}{
		"content": message,
	}

	// 添加可选配置
	if username, ok := channel.Config["username"].(string); ok && username != "" {
		payload["username"] = username
	}

	if avatarURL, ok := channel.Config["avatarUrl"].(string); ok && avatarURL != "" {
		payload["avatar_url"] = avatarURL
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send discord message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *DiscordSender) GetType() string {
	return model.AlertChannelTypeDiscord
}

func (s *DiscordSender) Validate(config map[string]interface{}) error {
	webhookURL, ok := config["webhookUrl"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhookUrl is required")
	}

	return nil
}
