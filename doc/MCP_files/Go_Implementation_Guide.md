# Go语言 WAF MCP 服务器 - 完整实现指南

## 第一部分：项目结构与初始化

### 1.1 项目目录结构

```
waf-mcp-go/
├── cmd/
│   ├── server/
│   │   └── main.go              # 主程序入口
│   └── cli/
│       └── main.go              # CLI工具
├── internal/
│   ├── models/
│   │   ├── blocked_ip.go
│   │   ├── rule.go
│   │   ├── attack_log.go
│   │   └── response.go
│   ├── service/
│   │   ├── ip_blacklist.go      # IP黑名单服务
│   │   ├── rule_manager.go      # 规则管理
│   │   ├── attack_analyzer.go   # 攻击分析
│   │   ├── monitoring.go        # 监控统计
│   │   └── extended_tools.go    # 扩展工具
│   ├── api/
│   │   ├── client.go            # WAF API客户端
│   │   ├── handlers.go          # HTTP处理程序
│   │   └── middleware.go        # 中间件
│   ├── storage/
│   │   ├── cache.go             # 缓存实现
│   │   ├── database.go          # 数据库访问
│   │   └── repository.go        # 仓储模式
│   ├── ml/
│   │   ├── prediction.go        # 威胁预测
│   │   ├── anomaly.go           # 异常检测
│   │   └── suggestion.go        # 规则建议
│   └── config/
│       └── config.go            # 配置管理
├── pkg/
│   ├── logger/
│   │   └── logger.go            # 日志工具
│   ├── validator/
│   │   └── validator.go         # 数据验证
│   └── utils/
│       └── utils.go             # 工具函数
├── tests/
│   ├── unit/                    # 单元测试
│   ├── integration/             # 集成测试
│   └── fixtures/                # 测试数据
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── k8s/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── configmap.yaml
│   └── scripts/
│       ├── install.sh
│       └── health_check.sh
├── docs/
│   ├── API.md
│   ├── ARCHITECTURE.md
│   └── DEVELOPMENT.md
├── go.mod
├── go.sum
├── Makefile
├── .env.example
└── README.md
```

### 1.2 go.mod 配置

```go
module github.com/yourorg/waf-mcp-go

go 1.21

require (
    // Web框架
    github.com/gorilla/mux v1.8.1
    
    // HTTP客户端
    github.com/go-resty/resty/v2 v2.11.0
    
    // 日志
    go.uber.org/zap v1.26.0
    
    // 缓存
    github.com/patrickmn/go-cache v2.1.0+incompatible
    
    // 数据库
    github.com/jmoiron/sqlx v1.3.5
    github.com/lib/pq v1.10.9
    
    // 数据验证
    github.com/go-playground/validator/v10 v10.17.0
    
    // 环境变量
    github.com/joho/godotenv v1.5.1
    
    // UUID
    github.com/google/uuid v1.5.0
    
    // 时间处理
    github.com/golang-module/carbon/v2 v2.2.12
    
    // JSON处理
    encoding/json (标准库)
)
```

### 1.3 Makefile

```makefile
.PHONY: build run test clean install lint fmt docker

# 构建
build:
	go build -o bin/waf-mcp-server cmd/server/main.go

# 运行
run:
	go run cmd/server/main.go

# 测试
test:
	go test -v -race ./...

# 覆盖率测试
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# 代码格式化
fmt:
	go fmt ./...
	goimports -w .

# 代码检查
lint:
	golangci-lint run ./...

# 清理
clean:
	rm -rf bin/
	go clean

# 依赖安装
install:
	go mod download
	go mod tidy

# Docker构建
docker:
	docker build -t waf-mcp:latest .
	docker-compose up -d

# 帮助
help:
	@echo "Available targets:"
	@echo "  make build           - Build the application"
	@echo "  make run             - Run the application"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Lint code"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make install         - Install dependencies"
	@echo "  make docker          - Build and run Docker container"
```

---

## 第二部分：核心服务实现

### 2.1 配置管理 (config/config.go)

