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

// WeComSender 企业微信告警发送器
type WeComSender struct {
	client *http.Client
}

// NewWeComSender 创建企业微信发送器
func NewWeComSender() Sender {
	return &WeComSender{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *WeComSender) Send(ctx context.Context, channel *model.AlertChannel, message string) error {
	webhookURL, ok := channel.Config["webhookUrl"].(string)
	if !ok {
		return fmt.Errorf("wecom webhook URL not configured")
	}

	// 构建企业微信消息格式
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": message,
		},
	}

	// 添加 @ 功能
	if mentionedList, ok := channel.Config["mentionedList"].([]interface{}); ok && len(mentionedList) > 0 {
		mentioned := make([]string, 0, len(mentionedList))
		for _, item := range mentionedList {
			if str, ok := item.(string); ok {
				mentioned = append(mentioned, str)
			}
		}
		if len(mentioned) > 0 {
			payload["text"].(map[string]interface{})["mentioned_list"] = mentioned
		}
	}

	if mentionedMobileList, ok := channel.Config["mentionedMobileList"].([]interface{}); ok && len(mentionedMobileList) > 0 {
		mobileList := make([]string, 0, len(mentionedMobileList))
		for _, item := range mentionedMobileList {
			if str, ok := item.(string); ok {
				mobileList = append(mobileList, str)
			}
		}
		if len(mobileList) > 0 {
			payload["text"].(map[string]interface{})["mentioned_mobile_list"] = mobileList
		}
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
		return fmt.Errorf("failed to send wecom message: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("wecom returned error: %s", result.ErrMsg)
	}

	return nil
}

func (s *WeComSender) GetType() string {
	return model.AlertChannelTypeWeCom
}

func (s *WeComSender) Validate(config map[string]interface{}) error {
	webhookURL, ok := config["webhookUrl"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhookUrl is required")
	}

	return nil
}
