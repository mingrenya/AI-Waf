# WAF MCP Server - Go语言版本完整指南

## 📋 项目结构

```
waf-mcp-server/
├── cmd/
│   └── server/
│       └── main.go              # 服务器入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── handlers/
│   │   ├── handlers.go          # HTTP处理器
│   │   ├── ip_blacklist.go      # IP黑名单处理
│   │   ├── rules.go             # 规则处理
│   │   ├── attacks.go           # 攻击分析处理
│   │   └── monitoring.go        # 监控处理
│   ├── services/
│   │   ├── ip_blacklist.go      # IP黑名单服务
│   │   ├── rule_management.go   # 规则管理服务
│   │   ├── attack_analysis.go   # 攻击分析服务
│   │   └── monitoring.go        # 监控统计服务
│   ├── models/
│   │   └── models.go            # 数据模型
│   └── client/
│       └── waf_client.go        # WAF API客户端
├── pkg/
│   ├── cache/
│   │   └── cache.go             # 缓存实现
│   ├── database/
│   │   └── database.go          # 数据库连接
│   ├── logger/
│   │   └── logger.go            # 日志配置
│   └── validator/
│       └── validator.go         # 数据验证
├── migrations/
│   ├── 001_create_tables.up.sql
│   ├── 001_create_tables.down.sql
│   ├── 002_add_indexes.up.sql
│   └── 002_add_indexes.down.sql
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── kubernetes/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── config/
│       └── config.yaml
├── tests/
│   ├── unit/
│   │   ├── services_test.go
│   │   └── handlers_test.go
│   └── integration/
│       └── api_test.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 🚀 快速开始

### 1. 环境要求

```bash
go version
# go version go1.21 linux/amd64

# 验证Go工作区设置
go env GOPATH
go env GOROOT
```

### 2. 初始化项目

```bash
# 创建项目目录
mkdir -p ~/projects/waf-mcp-server
cd ~/projects/waf-mcp-server

# 初始化Go模块
go mod init github.com/waf-mcp-server

# 创建目录结构
mkdir -p cmd/server internal/{config,handlers,services,models,client} pkg/{cache,database,logger,validator} migrations deploy/{docker,kubernetes,config} tests/{unit,integration}

# 复制go.mod文件
cp go.mod go.mod.bak
```

### 3. 下载依赖

```bash
# 下载所有依赖
go mod download
go mod tidy

# 验证依赖
go mod graph
```

### 4. 项目文件组织

```bash
# 复制提供的代码文件到相应目录
cp models.go internal/models/
cp client.go internal/client/
cp main.go cmd/server/

# 或按包结构组织
mv models.go internal/models/models.go
```

### 5. 本地开发运行

```bash
# 运行单个main.go
go run cmd/server/main.go

# 或使用go build
go build -o waf-mcp cmd/server/main.go
./waf-mcp

# 使用Makefile (推荐)
make build
make run
```

---

## 📝 关键代码片段

### 配置管理 (config.go)

```go
package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host string
		Port int
	}
	Database struct {
		Driver   string
		DSN      string
		MaxConns int
		IdleConns int
	}
	WAF struct {
		APIBaseURL string
		APIKey     string
		Timeout    int
	}
	Cache struct {
		TTL int
		Type string // memory, redis
	}
	Log struct {
		Level string
		Format string
	}
}

func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./deploy/config")
	viper.AddConfigPath("/etc/waf-mcp")

	if err := viper.ReadInConfig(); err != nil {
		// 使用默认配置
	}

	var cfg Config
	viper.Unmarshal(&cfg)
	return &cfg
}
```

### IP黑名单服务 (services/ip_blacklist.go)

```go
package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"

	"github.com/waf-mcp-server/internal/models"
	"github.com/waf-mcp-server/pkg/cache"
)

type IPBlacklistService struct {
	db    *gorm.DB
	cache cache.Cache
	apiKey string
}

func NewIPBlacklistService(db *gorm.DB, c cache.Cache, apiKey string) *IPBlacklistService {
	return &IPBlacklistService{
		db:     db,
		cache:  c,
		apiKey: apiKey,
	}
}