```go
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// 服务器配置
	Server ServerConfig
	// WAF API配置
	WAF WAFConfig
	// 数据库配置
	Database DatabaseConfig
	// 日志配置
	Logger LoggerConfig
	// 缓存配置
	Cache CacheConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type WAFConfig struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	RetryTimes int
	RetryDelay time.Duration
}

type DatabaseConfig struct {
	Driver      string
	DSN         string
	MaxOpenConn int
	MaxIdleConn int
}

type LoggerConfig struct {
	Level string // debug, info, warn, error
	File  string // 日志文件路径
}

type CacheConfig struct {
	TTL       time.Duration
	MaxSize   int
	CleanupInterval time.Duration
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 加载.env文件
	godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "127.0.0.1"),
			Port:         getEnvInt("SERVER_PORT", 8000),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		WAF: WAFConfig{
			BaseURL:    getEnv("WAF_API_BASE_URL", "http://localhost:2342"),
			APIKey:     getEnv("WAF_API_KEY", ""),
			Timeout:    getEnvDuration("WAF_API_TIMEOUT", 30*time.Second),
			RetryTimes: getEnvInt("WAF_RETRY_TIMES", 3),
			RetryDelay: getEnvDuration("WAF_RETRY_DELAY", 1*time.Second),
		},
		Database: DatabaseConfig{
			Driver:      getEnv("DB_DRIVER", "postgres"),
			DSN:         getEnv("DB_DSN", ""),
			MaxOpenConn: getEnvInt("DB_MAX_OPEN_CONN", 25),
			MaxIdleConn: getEnvInt("DB_MAX_IDLE_CONN", 5),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			File:  getEnv("LOG_FILE", ""),
		},
		Cache: CacheConfig{
			TTL:             getEnvDuration("CACHE_TTL", 5*time.Minute),
			MaxSize:         getEnvInt("CACHE_MAX_SIZE", 10000),
			CleanupInterval: getEnvDuration("CACHE_CLEANUP", 10*time.Minute),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	return defaultVal
}
```

### 2.2 模型定义 (internal/models/)

```go
// models/blocked_ip.go
package models

import "time"

type BlockedIP struct {
	ID                  string     `json:"id" db:"id"`
	IPAddress           string     `json:"ip_address" db:"ip_address"`
	Reason              string     `json:"reason" db:"reason"`
	Severity            string     `json:"severity" db:"severity"`
	DurationSeconds     int        `json:"duration_seconds" db:"duration_seconds"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt           *time.Time `json:"expires_at" db:"expires_at"`
	Tags                []string   `json:"tags" db:"tags"`
	BlockedRequestCount int        `json:"blocked_request_count" db:"blocked_request_count"`
	LastAttackTime      *time.Time `json:"last_attack_time" db:"last_attack_time"`
}

// models/rule.go
package models

import "time"

type RuleConditions struct {
	Method              []string               `json:"method"`
	Path                string                 `json:"path"`
	SourceIP            string                 `json:"source_ip"`
	UserAgent           string                 `json:"user_agent"`
	RequestBodyContains string                 `json:"request_body_contains"`
	ResponseCode        []int                  `json:"response_code"`
	CountryCode         []string               `json:"country_code"`
	RateLimit           map[string]interface{} `json:"rate_limit"`
}

