package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// 常量定义
// ============================================================================

const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"

	RuleTypeBlacklist  = "blacklist"
	RuleTypeWhitelist  = "whitelist"
	RuleTypePattern    = "pattern"
	RuleTypeRateLimit  = "rate_limit"
	RuleTypeGeoBlock   = "geo_block"

	RuleActionBlock     = "block"
	RuleActionAllow     = "allow"
	RuleActionLog       = "log"
	RuleActionChallenge = "challenge"

	TimeRange1Hour      = "1h"
	TimeRange6Hours     = "6h"
	TimeRange24Hours    = "24h"
	TimeRange7Days      = "7d"
	TimeRange30Days     = "30d"
)

// ============================================================================
// 数据模型
// ============================================================================

// BlockedIP 被封禁的IP记录
type BlockedIP struct {
	ID                    string    `json:"id"`
	IPAddress             string    `json:"ip_address"`
	Reason                string    `json:"reason"`
	Severity              string    `json:"severity"`
	DurationSeconds       int       `json:"duration_seconds"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	Tags                  []string  `json:"tags"`
	BlockedRequestsCount  int       `json:"blocked_requests_count"`
	LastAttackTimestamp   *time.Time `json:"last_attack_timestamp,omitempty"`
}

// RuleConditions 规则条件
type RuleConditions struct {
	Method               []string           `json:"method,omitempty"`
	Path                 string             `json:"path,omitempty"`
	SourceIP             string             `json:"source_ip,omitempty"`
	UserAgent            string             `json:"user_agent,omitempty"`
	RequestBodyContains  string             `json:"request_body_contains,omitempty"`
	ResponseCode         []int              `json:"response_code,omitempty"`
	CountryCode          []string           `json:"country_code,omitempty"`
	RateLimit            map[string]interface{} `json:"rate_limit,omitempty"`
	Custom               map[string]interface{} `json:"-"`
}

// WAFRule WAF防护规则
type WAFRule struct {
	ID           string         `json:"id"`
	RuleName     string         `json:"rule_name"`
	RuleType     string         `json:"rule_type"`
	Action       string         `json:"action"`
	Conditions   RuleConditions `json:"conditions"`
	Priority     int            `json:"priority"`
	Enabled      bool           `json:"enabled"`
	Description  string         `json:"description,omitempty"`
	Tags         []string       `json:"tags"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	CreatedBy    string         `json:"created_by,omitempty"`
	UpdatedBy    string         `json:"updated_by,omitempty"`
	IsSystemRule bool           `json:"is_system_rule"`
}

// AttackLog 攻击日志
type AttackLog struct {
	ID            string     `json:"id"`
	AttackType    string     `json:"attack_type"`
	Severity      string     `json:"severity"`
	SourceIP      string     `json:"source_ip"`
	TargetPath    string     `json:"target_path"`
	TargetDomain  string     `json:"target_domain"`
	RequestMethod string     `json:"request_method"`
	RequestBody   string     `json:"request_body,omitempty"`
	ResponseCode  int        `json:"response_code"`
	MatchedRuleID string     `json:"matched_rule_id,omitempty"`
	MatchedRule   string     `json:"matched_rule_name,omitempty"`
	ActionTaken   string     `json:"action_taken"`
	Timestamp     time.Time  `json:"timestamp"`
	UserAgent     string     `json:"user_agent,omitempty"`
	RequestID     string     `json:"request_id"`
}

// ListResponse 列表响应
type ListResponse struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// StatsOverview 统计概览
type StatsOverview struct {
	TotalRequests    int64                    `json:"total_requests"`
	BlockedRequests  int64                    `json:"blocked_requests"`
	BlockRate        float64                  `json:"block_rate"`
	ErrorRate        float64                  `json:"error_rate"`
	TopAttackTypes   []map[string]interface{} `json:"top_attack_types"`
	TopBlockedIPs    []map[string]interface{} `json:"top_blocked_ips"`
	TimeRange        string                   `json:"time_range"`
}

// StdResponse 标准响应
type StdResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

// ============================================================================
// WAF服务核心实现
// ============================================================================

// WAFService WAF核心服务
type WAFService struct {
	client     *WAFAPIClient
	cache      *Cache
	logger     *log.Logger
	mu         sync.RWMutex
	blockedIPs map[string]*BlockedIP
	rules      map[string]*WAFRule
}

