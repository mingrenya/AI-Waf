// tools/client.go
// AI-Waf后端API客户端
package tools

import (
	"bytes"
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
}

// NewAPIClient 创建新的API客户端
func NewAPIClient(baseURL, token string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get 发送GET请求
func (c *APIClient) Get(path string) ([]byte, error) {
	url := c.BaseURL + path
	log.Printf("[API请求] GET %s", url)
	start := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token[:20]+"...")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[API错误] 读取响应失败: %v", err)
		return nil, err
	}

	duration := time.Since(start)
	log.Printf("[API响应] GET %s - 状态码: %d - 耗时: %v - 响应大小: %d bytes", 
		path, resp.StatusCode, duration, len(body))

	if resp.StatusCode >= 400 {
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API错误 %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Post 发送POST请求
func (c *APIClient) Post(path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[API错误] JSON序列化失败: %v", err)
		return nil, err
	}

	url := c.BaseURL + path
	log.Printf("[API请求] POST %s - 数据: %s", url, string(jsonData))
	start := time.Now()
	elapsed := time.Since(start)
	log.Printf("[API请求] 响应耗时: %v", elapsed)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token[:20]+"...")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API错误 %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Patch 发送PATCH请求
func (c *APIClient) Patch(path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[API错误] JSON序列化失败: %v", err)
		return nil, err
	}

	url := c.BaseURL + path
	log.Printf("[API请求] PATCH %s - 数据: %s", url, string(jsonData))
	start := time.Now()

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token[:20]+"...")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[API错误] 读取响应失败: %v", err)
		return nil, err
	}

	duration := time.Since(start)
	log.Printf("[API响应] PATCH %s - 状态码: %d - 耗时: %v - 响应大小: %d bytes", 
		path, resp.StatusCode, duration, len(body))

	if resp.StatusCode >= 400 {
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API错误 %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Put 发送PUT请求
func (c *APIClient) Put(path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[API错误] JSON序列化失败: %v", err)
		return nil, err
	}

	url := c.BaseURL + path
	log.Printf("[API请求] PUT %s - 数据: %s", url, string(jsonData))
	start := time.Now()

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token[:20]+"...")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[API错误] 读取响应失败: %v", err)
		return nil, err
	}

	duration := time.Since(start)
	log.Printf("[API响应] PUT %s - 状态码: %d - 耗时: %v - 响应大小: %d bytes", 
		path, resp.StatusCode, duration, len(body))

	if resp.StatusCode >= 400 {
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API错误 %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Delete 发送DELETE请求
func (c *APIClient) Delete(path string) error {
	url := c.BaseURL + path
	log.Printf("[API请求] DELETE %s", url)
	start := time.Now()

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("[API错误] 创建请求失败: %v", err)
		return err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token[:20]+"...")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[API错误] 请求失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	log.Printf("[API响应] DELETE %s - 状态码: %d - 耗时: %v", path, resp.StatusCode, duration)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[API错误] %d - %s", resp.StatusCode, string(body))
		return fmt.Errorf("API错误 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