type WAFRule struct {
	ID           string         `json:"id" db:"id"`
	RuleName     string         `json:"rule_name" db:"rule_name"`
	RuleType     string         `json:"rule_type" db:"rule_type"`
	Action       string         `json:"action" db:"action"`
	Conditions   RuleConditions `json:"conditions" db:"conditions"`
	Priority     int            `json:"priority" db:"priority"`
	Enabled      bool           `json:"enabled" db:"enabled"`
	Description  string         `json:"description" db:"description"`
	Tags         []string       `json:"tags" db:"tags"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy    string         `json:"created_by" db:"created_by"`
	UpdatedBy    string         `json:"updated_by" db:"updated_by"`
	IsSystemRule bool           `json:"is_system_rule" db:"is_system_rule"`
}
```

### 2.3 IP黑名单服务 (internal/service/ip_blacklist.go)

```go
package service

import (
	"context"
	"fmt"
	"sync"

	"waf-mcp-go/internal/models"
	"waf-mcp-go/pkg/logger"
)

type IPBlacklistService struct {
	apiClient  *APIClient
	cache      Cache
	logger     logger.Logger
	mu         sync.RWMutex
	blockedIPs map[string]*models.BlockedIP
}

// NewIPBlacklistService 创建IP黑名单服务
func NewIPBlacklistService(apiClient *APIClient, cache Cache, log logger.Logger) *IPBlacklistService {
	return &IPBlacklistService{
		apiClient:  apiClient,
		cache:      cache,
		logger:     log,
		blockedIPs: make(map[string]*models.BlockedIP),
	}
}

// BlockIP 添加IP到黑名单
func (s *IPBlacklistService) BlockIP(ctx context.Context, ipAddr, reason string, durationSeconds int, severity string, tags []string) (*models.BlockedIP, error) {
	// 验证IP地址
	if err := validateIPAddress(ipAddr); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建黑名单项
	blockedIP := &models.BlockedIP{
		ID:              generateID(),
		IPAddress:       ipAddr,
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

	s.blockedIPs[ipAddr] = blockedIP

	// 同步到WAF API
	if err := s.apiClient.BlockIP(ctx, ipAddr, reason, durationSeconds, severity); err != nil {
		s.logger.Warnf("Failed to sync with WAF API: %v", err)
		// 继续运行，本地缓存可用
	}

	// 清空缓存
	s.cache.Delete("blocked_ips_list")

	return blockedIP, nil
}

// UnblockIP 从黑名单移除IP
func (s *IPBlacklistService) UnblockIP(ctx context.Context, ipAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.blockedIPs, ipAddr)

	// 同步到WAF API
	if err := s.apiClient.UnblockIP(ctx, ipAddr); err != nil {
		s.logger.Warnf("Failed to unblock IP in WAF API: %v", err)
	}

	s.cache.Delete("blocked_ips_list")
	return nil
}

// ListBlockedIPs 列出黑名单
func (s *IPBlacklistService) ListBlockedIPs(ctx context.Context, page, pageSize int, filters map[string]interface{}) (*models.ListResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 从缓存获取
	if cached, found := s.cache.Get("blocked_ips_list"); found {
		return cached.(*models.ListResponse), nil
	}

	total := len(s.blockedIPs)
	totalPages := (total + pageSize - 1) / pageSize

	var items []*models.BlockedIP
	for _, ip := range s.blockedIPs {
		if s.matchesFilters(ip, filters) {
			items = append(items, ip)
		}
	}

	resp := &models.ListResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	// 缓存结果
	s.cache.Set("blocked_ips_list", resp)

	return resp, nil
}

// matchesFilters 检查IP是否匹配过滤器
func (s *IPBlacklistService) matchesFilters(ip *models.BlockedIP, filters map[string]interface{}) bool {
	if severity, ok := filters["severity"].(string); ok && ip.Severity != severity {
		return false
	}
	return true
}

// GetStats 获取黑名单统计
func (s *IPBlacklistService) GetStats(ctx context.Context) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]int)
	for _, ip := range s.blockedIPs {
		stats[ip.Severity]++
	}

	return map[string]interface{}{
		"total_blocked":   len(s.blockedIPs),
		"stats_by_severity": stats,
	}
}
```

### 2.4 规则管理服务 (internal/service/rule_manager.go)

```go
package service

import (
	"context"
	"fmt"
	"sync"

	"waf-mcp-go/internal/models"
	"waf-mcp-go/pkg/logger"
)

type RuleManagerService struct {
	apiClient *APIClient
	cache     Cache
	logger    logger.Logger
	mu        sync.RWMutex
	rules     map[string]*models.WAFRule
}

// NewRuleManagerService 创建规则管理服务
func NewRuleManagerService(apiClient *APIClient, cache Cache, log logger.Logger) *RuleManagerService {
	return &RuleManagerService{
		apiClient: apiClient,
		cache:     cache,
		logger:    log,
		rules:     make(map[string]*models.WAFRule),
	}
}

// CreateRule 创建规则
func (s *RuleManagerService) CreateRule(ctx context.Context, ruleName, ruleType, action string, conditions models.RuleConditions, priority int) (*models.WAFRule, error) {
	// 验证优先级
	if priority < 1 || priority > 10000 {
		return nil, fmt.Errorf("priority must be between 1 and 10000")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	rule := &models.WAFRule{
		ID:         generateID(),
		RuleName:   ruleName,
		RuleType:   ruleType,
		Action:     action,
		Conditions: conditions,
		Priority:   priority,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.rules[rule.ID] = rule

	// 同步到WAF API
	if err := s.apiClient.CreateRule(ctx, ruleName, ruleType, action, conditions, priority); err != nil {
		s.logger.Warnf("Failed to sync rule with WAF API: %v", err)
	}

	s.cache.Delete("rules_list")
	return rule, nil
}

// UpdateRule 更新规则
func (s *RuleManagerService) UpdateRule(ctx context.Context, ruleID string, updates map[string]interface{}) (*models.WAFRule, error) {
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

	// 同步到WAF API
	if err := s.apiClient.UpdateRule(ctx, ruleID, updates); err != nil {
		s.logger.Warnf("Failed to update rule in WAF API: %v", err)
	}

	s.cache.Delete("rules_list")
	return rule, nil
}

// DeleteRule 删除规则
func (s *RuleManagerService) DeleteRule(ctx context.Context, ruleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.rules[ruleID]; !ok {
		return fmt.Errorf("rule not found")
	}

	delete(s.rules, ruleID)

	// 同步到WAF API
	if err := s.apiClient.DeleteRule(ctx, ruleID); err != nil {
		s.logger.Warnf("Failed to delete rule in WAF API: %v", err)
	}

	s.cache.Delete("rules_list")
	return nil
}

// ListRules 列出规则
func (s *RuleManagerService) ListRules(ctx context.Context, page, pageSize int) (*models.ListResponse, error) {
	// 从缓存获取
	if cached, found := s.cache.Get("rules_list"); found {
		return cached.(*models.ListResponse), nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.rules)
	totalPages := (total + pageSize - 1) / pageSize

	var items []*models.WAFRule
	for _, rule := range s.rules {
		items = append(items, rule)
	}

	resp := &models.ListResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	s.cache.Set("rules_list", resp)
	return resp, nil
}
```

---

## 第三部分：HTTP处理程序与路由

### 3.1 HTTP处理程序 (internal/api/handlers.go)

```go
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"waf-mcp-go/internal/models"
	"waf-mcp-go/internal/service"
	"waf-mcp-go/pkg/logger"
)

type Handler struct {
	ipBlacklistService *service.IPBlacklistService
	ruleService        *service.RuleManagerService
	logger             logger.Logger
}

// NewHandler 创建处理程序
func NewHandler(ipService *service.IPBlacklistService, ruleService *service.RuleManagerService, log logger.Logger) *Handler {
	return &Handler{
		ipBlacklistService: ipService,
		ruleService:        ruleService,
		logger:             log,
	}
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
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 默认值
	if req.DurationSeconds == 0 {
		req.DurationSeconds = 3600
	}
	if req.Severity == "" {
		req.Severity = "medium"
	}

	blockedIP, err := h.ipBlacklistService.BlockIP(r.Context(), req.IPAddress, req.Reason, req.DurationSeconds, req.Severity, req.Tags)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, blockedIP, "IP blocked successfully")
}

// HandleListBlockedIPs 处理黑名单查询
func (h *Handler) HandleListBlockedIPs(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}

	pageSize := 20
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil {
			pageSize = val
		}
	}

	filters := make(map[string]interface{})
	if severity := r.URL.Query().Get("severity"); severity != "" {
		filters["severity"] = severity
	}

	list, err := h.ipBlacklistService.ListBlockedIPs(r.Context(), page, pageSize, filters)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, list, "OK")
}

// HandleCreateRule 处理规则创建
func (h *Handler) HandleCreateRule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RuleName   string                 `json:"rule_name"`
		RuleType   string                 `json:"rule_type"`
		Action     string                 `json:"action"`
		Conditions models.RuleConditions  `json:"conditions"`
		Priority   int                    `json:"priority"`
		Enabled    bool                   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	rule, err := h.ruleService.CreateRule(r.Context(), req.RuleName, req.RuleType, req.Action, req.Conditions, req.Priority)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, rule, "Rule created successfully")
}

// HandleListRules 处理规则列表查询
func (h *Handler) HandleListRules(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}

	pageSize := 20
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil {
			pageSize = val
		}
	}

	list, err := h.ruleService.ListRules(r.Context(), page, pageSize)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, list, "OK")
}

// respondSuccess 返回成功响应
func (h *Handler) respondSuccess(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
		"message": message,
	})
}

// respondError 返回错误响应
func (h *Handler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
```

### 3.2 路由设置

```go
// main.go
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"waf-mcp-go/cmd/server/config"
	"waf-mcp-go/internal/api"
	"waf-mcp-go/internal/service"
	"waf-mcp-go/pkg/logger"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	log := logger.NewLogger(cfg.Logger.Level)

	// 初始化缓存
	cache := service.NewCache(cfg.Cache.TTL, cfg.Cache.MaxSize)

	// 初始化API客户端
	apiClient := api.NewClient(cfg.WAF.BaseURL, cfg.WAF.APIKey, cfg.WAF.Timeout, cfg.WAF.RetryTimes)

	// 初始化服务
	ipBlacklistService := service.NewIPBlacklistService(apiClient, cache, log)
	ruleService := service.NewRuleManagerService(apiClient, cache, log)

	// 初始化处理程序
	handler := api.NewHandler(ipBlacklistService, ruleService, log)

	// 设置路由
	router := mux.NewRouter()

	// IP黑名单路由
	router.HandleFunc("/waf/block-ip", handler.HandleBlockIP).Methods("POST")
	router.HandleFunc("/waf/blocked-ips", handler.HandleListBlockedIPs).Methods("GET")
	router.HandleFunc("/waf/unblock-ip/{ip}", handler.HandleUnblockIP).Methods("POST")

	// 规则管理路由
	router.HandleFunc("/waf/rules", handler.HandleCreateRule).Methods("POST")
	router.HandleFunc("/waf/rules", handler.HandleListRules).Methods("GET")
	router.HandleFunc("/waf/rules/{id}", handler.HandleGetRule).Methods("GET")
	router.HandleFunc("/waf/rules/{id}", handler.HandleUpdateRule).Methods("PUT")
	router.HandleFunc("/waf/rules/{id}", handler.HandleDeleteRule).Methods("DELETE")

	// 健康检查
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	}).Methods("GET")

	// 启动服务器
	addr := cfg.Server.Host + ":" + string(rune(cfg.Server.Port))
	log.Infof("Starting WAF MCP server on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

---

## 第四部分：测试

### 4.1 单元测试示例

```go
// tests/unit/ip_blacklist_test.go
package unit

import (
	"context"
	"testing"

	"waf-mcp-go/internal/service"
)

func TestBlockIP(t *testing.T) {
	// 初始化测试环境
	cache := service.NewCache(1, 100)
	apiClient := &mockAPIClient{}
	service := service.NewIPBlacklistService(apiClient, cache, nil)

	// 测试
	blockedIP, err := service.BlockIP(context.Background(), "192.168.1.1", "Test", 3600, "high", []string{})
	if err != nil {
		t.Fatalf("BlockIP failed: %v", err)
	}

	if blockedIP.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", blockedIP.IPAddress)
	}
}