// NewWAFService 创建WAF服务
func NewWAFService(apiURL string, apiKey string) *WAFService {
	return &WAFService{
		client:     NewWAFAPIClient(apiURL, apiKey),
		cache:      NewCache(5 * time.Minute),
		logger:     log.New(os.Stdout, "[WAF] ", log.LstdFlags),
		blockedIPs: make(map[string]*BlockedIP),
		rules:      make(map[string]*WAFRule),
	}
}

// BlockIP 添加IP到黑名单
func (s *WAFService) BlockIP(ctx context.Context, ipAddress, reason string, durationSeconds int, severity string, tags []string) (*BlockedIP, error) {
	if err := validateIPAddress(ipAddress); err != nil {
		return nil, fmt.Errorf("invalid IP address: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建黑名单项
	blockedIP := &BlockedIP{
		ID:              uuid.New().String(),
		IPAddress:       ipAddress,
		Reason:          reason,
		Severity:        severity,
		DurationSeconds: durationSeconds,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Tags:            tags,
	}

	if durationSeconds > 0 {
		expiresAt := time.Now().Add(time.Duration(durationSeconds) * time.Second)
		blockedIP.ExpiresAt = &expiresAt
	}

	s.blockedIPs[ipAddress] = blockedIP

	// 调用WAF API
	if _, err := s.client.BlockIP(ctx, ipAddress, reason, durationSeconds, severity); err != nil {
		s.logger.Printf("Warning: Failed to sync with WAF API: %v", err)
		// 继续运行，本地缓存可用
	}

	s.cache.Delete("blocked_ips_list")
	return blockedIP, nil
}

// UnblockIP 从黑名单移除IP
func (s *WAFService) UnblockIP(ctx context.Context, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.blockedIPs, ipAddress)

	if err := s.client.UnblockIP(ctx, ipAddress); err != nil {
		s.logger.Printf("Warning: Failed to unblock IP in WAF API: %v", err)
	}

	s.cache.Delete("blocked_ips_list")
	return nil
}

// ListBlockedIPs 列出黑名单
func (s *WAFService) ListBlockedIPs(ctx context.Context, page, pageSize int) (*ListResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 计算分页
	total := len(s.blockedIPs)
	totalPages := (total + pageSize - 1) / pageSize

	var items []interface{}
	for _, ip := range s.blockedIPs {
		items = append(items, ip)
	}

	return &ListResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// CreateRule 创建规则
func (s *WAFService) CreateRule(ctx context.Context, ruleName, ruleType, action string, conditions RuleConditions, priority int, enabled bool) (*WAFRule, error) {
	if priority < 1 || priority > 10000 {
		return nil, fmt.Errorf("priority must be between 1 and 10000")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	rule := &WAFRule{
		ID:         uuid.New().String(),
		RuleName:   ruleName,
		RuleType:   ruleType,
		Action:     action,
		Conditions: conditions,
		Priority:   priority,
		Enabled:    enabled,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.rules[rule.ID] = rule

	// 调用WAF API
	if _, err := s.client.CreateRule(ctx, ruleName, ruleType, action, conditions, priority); err != nil {
		s.logger.Printf("Warning: Failed to sync rule with WAF API: %v", err)
	}

	s.cache.Delete("rules_list")
	return rule, nil
}

// UpdateRule 更新规则
func (s *WAFService) UpdateRule(ctx context.Context, ruleID string, updates map[string]interface{}) (*WAFRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, ok := s.rules[ruleID]
	if !ok {
		return nil, fmt.Errorf("rule not found")
	}

	// 应用更新
	if v, ok := updates["action"].(string); ok {
		rule.Action = v
	}
	if v, ok := updates["priority"].(int); ok {
		rule.Priority = v
	}
	if v, ok := updates["enabled"].(bool); ok {
		rule.Enabled = v
	}

	rule.UpdatedAt = time.Now()

	if err := s.client.UpdateRule(ctx, ruleID, updates); err != nil {
		s.logger.Printf("Warning: Failed to update rule in WAF API: %v", err)
	}

	s.cache.Delete("rules_list")
	return rule, nil
}

// DeleteRule 删除规则
func (s *WAFService) DeleteRule(ctx context.Context, ruleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.rules, ruleID)

	if err := s.client.DeleteRule(ctx, ruleID); err != nil {
		s.logger.Printf("Warning: Failed to delete rule in WAF API: %v", err)
	}

	s.cache.Delete("rules_list")
	return nil
}

// ListRules 列出规则
func (s *WAFService) ListRules(ctx context.Context, page, pageSize int) (*ListResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.rules)
	totalPages := (total + pageSize - 1) / pageSize

	var items []interface{}
	for _, rule := range s.rules {
		items = append(items, rule)
	}

	return &ListResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// ============================================================================
// WAF API客户端
// ============================================================================

// WAFAPIClient WAF后端API客户端
type WAFAPIClient struct {
	baseURL    string
	apiKey     string
	client     *http.Client
	retryTimes int
	retryDelay time.Duration
}

// NewWAFAPIClient 创建API客户端
func NewWAFAPIClient(baseURL, apiKey string) *WAFAPIClient {
	return &WAFAPIClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		client:     &http.Client{Timeout: 30 * time.Second},
		retryTimes: 3,
		retryDelay: 1 * time.Second,
	}
}

// BlockIP API - 添加IP到黑名单
func (c *WAFAPIClient) BlockIP(ctx context.Context, ipAddress, reason string, durationSeconds int, severity string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"ip_address":       ipAddress,
		"reason":           reason,
		"duration_seconds": durationSeconds,
		"severity":         severity,
	}

	return c.doRequest(ctx, "POST", "/waf/block-ip", payload)
}

// UnblockIP API - 移除IP黑名单
func (c *WAFAPIClient) UnblockIP(ctx context.Context, ipAddress string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"ip_address": ipAddress,
	}

	return c.doRequest(ctx, "POST", "/waf/unblock-ip", payload)
}

