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

// SlackSender Slack 告警发送器
type SlackSender struct {
	client *http.Client
}

// NewSlackSender 创建 Slack 发送器
func NewSlackSender() Sender {
	return &SlackSender{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *SlackSender) Send(ctx context.Context, channel *model.AlertChannel, message string) error {
	webhookURL, ok := channel.Config["webhookUrl"].(string)
	if !ok {
		return fmt.Errorf("slack webhook URL not configured")
	}

	// 构建 Slack 消息格式
	payload := map[string]interface{}{
		"text": message,
	}

	// 添加可选配置
	if channelName, ok := channel.Config["channel"].(string); ok && channelName != "" {
		payload["channel"] = channelName
	}

	if username, ok := channel.Config["username"].(string); ok && username != "" {
		payload["username"] = username
	}

	if iconEmoji, ok := channel.Config["iconEmoji"].(string); ok && iconEmoji != "" {
		payload["icon_emoji"] = iconEmoji
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
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *SlackSender) GetType() string {
	return model.AlertChannelTypeSlack
}

func (s *SlackSender) Validate(config map[string]interface{}) error {
	webhookURL, ok := config["webhookUrl"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhookUrl is required")
	}

	return nil
}