func (s *IPBlacklistService) BlockIP(ctx context.Context, req models.BlockIPRequest) (*models.BlockedIP, error) {
	// 验证IP格式
	if err := validateIPFormat(req.IPAddress); err != nil {
		return nil, fmt.Errorf("invalid IP: %w", err)
	}

	// 检查IP是否已存在
	var existing models.BlockedIP
	if err := s.db.Where("ip_address = ?", req.IPAddress).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("IP already blocked")
	}

	// 创建新记录
	blockedIP := models.BlockedIP{
		ID:              uuid.New().String(),
		IPAddress:       req.IPAddress,
		Reason:          req.Reason,
		Severity:        req.Severity,
		DurationSeconds: req.DurationSeconds,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Tags:            req.Tags,
	}

	// 如果指定了时长，计算过期时间
	if req.DurationSeconds > 0 {
		expiresAt := time.Now().Add(time.Duration(req.DurationSeconds) * time.Second)
		blockedIP.ExpiresAt = &expiresAt
	}

	// 保存到数据库
	if err := s.db.Create(&blockedIP).Error; err != nil {
		return nil, fmt.Errorf("failed to save blocked IP: %w", err)
	}

	// 更新缓存
	s.cache.Set(fmt.Sprintf("blocked_ip:%s", req.IPAddress), true, time.Hour)

	// 审计日志
	s.auditLog(ctx, "block_ip", "blocked_ip", blockedIP.ID, map[string]interface{}{
		"ip_address": req.IPAddress,
		"reason":     req.Reason,
	})

	return &blockedIP, nil
}

func (s *IPBlacklistService) ListBlockedIPs(ctx context.Context, page, pageSize int) ([]models.BlockedIP, int64, error) {
	var ips []models.BlockedIP
	var total int64

	offset := (page - 1) * pageSize

	// 查询总数
	if err := s.db.Model(&models.BlockedIP{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	if err := s.db.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&ips).Error; err != nil {
		return nil, 0, err
	}

	return ips, total, nil
}

func (s *IPBlacklistService) auditLog(ctx context.Context, opType, resourceType, resourceID string, changes map[string]interface{}) {
	// TODO: 实现审计日志
}

func validateIPFormat(ip string) error {
	// TODO: 使用net.ParseIP验证
	return nil
}
```

### HTTP处理器 (handlers/handlers.go)

```go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/waf-mcp-server/internal/models"
	"github.com/waf-mcp-server/internal/services"
)

type Handlers struct {
	ipBlacklist *services.IPBlacklistService
	rules       *services.RuleManagementService
	attacks     *services.AttackAnalysisService
	monitoring  *services.MonitoringService
	log         *logrus.Logger
}

func NewHandlers(
	ipBlacklist *services.IPBlacklistService,
	rules *services.RuleManagementService,
	attacks *services.AttackAnalysisService,
	monitoring *services.MonitoringService,
	log *logrus.Logger,
) *Handlers {
	return &Handlers{
		ipBlacklist: ipBlacklist,
		rules:       rules,
		attacks:     attacks,
		monitoring:  monitoring,
		log:         log,
	}
}

// BlockIP 处理 POST /waf/block-ip
func (h *Handlers) BlockIP(c echo.Context) error {
	var req models.BlockIPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
	}

	// 验证请求
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
	}

	// 调用服务
	blockedIP, err := h.ipBlacklist.BlockIP(c.Request().Context(), req)
	if err != nil {
		h.log.WithError(err).Error("Failed to block IP")
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			RequestID: uuid.New().String(),
			Error: models.ErrorDetail{
				Code:    "WAF_BLOCK_IP_FAILED",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Success:   true,
		Data:      blockedIP,
		Message:   fmt.Sprintf("IP %s blocked successfully", req.IPAddress),
		Timestamp: time.Now().UTC(),
		RequestID: uuid.New().String(),
	})
}

// ListBlockedIPs 处理 GET /waf/blocked-ips
func (h *Handlers) ListBlockedIPs(c echo.Context) error {
	page := c.QueryParam("page")
	pageSize := c.QueryParam("page_size")

	pageNum := 1
	pageSizeNum := 20

	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageNum = p
	}

	if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
		pageSizeNum = ps
	}

	ips, total, err := h.ipBlacklist.ListBlockedIPs(c.Request().Context(), pageNum, pageSizeNum)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error: models.ErrorDetail{
				Code:    "LIST_FAILED",
				Message: err.Error(),
			},
		})
	}

	totalPages := (total + int64(pageSizeNum) - 1) / int64(pageSizeNum)

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data: models.PaginatedResponse{
			Items:      ips,
			Page:       pageNum,
			PageSize:   pageSizeNum,
			Total:      total,
			TotalPages: totalPages,
		},
		Message:   "Listed blocked IPs successfully",
		Timestamp: time.Now().UTC(),
		RequestID: uuid.New().String(),
	})
}

