// tools/client.go
// AI-Waf后端API客户端
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// APIClient WAF后端API客户端
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
	// 用于 Token 过期时自动重新登录
	username string
	password string
}

// AutoLogin 使用用户名密码自动登录，获取长期Token（90天）
// 调用 /api/v1/auth/login-service 接口
func AutoLogin(baseURL, username, password string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	payload, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := client.Post(baseURL+"/api/v1/auth/login-service", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("自动登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取登录响应失败: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("自动登录失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %w", err)
	}
	if result.Data.Token == "" {
		return "", fmt.Errorf("登录响应中未包含Token")
	}
	return result.Data.Token, nil
}

// NewAPIClient 创建新的API客户端
func NewAPIClient(baseURL, token string) *APIClient {
	return newAPIClient(baseURL, token, "", "")
}

// NewAPIClientWithCredentials 创建支持自动 Token 刷新的API客户端
func NewAPIClientWithCredentials(baseURL, token, username, password string) *APIClient {
	return newAPIClient(baseURL, token, username, password)
}

func newAPIClient(baseURL, token, username, password string) *APIClient {
	// 配置HTTP Transport以优化连接池性能
	transport := &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 10,               // 每个host的最大空闲连接
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时
		DisableKeepAlives:   false,            // 启用Keep-Alive
	}

	return &APIClient{
		BaseURL:  baseURL,
		Token:    token,
		username: username,
		password: password,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// refreshToken 尝试用用户名密码重新获取 Token（仅当配置了凭据时）
func (c *APIClient) refreshToken() error {
	if c.username == "" || c.password == "" {
		return fmt.Errorf("Token 已过期且未配置自动登录凭据（WAF_USERNAME/WAF_PASSWORD）")
	}
	log.Println("[认证] Token 已过期，正在自动重新登录...")
	token, err := AutoLogin(c.BaseURL, c.username, c.password)
	if err != nil {
		return fmt.Errorf("自动重新登录失败: %w", err)
	}
	c.Token = token
	log.Printf("[认证] Token 刷新成功 (长度: %d 字符)", len(token))
	return nil
}

// Get 发送GET请求
func (c *APIClient) Get(path string) ([]byte, error) {
	return c.GetWithContext(context.Background(), path)
}

// GetWithContext 发送带context的GET请求（支持超时控制）
func (c *APIClient) GetWithContext(ctx context.Context, path string) ([]byte, error) {
	url := c.BaseURL + path
	log.Printf("[API请求] GET %s", url)
	start := time.Now()

	// 创建带超时的context（单个请求最多10秒）
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		// 判断是否为网络或超时错误
		return nil, NewNetworkError("GET "+path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[API错误] 读取响应失败: %v", err)
		return nil, NewNetworkError("读取响应", err)
	}

	duration := time.Since(start)
	log.Printf("[API响应] GET %s - 状态码: %d - 耗时: %v - 响应大小: %d bytes",
		path, resp.StatusCode, duration, len(body))

	if resp.StatusCode == 401 {
		if err := c.refreshToken(); err != nil {
			log.Printf("[认证] Token 刷新失败: %v", err)
			return nil, FormatAPIError("GET "+path, resp.StatusCode, body, nil)
		}
		// 用新 Token 重试一次
		return c.GetWithContext(ctx, path)
	}

	if resp.StatusCode >= 400 {
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return nil, FormatAPIError("GET "+path, resp.StatusCode, body, nil)
	}

	return body, nil
}

// Post 发送POST请求
func (c *APIClient) Post(path string, data interface{}) ([]byte, error) {
	return c.PostWithContext(context.Background(), path, data)
}

// PostWithContext 发送带context的POST请求（支持超时控制）
func (c *APIClient) PostWithContext(ctx context.Context, path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[API错误] JSON序列化失败: %v", err)
		return nil, err
	}

	url := c.BaseURL + path
	log.Printf("[API请求] POST %s - 数据: %s", url, string(jsonData))
	start := time.Now()

	// 创建带超时的context（单个请求最多10秒）
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		return nil, NewNetworkError("POST "+path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[API错误] 读取响应失败: %v", err)
		return nil, NewNetworkError("读取响应", err)
	}

	duration := time.Since(start)
	log.Printf("[API响应] POST %s - 状态码: %d - 耗时: %v - 响应大小: %d bytes",
		path, resp.StatusCode, duration, len(body))

	if resp.StatusCode == 401 {
		if err := c.refreshToken(); err != nil {
			log.Printf("[认证] Token 刷新失败: %v", err)
			return nil, FormatAPIError("POST "+path, resp.StatusCode, body, nil)
		}
		return c.PostWithContext(ctx, path, data)
	}

	if resp.StatusCode >= 400 {
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return nil, FormatAPIError("POST "+path, resp.StatusCode, body, nil)
	}

	return body, nil
}

// Patch 发送PATCH请求
func (c *APIClient) Patch(path string, data interface{}) ([]byte, error) {
	return c.PatchWithContext(context.Background(), path, data)
}

// PatchWithContext 发送带context的PATCH请求（支持超时控制）
func (c *APIClient) PatchWithContext(ctx context.Context, path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[API错误] JSON序列化失败: %v", err)
		return nil, err
	}

	url := c.BaseURL + path
	log.Printf("[API请求] PATCH %s - 数据: %s", url, string(jsonData))
	start := time.Now()

	// 创建带超时的context（单个请求最多10秒）
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		return nil, NewNetworkError("PATCH "+path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[API错误] 读取响应失败: %v", err)
		return nil, NewNetworkError("读取响应", err)
	}

	duration := time.Since(start)
	log.Printf("[API响应] PATCH %s - 状态码: %d - 耗时: %v - 响应大小: %d bytes",
		path, resp.StatusCode, duration, len(body))

	if resp.StatusCode == 401 {
		if err := c.refreshToken(); err != nil {
			log.Printf("[认证] Token 刷新失败: %v", err)
			return nil, FormatAPIError("PATCH "+path, resp.StatusCode, body, nil)
		}
		return c.PatchWithContext(ctx, path, data)
	}

	if resp.StatusCode >= 400 {
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return nil, FormatAPIError("PATCH "+path, resp.StatusCode, body, nil)
	}

	return body, nil
}

