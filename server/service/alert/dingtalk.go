package alert

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mingrenya/AI-Waf/server/model"
)

// DingTalkSender 钉钉告警发送器
type DingTalkSender struct {
	client *http.Client
}

// NewDingTalkSender 创建钉钉发送器
func NewDingTalkSender() Sender {
	return &DingTalkSender{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *DingTalkSender) Send(ctx context.Context, channel *model.AlertChannel, message string) error {
	webhookURL, ok := channel.Config["webhookUrl"].(string)
	if !ok {
		return fmt.Errorf("dingtalk webhook URL not configured")
	}

	// 如果配置了 secret，需要计算签名
	secret, _ := channel.Config["secret"].(string)
	if secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := s.generateSign(timestamp, secret)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
	}

	// 构建钉钉消息格式
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": message,
		},
	}

	// 添加 @ 功能
	at := make(map[string]interface{})
	if isAtAll, ok := channel.Config["isAtAll"].(bool); ok {
		at["isAtAll"] = isAtAll
	}

	if atMobiles, ok := channel.Config["atMobiles"].([]interface{}); ok && len(atMobiles) > 0 {
		mobiles := make([]string, 0, len(atMobiles))
		for _, mobile := range atMobiles {
			if mobileStr, ok := mobile.(string); ok {
				mobiles = append(mobiles, mobileStr)
			}
		}
		if len(mobiles) > 0 {
			at["atMobiles"] = mobiles
		}
	}

	if len(at) > 0 {
		payload["at"] = at
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
		return fmt.Errorf("failed to send dingtalk message: %w", err)
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
		return fmt.Errorf("dingtalk returned error: %s", result.ErrMsg)
	}

	return nil
}

func (s *DingTalkSender) GetType() string {
	return model.AlertChannelTypeDingTalk
}

func (s *DingTalkSender) Validate(config map[string]interface{}) error {
	webhookURL, ok := config["webhookUrl"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhookUrl is required")
	}

	return nil
}

// generateSign 生成钉钉签名
func (s *DingTalkSender) generateSign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