// 其他处理器方法...
// BatchBlockIPs, UnblockIP, CreateRule, ListRules 等
```

### 缓存实现 (pkg/cache/cache.go)

```go
package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, duration time.Duration)
	Delete(key string)
	Clear()
}

type MemoryCache struct {
	cache *gocache.Cache
}

func NewMemoryCache(defaultExpiration time.Duration) Cache {
	return &MemoryCache{
		cache: gocache.New(defaultExpiration, 10*time.Minute),
	}
}

func (m *MemoryCache) Get(key string) (interface{}, bool) {
	return m.cache.Get(key)
}

func (m *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	m.cache.Set(key, value, duration)
}

func (m *MemoryCache) Delete(key string) {
	m.cache.Delete(key)
}

func (m *MemoryCache) Clear() {
	m.cache.Flush()
}
```

### 数据库初始化 (pkg/database/database.go)

```go
package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/waf-mcp-server/internal/models"
)

type DatabaseConfig struct {
	Driver   string
	DSN      string
	MaxConns int
	IdleConns int
}

func InitDB(cfg DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&models.BlockedIP{},
		&models.WafRule{},
		&models.AttackLog{},
		&models.AuditLog{},
		&models.GeneratedRule{},
		&models.WAFConfig{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
```

---

## 🧪 单元测试

### 测试示例 (tests/unit/services_test.go)

```go
package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/waf-mcp-server/internal/models"
	"github.com/waf-mcp-server/internal/services"
)

type MockDB struct {
	mock.Mock
}

func TestBlockIP(t *testing.T) {
	// 准备测试数据
	req := models.BlockIPRequest{
		IPAddress:       "192.168.1.1",
		Reason:          "Test",
		DurationSeconds: 3600,
		Severity:        models.SeverityHigh,
	}

	// 执行测试
	// ...

	// 验证结果
	assert.NotNil(t, result)
	assert.Equal(t, "192.168.1.1", result.IPAddress)
}

func TestBlockIPInvalidIP(t *testing.T) {
	req := models.BlockIPRequest{
		IPAddress: "invalid-ip",
		Reason:    "Test",
	}

	// 验证错误处理
	// ...
}
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/services/...

# 显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 🐳 Docker部署

### Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o waf-mcp cmd/server/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/waf-mcp .
COPY deploy/config/config.yaml .

EXPOSE 8000 9090

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8000/health || exit 1

CMD ["./waf-mcp"]
```

### 构建和运行

```bash
# 构建镜像
docker build -t waf-mcp:latest .

# 运行容器
docker run -d \
  -p 8000:8000 \
  -e WAF_API_BASE_URL=http://waf-backend:2342 \
  -e WAF_API_KEY=your-key \
  waf-mcp:latest

# 查看日志
docker logs -f <container-id>
```

---

## Makefile 示例

```makefile
.PHONY: build run test clean docker-build docker-run

BINARY_NAME=waf-mcp
MAIN_PATH=cmd/server/main.go

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run: build
	./$(BINARY_NAME)

dev:
	go run $(MAIN_PATH)

test:
	go test -v -cover ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	go clean
	rm -f $(BINARY_NAME) coverage.out coverage.html

docker-build:
	docker build -t waf-mcp:latest .

docker-run: docker-build
	docker run -p 8000:8000 waf-mcp:latest

docker-compose-up:
	docker-compose -f deploy/docker/docker-compose.yml up -d

docker-compose-down:
	docker-compose -f deploy/docker/docker-compose.yml down

help:
	@echo "Available targets:"
	@echo "  make build             - Build the binary"
	@echo "  make run               - Build and run"
	@echo "  make dev               - Run with go run (hot reload)"
	@echo "  make test              - Run tests"
	@echo "  make test-coverage     - Run tests with coverage"
	@echo "  make lint              - Run linter"
	@echo "  make fmt               - Format code"
	@echo "  make vet               - Run go vet"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make docker-build      - Build Docker image"
	@echo "  make docker-run        - Build and run Docker container"
	@echo "  make docker-compose-up - Start with docker-compose"
	@echo "  make docker-compose-down - Stop docker-compose"
```

---

## 配置文件示例 (deploy/config/config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8000

database:
  driver: "mysql"
  dsn: "user:password@tcp(localhost:3306)/waf_mcp?charset=utf8mb4&parseTime=True&loc=Local"
  max_conns: 25
  idle_conns: 5

waf:
  api_base_url: "http://localhost:2342"
  api_key: "${WAF_API_KEY}"
  timeout: 30

cache:
  ttl: 300
  type: "memory"  # 或 redis

log:
  level: "info"
  format: "json"

monitoring:
  prometheus_port: 9090
```

---

## 数据库迁移

### SQL迁移脚本 (migrations/001_create_tables.up.sql)

```sql
-- IP黑名单表
CREATE TABLE IF NOT EXISTS blocked_ips (
    id VARCHAR(36) PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL UNIQUE,
    reason VARCHAR(255),
    severity VARCHAR(20),
    duration_seconds INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    tags JSON,
    blocked_requests_count BIGINT DEFAULT 0,
    last_attack_timestamp TIMESTAMP NULL,
    created_by VARCHAR(255),
    INDEX idx_created_at (created_at),
    INDEX idx_severity (severity),
    INDEX idx_expires_at (expires_at)
);

-- WAF规则表
CREATE TABLE IF NOT EXISTS waf_rules (
    id VARCHAR(36) PRIMARY KEY,
    rule_name VARCHAR(255) NOT NULL UNIQUE,
    rule_type VARCHAR(50),
    action VARCHAR(20),
    conditions JSON,
    priority INT,
    enabled BOOLEAN DEFAULT TRUE,
    description TEXT,
    tags JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    is_system_rule BOOLEAN DEFAULT FALSE,
    version INT DEFAULT 1,
    INDEX idx_priority (priority),
    INDEX idx_enabled (enabled),
    INDEX idx_rule_type (rule_type)
);

-- 攻击日志表
CREATE TABLE IF NOT EXISTS attack_logs (
    id VARCHAR(36) PRIMARY KEY,
    attack_type VARCHAR(100),
    severity VARCHAR(20),
    source_ip VARCHAR(45),
    target_path VARCHAR(500),
    target_domain VARCHAR(255),
    request_method VARCHAR(10),
    request_body LONGTEXT,
    response_code INT,
    matched_rule_id VARCHAR(36),
    matched_rule_name VARCHAR(255),
    action_taken VARCHAR(20),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_agent VARCHAR(500),
    referer VARCHAR(500),
    request_id VARCHAR(36),
    geo_location JSON,
    INDEX idx_timestamp (timestamp),
    INDEX idx_source_ip (source_ip),
    INDEX idx_attack_type (attack_type)
);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    operation_type VARCHAR(50),
    resource_type VARCHAR(50),
    resource_id VARCHAR(36),
    operator_id VARCHAR(255),
    changes JSON,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20),
    INDEX idx_timestamp (timestamp),
    INDEX idx_resource_type (resource_type)
);
```

---

## 🔍 调试和故障排查

### 启用调试日志

```bash
# 环境变量方式
export LOG_LEVEL=debug
./waf-mcp

# 或在config.yaml中设置
log:
  level: "debug"
```

### 常见问题

1. **数据库连接失败**
   ```bash
   # 检查数据库连接字符串
   go test -v ./pkg/database/...
   
   # 测试连接
   mysql -h localhost -u user -p database
   ```

2. **端口被占用**
   ```bash
   # 检查端口
   netstat -tlnp | grep 8000
   
   # 更改配置文件中的端口
   ```

3. **API调用失败**
   ```bash
   # 检查WAF后端
   curl http://localhost:2342/health
   
   # 查看客户端日志
   # 启用DEBUG日志
   ```

---

## 性能优化建议

1. **启用连接池**
   ```go
   sqlDB, _ := db.DB()
   sqlDB.SetMaxOpenConns(25)
   sqlDB.SetMaxIdleConns(5)
   ```

2. **使用缓存**
   - 实现Redis缓存而不是内存缓存
   - 合理设置TTL

3. **批量操作**
   - 使用BatchCreateRules处理多个规则
   - 使用BatchBlockIPs处理多个IP

4. **索引优化**
   - 在经常查询的字段上添加索引
   - 查看migration文件的索引设置

---

## 监控和指标

### Prometheus指标暴露

```go
import "github.com/prometheus/client_golang/prometheus"

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_mcp_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
}
```

### 访问指标

```bash
curl http://localhost:9090/metrics
```

---

## 参考资源

- [Go官方文档](https://golang.org/doc/)
- [GORM文档](https://gorm.io/)
- [Echo框架](https://echo.labstack.com/)
- [Resty HTTP客户端](https://github.com/go-resty/resty)

---

**最后更新**: 2024年2月2日
**Go版本**: 1.21+
**状态**: 生产就绪