// Put 发送PUT请求
func (c *APIClient) Put(path string, data interface{}) ([]byte, error) {
	return c.PutWithContext(context.Background(), path, data)
}

// PutWithContext 发送带context的PUT请求（支持超时控制）
func (c *APIClient) PutWithContext(ctx context.Context, path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[API错误] JSON序列化失败: %v", err)
		return nil, err
	}

	url := c.BaseURL + path
	log.Printf("[API请求] PUT %s - 数据: %s", url, string(jsonData))
	start := time.Now()

	// 创建带超时的context（单个请求最多10秒）
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		return nil, NewNetworkError("PUT "+path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[API错误] 读取响应失败: %v", err)
		return nil, NewNetworkError("读取响应", err)
	}

	duration := time.Since(start)
	log.Printf("[API响应] PUT %s - 状态码: %d - 耗时: %v - 响应大小: %d bytes",
		path, resp.StatusCode, duration, len(body))

	if resp.StatusCode == 401 {
		if err := c.refreshToken(); err != nil {
			log.Printf("[认证] Token 刷新失败: %v", err)
			return nil, FormatAPIError("PUT "+path, resp.StatusCode, body, nil)
		}
		return c.PutWithContext(ctx, path, data)
	}

	if resp.StatusCode >= 400 {
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return nil, FormatAPIError("PUT "+path, resp.StatusCode, body, nil)
	}

	return body, nil
}

// Delete 发送DELETE请求
func (c *APIClient) Delete(path string) error {
	return c.DeleteWithContext(context.Background(), path)
}

// DeleteWithContext 发送带context的DELETE请求（支持超时控制）
func (c *APIClient) DeleteWithContext(ctx context.Context, path string) error {
	url := c.BaseURL + path
	log.Printf("[API请求] DELETE %s", url)
	start := time.Now()

	// 创建带超时的context（单个请求最多10秒）
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "DELETE", url, nil)
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		return NewNetworkError("DELETE "+path, err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	log.Printf("[API响应] DELETE %s - 状态码: %d - 耗时: %v", path, resp.StatusCode, duration)

	if resp.StatusCode == 401 {
		body, _ := io.ReadAll(resp.Body)
		if err := c.refreshToken(); err != nil {
			log.Printf("[认证] Token 刷新失败: %v", err)
			return FormatAPIError("DELETE "+path, resp.StatusCode, body, nil)
		}
		return c.DeleteWithContext(ctx, path)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return FormatAPIError("DELETE "+path, resp.StatusCode, body, nil)
	}

	return nil
}
