package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mingrenya/AI-Waf/server/model"
)

// WebhookSender Webhook 告警发送器
type WebhookSender struct {
	client *http.Client
}

// NewWebhookSender 创建 Webhook 发送器
func NewWebhookSender() Sender {
	return &WebhookSender{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *WebhookSender) Send(ctx context.Context, channel *model.AlertChannel, message string) error {
	url, ok := channel.Config["url"].(string)
	if !ok {
		return fmt.Errorf("webhook url not configured")
	}

	method, ok := channel.Config["method"].(string)
	if !ok {
		method = "POST"
	}

	timeout, ok := channel.Config["timeout"].(float64)
	if ok {
		s.client.Timeout = time.Duration(timeout) * time.Second
	}

	// 构建请求体
	payload := map[string]interface{}{
		"message":   message,
		"timestamp": time.Now().Unix(),
		"channel":   channel.Name,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加自定义 headers
	if headers, ok := channel.Config["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *WebhookSender) GetType() string {
	return model.AlertChannelTypeWebhook
}

func (s *WebhookSender) Validate(config map[string]interface{}) error {
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("url is required")
	}

	if method, ok := config["method"].(string); ok {
		if method != "GET" && method != "POST" && method != "PUT" {
			return fmt.Errorf("method must be GET, POST, or PUT")
		}
	}

	return nil
}