func TestListBlockedIPs(t *testing.T) {
	cache := service.NewCache(1, 100)
	apiClient := &mockAPIClient{}
	service := service.NewIPBlacklistService(apiClient, cache, nil)

	// 添加测试数据
	service.BlockIP(context.Background(), "192.168.1.1", "Test", 3600, "high", []string{})
	service.BlockIP(context.Background(), "10.0.0.1", "Test", 3600, "medium", []string{})

	// 查询
	list, err := service.ListBlockedIPs(context.Background(), 1, 20, nil)
	if err != nil {
		t.Fatalf("ListBlockedIPs failed: %v", err)
	}

	if list.Total != 2 {
		t.Errorf("Expected 2 items, got %d", list.Total)
	}
}

type mockAPIClient struct{}

func (m *mockAPIClient) BlockIP(ctx context.Context, ip, reason string, duration int, severity string) error {
	return nil
}

func (m *mockAPIClient) UnblockIP(ctx context.Context, ip string) error {
	return nil
}
```

---

## 总结

本指南提供了完整的Go语言WAF MCP服务器实现方案，包括：

1. **项目结构** - 符合Go最佳实践的目录组织
2. **核心服务** - IP黑名单、规则管理等业务逻辑
3. **API层** - HTTP处理程序和客户端
4. **缓存机制** - 提高性能和降低延迟
5. **错误处理** - 全面的异常处理和恢复
6. **测试** - 单元测试和集成测试

建议后续步骤：
- 完成数据库层实现
- 添加认证和授权
- 实现监控和指标
- 部署到生产环境