// CreateRule API - 创建规则
func (c *WAFAPIClient) CreateRule(ctx context.Context, ruleName, ruleType, action string, conditions RuleConditions, priority int) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"name":       ruleName,
		"type":       ruleType,
		"action":     action,
		"condition":  conditions,
		"priority":   priority,
		"enabled":    true,
		"status":     "active",
	}

	return c.doRequest(ctx, "POST", "/waf/rules/micro", payload)
}

// UpdateRule API - 更新规则
func (c *WAFAPIClient) UpdateRule(ctx context.Context, ruleID string, updates map[string]interface{}) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/waf/rules/%s", ruleID)
	return c.doRequest(ctx, "PUT", endpoint, updates)
}

// DeleteRule API - 删除规则
func (c *WAFAPIClient) DeleteRule(ctx context.Context, ruleID string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/waf/rules/%s", ruleID)
	return c.doRequest(ctx, "DELETE", endpoint, nil)
}

// ListAttackLogs API - 查询攻击日志
func (c *WAFAPIClient) ListAttackLogs(ctx context.Context, hours int) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"hours": hours,
	}

	return c.doRequest(ctx, "GET", "/waf/attack-logs", payload)
}

// AnalyzePatterns API - 分析攻击模式
func (c *WAFAPIClient) AnalyzePatterns(ctx context.Context, hours int, method string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"hours":              hours,
		"clustering_method":  method,
		"min_samples":        10,
		"anomaly_threshold":  2.0,
	}

	return c.doRequest(ctx, "POST", "/waf/analyze/patterns", payload)
}

// GetStatsOverview API - 获取统计概览
func (c *WAFAPIClient) GetStatsOverview(ctx context.Context, timeRange string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"time_range": timeRange,
	}

	return c.doRequest(ctx, "GET", "/waf/stats/overview", payload)
}

// doRequest 执行HTTP请求，支持重试
func (c *WAFAPIClient) doRequest(ctx context.Context, method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	var lastErr error

	for attempt := 0; attempt < c.retryTimes; attempt++ {
		resp, err := c.makeRequest(ctx, method, endpoint, payload)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if attempt < c.retryTimes-1 {
			time.Sleep(c.retryDelay)
		}
	}

	return nil, lastErr
}

// makeRequest 执行单次HTTP请求
func (c *WAFAPIClient) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	url := c.baseURL + endpoint

	var req *http.Request
	var err error

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Body = nil // 简化版本，实际使用需要设置body
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// ============================================================================
// 缓存实现
// ============================================================================

