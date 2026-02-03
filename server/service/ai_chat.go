package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/rs/zerolog"
)

// AIChatService AI聊天服务接口
type AIChatService interface {
	Chat(ctx context.Context, req *dto.AIChatRequest) (*dto.AIChatResponse, error)
	ChatStream(ctx context.Context, req *dto.AIChatRequest, writer io.Writer) error
}

type aiChatServiceImpl struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	logger     zerolog.Logger
}

// DeepSeek API请求/响应结构
type deepseekChatRequest struct {
	Model    string                `json:"model"`
	Messages []deepseekChatMessage `json:"messages"`
	Stream   bool                  `json:"stream"`
}

type deepseekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                 `json:"index"`
		Message      deepseekChatMessage `json:"message"`
		FinishReason string              `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type deepseekStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
			Role    string `json:"role,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// NewAIChatService 创建AI聊天服务
func NewAIChatService() AIChatService {
	logger := config.GetServiceLogger("ai-chat")

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		logger.Warn().Msg("DEEPSEEK_API_KEY not set, AI chat service will not work")
	}

	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	return &aiChatServiceImpl{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

// Chat 发送聊天消息（非流式）
func (s *aiChatServiceImpl) Chat(ctx context.Context, req *dto.AIChatRequest) (*dto.AIChatResponse, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY not configured")
	}

	// 构建消息列表
	messages := make([]deepseekChatMessage, 0, len(req.Messages)+2)

	// 添加系统提示
	messages = append(messages, deepseekChatMessage{
		Role:    "system",
		Content: "你是一个AI安全助手，专门帮助用户分析WAF日志、生成防护规则、评估安全威胁。请用简洁专业的语言回答问题。",
	})

	// 添加历史消息
	for _, msg := range req.Messages {
		messages = append(messages, deepseekChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 添加当前消息
	messages = append(messages, deepseekChatMessage{
		Role:    "user",
		Content: req.Message,
	})

	// 构建请求
	deepseekReq := deepseekChatRequest{
		Model:    s.model,
		Messages: messages,
		Stream:   false,
	}

	reqBody, err := json.Marshal(deepseekReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	s.logger.Debug().
		Str("model", s.model).
		Int("message_count", len(messages)).
		Msg("Sending chat request to DeepSeek API")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error().
			Int("status_code", resp.StatusCode).
			Str("response", string(body)).
			Msg("DeepSeek API returned error")
		return nil, fmt.Errorf("DeepSeek API error: %s", string(body))
	}

	// 解析响应
	var deepseekResp deepseekChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from DeepSeek API")
	}

	message := deepseekResp.Choices[0].Message.Content

	s.logger.Info().
		Int("prompt_tokens", deepseekResp.Usage.PromptTokens).
		Int("completion_tokens", deepseekResp.Usage.CompletionTokens).
		Int("total_tokens", deepseekResp.Usage.TotalTokens).
		Msg("Successfully received chat response")

	return &dto.AIChatResponse{
		Message:   message,
		Timestamp: time.Now(),
	}, nil
}

// ChatStream 发送聊天消息（流式）
func (s *aiChatServiceImpl) ChatStream(ctx context.Context, req *dto.AIChatRequest, writer io.Writer) error {
	if s.apiKey == "" {
		return fmt.Errorf("DEEPSEEK_API_KEY not configured")
	}

	// 构建消息列表
	messages := make([]deepseekChatMessage, 0, len(req.Messages)+2)

	// 添加系统提示
	messages = append(messages, deepseekChatMessage{
		Role:    "system",
		Content: "你是一个AI安全助手，专门帮助用户分析WAF日志、生成防护规则、评估安全威胁。请用简洁专业的语言回答问题。",
	})

	// 添加历史消息
	for _, msg := range req.Messages {
		messages = append(messages, deepseekChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 添加当前消息
	messages = append(messages, deepseekChatMessage{
		Role:    "user",
		Content: req.Message,
	})

	// 构建请求
	deepseekReq := deepseekChatRequest{
		Model:    s.model,
		Messages: messages,
		Stream:   true,
	}

	reqBody, err := json.Marshal(deepseekReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DeepSeek API error: %s", string(body))
	}

	// 读取流式响应
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行和注释
		if line == "" || line == "data: [DONE]" {
			continue
		}

		// 移除 "data: " 前缀
		if len(line) > 6 && line[:6] == "data: " {
			line = line[6:]
		}

		// 解析JSON
		var streamResp deepseekStreamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			s.logger.Warn().Err(err).Str("line", line).Msg("Failed to parse stream response")
			continue
		}

		// 写入内容
		if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
			content := streamResp.Choices[0].Delta.Content
			responseData := dto.AIChatStreamResponse{
				Content: content,
				Done:    streamResp.Choices[0].FinishReason != "",
			}

			jsonData, err := json.Marshal(responseData)
			if err != nil {
				continue
			}

			// 写入SSE格式
			fmt.Fprintf(writer, "data: %s\n\n", jsonData)
			if flusher, ok := writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	// 发送完成标记
	doneData := dto.AIChatStreamResponse{
		Content: "",
		Done:    true,
	}
	jsonData, _ := json.Marshal(doneData)
	fmt.Fprintf(writer, "data: %s\n\n", jsonData)
	if flusher, ok := writer.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}