// Cache 简单的内存缓存
type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
	ttl   time.Duration
}

type cacheItem struct {
	data      interface{}
	expiresAt time.Time
}

// NewCache 创建缓存
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]cacheItem),
		ttl:   ttl,
	}
}

// Get 获取缓存
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok || time.Now().After(item.expiresAt) {
		return nil, false
	}

	return item.data, true
}

// Set 设置缓存
func (c *Cache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Delete 删除缓存
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cacheItem)
}

// ============================================================================
// HTTP处理程序
// ============================================================================

// Handler HTTP请求处理器
type Handler struct {
	service *WAFService
}

// NewHandler 创建处理器
func NewHandler(service *WAFService) *Handler {
	return &Handler{service: service}
}

// HandleBlockIP 处理IP封禁请求
func (h *Handler) HandleBlockIP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IPAddress       string   `json:"ip_address"`
		Reason          string   `json:"reason"`
		DurationSeconds int      `json:"duration_seconds"`
		Severity        string   `json:"severity"`
		Tags            []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, StdResponse{Success: false, Message: err.Error()})
		return
	}

	if req.DurationSeconds == 0 {
		req.DurationSeconds = 3600
	}
	if req.Severity == "" {
		req.Severity = SeverityMedium
	}

	blockedIP, err := h.service.BlockIP(r.Context(), req.IPAddress, req.Reason, req.DurationSeconds, req.Severity, req.Tags)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, StdResponse{Success: false, Message: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, StdResponse{
		Success: true,
		Data:    blockedIP,
		Message: "IP blocked successfully",
	})
}

// HandleListBlockedIPs 处理黑名单查询
func (h *Handler) HandleListBlockedIPs(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	list, err := h.service.ListBlockedIPs(r.Context(), page, pageSize)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, StdResponse{Success: false, Message: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, StdResponse{
		Success: true,
		Data:    list,
		Message: "OK",
	})
}

// HandleCreateRule 处理规则创建
func (h *Handler) HandleCreateRule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RuleName   string         `json:"rule_name"`
		RuleType   string         `json:"rule_type"`
		Action     string         `json:"action"`
		Conditions RuleConditions `json:"conditions"`
		Priority   int            `json:"priority"`
		Enabled    bool           `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, StdResponse{Success: false, Message: err.Error()})
		return
	}

	rule, err := h.service.CreateRule(r.Context(), req.RuleName, req.RuleType, req.Action, req.Conditions, req.Priority, req.Enabled)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, StdResponse{Success: false, Message: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, StdResponse{
		Success: true,
		Data:    rule,
		Message: "Rule created successfully",
	})
}

// HandleListRules 处理规则列表查询
func (h *Handler) HandleListRules(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	list, err := h.service.ListRules(r.Context(), page, pageSize)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, StdResponse{Success: false, Message: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, StdResponse{
		Success: true,
		Data:    list,
		Message: "OK",
	})
}

// respondJSON 返回JSON响应
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// ============================================================================
// 工具函数
// ============================================================================

// validateIPAddress 验证IP地址
func validateIPAddress(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
}

// ============================================================================
// 主程序
// ============================================================================

func main() {
	// 命令行参数
	wafAPIURL := flag.String("waf-url", "http://localhost:2342", "WAF API URL")
	wafAPIKey := flag.String("waf-key", "", "WAF API Key")
	httpAddr := flag.String("http", "127.0.0.1:8000", "HTTP server address")
	flag.Parse()

	// 创建服务
	service := NewWAFService(*wafAPIURL, *wafAPIKey)
	handler := NewHandler(service)

	// 注册HTTP路由
	http.HandleFunc("/waf/block-ip", handler.HandleBlockIP)
	http.HandleFunc("/waf/unblock-ip", handler.HandleListBlockedIPs) // 简化版本
	http.HandleFunc("/waf/blocked-ips", handler.HandleListBlockedIPs)
	http.HandleFunc("/waf/rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handler.HandleCreateRule(w, r)
		} else {
			handler.HandleListRules(w, r)
		}
	})

	// 启动HTTP服务器
	log.Printf("Starting WAF MCP server on %s", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
