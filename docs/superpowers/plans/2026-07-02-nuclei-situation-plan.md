# Nuclei 集成 + 态势感知 + 快速处置 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 AI-Waf 添加 Nuclei 漏洞扫描集成、基于 Loki/LogQL 的态势感知系统、以及日志界面的快速处置联动能力。

**Architecture:** 态势感知核心通过 LogQL 查询 Loki 做事件检测和攻击链构建，Redis 处理实时数据流和 Pub/Sub 推送，MongoDB 持久化攻击链和画像数据。Nuclei 通过官方 Go SDK 直接嵌入 server 进程。WebSocket 实现前后端实时通信。

**Tech Stack:** Go 1.24.1 + Gin + MongoDB (mongo-driver/v2) + Redis (go-redis/v9) + Loki (HTTP API) + nuclei/v3/lib + gorilla/websocket + React 18 + TypeScript

**设计文档:** `docs/superpowers/specs/2026-07-02-nuclei-situation-design.md`

## Global Constraints

- Go 版本最低 1.24.1，使用 go.work workspace 管理四个模块（coraza-spoa, mcp-server, pkg, server）
- 前端使用 React 18 + TypeScript + Vite + TailwindCSS + shadcn/ui (Radix)
- MongoDB 驱动使用 `go.mongodb.org/mongo-driver/v2`（v2 API，非 v1）
- 所有后端响应统一使用 `model.APIResponse`（`model.NewSuccessResponse` / `model.NewErrorResponse`）
- 控制器/服务/仓库使用接口+实现体命名模式：`XxxController` 接口 + `XxxControllerImpl` 实现体
- 日志使用 `zerolog`，通过 `config.GetControllerLogger(name)` / `config.GetServiceLogger(name)` / `config.GetRepositoryLogger(name)` 获取
- Redis 连接复用已有 `middleware.NewRedisRateLimiter` 的连接模式，环境变量 `REDIS_ADDR` 控制地址
- 前端 API 层文件放 `web/src/api/`，组件放 `web/src/feature/<name>/components/`，页面放 `web/src/pages/<name>/`
- Docker Compose 中新增的 Redis 容器使用服务名 `redis`，内部端口 6379
- 中文注释和错误消息
- 提交使用 Conventional Commits 格式

---

## Phase 0: 基础设施 (P0)

### Task P0-1: Redis 集成 — Docker Compose + 环境变量 + 连接池

**Files:**
- Modify: `docker-compose.yaml`
- Modify: `.env.template`
- Modify: `server/config/config.go`
- Create: `pkg/database/redis/redis.go`

**Interfaces:**
- Consumes: (none — first task)
- Produces:
  - `redis.NewClient(ctx) (*redis.Client, error)` — 创建 Redis 客户端，从环境变量读取地址
  - `config.Global.RedisConfig.Enabled bool` — Redis 开关
  - Docker Compose 中 `redis` 服务，使用 `waf-network`

**Steps:**

- [ ] **Step 1: 在 docker-compose.yaml 中添加 Redis 服务**

在 `docker-compose.yaml` 的 `services:` 块末尾（`networks:` 之前）添加：

```yaml
  # Redis — 实时数据层
  redis:
    image: redis:7-alpine
    container_name: waf-redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - waf-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
```

在 `volumes:` 块末尾添加：

```yaml
  redis_data:
    driver: local
```

- [ ] **Step 2: 更新 .env.template 添加 Redis 配置**

在 `.env.template` 末尾添加：

```
# Redis 配置
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
```

- [ ] **Step 3: 创建 pkg/database/redis/redis.go**

```go
package redis

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	once   sync.Once
	initErr error
)

// GetClient 获取 Redis 客户端单例。
// 若 REDIS_ADDR 未配置，返回 nil（Redis 不可用，系统降级运行）。
func GetClient(ctx context.Context) (*redis.Client, error) {
	once.Do(func() {
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			initErr = fmt.Errorf("REDIS_ADDR not configured, Redis disabled")
			return
		}
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("Redis ping failed: %w", err)
			client.Close()
			client = nil
		}
	})
	return client, initErr
}

// IsAvailable 检查 Redis 是否可用
func IsAvailable(ctx context.Context) bool {
	c, err := GetClient(ctx)
	return err == nil && c != nil
}
```

- [ ] **Step 4: 更新 server/config/config.go 添加 RedisConfig**

在 `Config` 结构体中添加：

```go
type RedisConfig struct {
	Addr     string
	Password string
	Enabled  bool
}
```

在 `Config` struct 的字段区（`IsK8s` 之后）添加：

```go
RedisConfig RedisConfig
```

在 `defaultConfig` 中初始化为：

```go
RedisConfig: RedisConfig{
	Addr:    "",
	Enabled: false,
},
```

在 `loadFromEnv` 方法中添加（`Global.IsK8s = env == "true"` 后面）：

```go
if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
	Global.RedisConfig.Addr = redisAddr
	Global.RedisConfig.Password = os.Getenv("REDIS_PASSWORD")
	Global.RedisConfig.Enabled = true
}
```

- [ ] **Step 5: 在 mrya 容器配置中添加 REDIS_ADDR 环境变量**

修改 `docker-compose.yaml` 中 `mrya` 服务的 `environment:` 块，添加：

```yaml
    REDIS_ADDR: redis:6379
    REDIS_PASSWORD: ${REDIS_PASSWORD}
```

- [ ] **Step 6: Commit**

```bash
git add docker-compose.yaml .env.template pkg/database/redis/ server/config/config.go
git commit -m "feat(infra): add Redis container and Go client integration
- Add redis:7-alpine service to docker-compose
- Add REDIS_ADDR/REDIS_PASSWORD env vars
- Create pkg/database/redis client singleton
- Add RedisConfig to server config"
```

---

### Task P0-2: WebSocket 基础设施

**Files:**
- Create: `server/websocket/hub.go`
- Create: `server/websocket/client.go`
- Create: `server/websocket/handler.go`
- Modify: `server/router/router.go`
- Modify: `server/go.mod`

**Interfaces:**
- Consumes: Redis client from Task P0-1
- Produces:
  - `websocket.Hub` — 连接管理中心（单例）
  - `websocket.WSHandler(c *gin.Context)` — Gin WebSocket 升级处理器
  - `GET /api/v1/ws` — WebSocket 端点（需认证）

**Steps:**

- [ ] **Step 1: 添加 gorilla/websocket 依赖**

```bash
cd server && go get github.com/gorilla/websocket
```

- [ ] **Step 2: 创建 server/websocket/hub.go**

```go
package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
)

// Message WebSocket 推送消息结构
type Message struct {
	Type    string      `json:"type"`    // situation:alert / situation:update / situation:attack / nuclei:progress
	Payload interface{} `json:"payload"`
	Time    time.Time   `json:"time"`
}

// Hub 管理所有 WebSocket 连接
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     zerolog.Logger
}

var (
	hubInstance *Hub
	hubOnce     sync.Once
)

// GetHub 获取 Hub 单例
func GetHub() *Hub {
	hubOnce.Do(func() {
		hubInstance = &Hub{
			clients:    make(map[*Client]bool),
			broadcast:  make(chan Message, 256),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			logger:     config.GetServiceLogger("websocket"),
		}
		go hubInstance.run()
	})
	return hubInstance
}

// run Hub 主循环
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast 广播消息到所有连接的客户端
func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- msg
}

// BroadcastJSON 将任意数据序列化后广播
func (h *Hub) BroadcastJSON(msgType string, payload interface{}) {
	msg := Message{Type: msgType, Payload: payload, Time: time.Now()}
	h.broadcast <- msg
}

// ClientCount 返回当前连接数
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// NoOp 空操作，用于检查 hub 是否已初始化
func NoOp() {}

// NewHubForTest 创建测试用 Hub（不启动 run 循环，避免 goroutine 泄漏）
func NewHubForTest(logger zerolog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// 编译时确保 Message 可序列化
var _ json.Marshaler = Message{}
```

- [ ] **Step 3: 创建 server/websocket/client.go**

```go
package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Client 表示单个 WebSocket 连接
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan Message
	userID string
	role   string
}

// readPump 从 WebSocket 连接读取消息（仅处理 ping/pong/close）
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.hub.logger.Warn().Err(err).Str("userID", c.userID).Msg("WebSocket read error")
			}
			break
		}
	}
}

// writePump 向 WebSocket 连接写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, _ := json.Marshal(msg)
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
```

- [ ] **Step 4: 创建 server/websocket/handler.go**

```go
package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 认证已在 Gin 中间件完成
	},
}

// WSHandler 将 HTTP 连接升级为 WebSocket
func WSHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	client := &Client{
		hub:    GetHub(),
		conn:   conn,
		send:   make(chan Message, 64),
		userID: userID.(string),
		role:   role.(string),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
```

- [ ] **Step 5: 在 router.go 中注册 WebSocket 路由**

在 `server/router/router.go` 的 `import` 块中添加：

```go
ws "github.com/mingrenya/AI-Waf/server/websocket"
```

在 `authenticated := api.Group("")` 块内（与其他路由平级）添加：

```go
// WebSocket 实时推送
wsRoutes := authenticated.Group("/ws")
{
	wsRoutes.GET("", ws.WSHandler)
}
```

- [ ] **Step 6: Commit**

```bash
git add server/websocket/ server/router/router.go server/go.mod server/go.sum
git commit -m "feat(websocket): add WebSocket hub, client, and handler infrastructure"
```

---

### Task P0-3: Promtail 日志标签标准化

**Files:**
- Modify: `promtail/promtail-config.yml`

**Interfaces:**
- Consumes: (none)
- Produces: Loki 中的结构化标签 `correlation_id`, `source_ip`, `attack_type`, `action`, `severity`, `geo_country`, `rule_id`, `waf_phase`, `site_id`

**Steps:**

- [ ] **Step 1: 更新 promtail pipeline 解析 WAF 日志**

修改 `promtail/promtail-config.yml` 中 `waf-engine` job 的 `pipeline_stages`：

```yaml
    pipeline_stages:
      # 解析 JSON 日志提取 WAF 标签
      - json:
          expressions:
            level: level
            message: message
            component: component
            time: time
            correlation_id: correlation_id
            source_ip: source_ip
            attack_type: attack_type
            action: action
            severity: severity
            geo_country: geo_country
            rule_id: rule_id
            waf_phase: waf_phase
            site_id: site_id
      - timestamp:
          source: time
          format: 'RFC3339'
          fallback_formats: ['Unix']
      - labels:
          level:
          component:
          correlation_id:
          source_ip:
          attack_type:
          action:
          severity:
          geo_country:
          rule_id:
          waf_phase:
          site_id:
```

- [ ] **Step 2: Commit**

```bash
git add promtail/promtail-config.yml
git commit -m "feat(loki): add structured WAF labels to Promtail pipeline

Extract correlation_id, source_ip, attack_type, action, severity,
geo_country, rule_id, waf_phase, site_id from JSON logs as Loki labels
for situation awareness LogQL queries"
```

---

## Phase 1: 态势感知系统 (P0)

### Task P1-1: 态势感知数据模型

**Files:**
- Create: `pkg/model/situation.go`
- Create: `server/dto/situation.go`

**Interfaces:**
- Consumes: (none — new models)
- Produces:
  - `model.AttackStage` — 攻击阶段枚举
  - `model.AttackChain` — 攻击链模型（含 `GetCollectionName() string`）
  - `model.AttackerProfile` — 攻击者画像模型（含 `GetCollectionName() string`）
  - `model.SituationRule` — 检测规则模型（含 `GetCollectionName() string`）
  - `model.SituationSnapshot` — 态势快照模型
  - `dto.*` — 所有 HTTP 请求/响应 DTO

**Steps:**

- [ ] **Step 1: 创建 pkg/model/situation.go**

```go
package model

import "time"

// AttackStage 攻击阶段
type AttackStage string

const (
	StageUnknown          AttackStage = "unknown"
	StageReconnaissance   AttackStage = "reconnaissance"
	StageScanning         AttackStage = "scanning"
	StageExploitation     AttackStage = "exploitation"
	StageLateralMovement  AttackStage = "lateral_movement"
	StageC2               AttackStage = "command_and_control"
	StageExfiltration     AttackStage = "exfiltration"
)

// StageOrder 返回攻击阶段的数值排序（用于判定阶段迁移）
func (s AttackStage) Order() int {
	switch s {
	case StageReconnaissance:
		return 1
	case StageScanning:
		return 2
	case StageExploitation:
		return 3
	case StageLateralMovement:
		return 4
	case StageC2:
		return 4
	case StageExfiltration:
		return 5
	default:
		return 0
	}
}

// StageMapping 攻击类型到 MITRE ATT&CK 阶段映射
var StageMapping = map[string]AttackStage{
	"scanner":               StageReconnaissance,
	"vulnerability_scanner": StageScanning,
	"port_scan":             StageScanning,
	"directory_bruteforce":  StageScanning,
	"sql_injection":         StageExploitation,
	"xss":                   StageExploitation,
	"rce":                   StageExploitation,
	"file_inclusion":        StageExploitation,
	"csrf":                  StageExploitation,
	"ssrf":                  StageExploitation,
	"command_injection":     StageExploitation,
	"webshell":              StageLateralMovement,
	"backdoor":              StageC2,
	"data_leak":             StageExfiltration,
}

// ChainStage 攻击链阶段
type ChainStage struct {
	Stage      AttackStage `json:"stage" bson:"stage"`
	Technique  string      `json:"technique" bson:"technique"`
	DetectedAt time.Time   `json:"detected_at" bson:"detected_at"`
	Evidence   []string    `json:"evidence" bson:"evidence"`
	Confidence float64     `json:"confidence" bson:"confidence"`
}

// AttackChain 攻击链
type AttackChain struct {
	ID             string       `json:"id" bson:"_id"`
	SourceIP       string       `json:"source_ip" bson:"source_ip"`
	GeoCountry     string       `json:"geo_country" bson:"geo_country"`
	Stages         []ChainStage `json:"stages" bson:"stages"`
	CorrelationIDs []string     `json:"correlation_ids" bson:"correlation_ids"`
	RiskScore      int          `json:"risk_score" bson:"risk_score"`
	FirstSeen      time.Time    `json:"first_seen" bson:"first_seen"`
	LastSeen       time.Time    `json:"last_seen" bson:"last_seen"`
	Active         bool         `json:"active" bson:"active"`
	CreatedAt      time.Time    `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" bson:"updated_at"`
}

func (AttackChain) GetCollectionName() string { return "attack_chains" }

// AttackerProfile 攻击者画像
type AttackerProfile struct {
	ID               string    `json:"id" bson:"_id"`
	SourceIP         string    `json:"source_ip" bson:"source_ip"`
	GeoCountry       string    `json:"geo_country" bson:"geo_country"`
	GeoCity          string    `json:"geo_city,omitempty" bson:"geo_city,omitempty"`
	TotalAttacks     int       `json:"total_attacks" bson:"total_attacks"`
	UniqueAttackTypes int     `json:"unique_attack_types" bson:"unique_attack_types"`
	TopAttackType    string    `json:"top_attack_type" bson:"top_attack_type"`
	UniqueTargetSites int     `json:"unique_target_sites" bson:"unique_target_sites"`
	ActiveHours      []int     `json:"active_hours" bson:"active_hours"`
	BurstIntervals   []string  `json:"burst_intervals" bson:"burst_intervals"`
	AttackPhase      string    `json:"attack_phase" bson:"attack_phase"`
	ToolsIdentified  string    `json:"tools_identified" bson:"tools_identified"`
	IsAutomated      bool      `json:"is_automated" bson:"is_automated"`
	IsPersistent     bool      `json:"is_persistent" bson:"is_persistent"`
	RiskScore        int       `json:"risk_score" bson:"risk_score"`
	RiskLabel        string    `json:"risk_label" bson:"risk_label"`
	LastSeen         time.Time `json:"last_seen" bson:"last_seen"`
	FirstSeen        time.Time `json:"first_seen" bson:"first_seen"`
	UpdatedAt        time.Time `json:"updated_at" bson:"updated_at"`
}

func (AttackerProfile) GetCollectionName() string { return "attacker_profiles" }

// SituationRule 态势检测规则
type SituationRule struct {
	ID             string    `json:"id" bson:"_id"`
	Name           string    `json:"name" bson:"name"`
	Stage          string    `json:"stage" bson:"stage"`
	LogQL          string    `json:"logql" bson:"logql"`
	Interval       int       `json:"interval_seconds" bson:"interval_seconds"`
	Threshold      int       `json:"threshold" bson:"threshold"`
	Severity       string    `json:"severity" bson:"severity"`
	MITRETactic    string    `json:"mitre_tactic" bson:"mitre_tactic"`
	MITRETechnique string    `json:"mitre_technique" bson:"mitre_technique"`
	Enabled        bool      `json:"enabled" bson:"enabled"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" bson:"updated_at"`
}

func (SituationRule) GetCollectionName() string { return "situation_rules" }

// SituationSnapshot 态势快照（定时聚合统计，用于趋势分析）
type SituationSnapshot struct {
	ID            string         `json:"id" bson:"_id"`
	Timestamp     time.Time      `json:"timestamp" bson:"timestamp"`
	TotalChains   int            `json:"total_chains" bson:"total_chains"`
	ActiveChains  int            `json:"active_chains" bson:"active_chains"`
	ByAttackType  map[string]int `json:"by_attack_type" bson:"by_attack_type"`
	ByCountry     map[string]int `json:"by_country" bson:"by_country"`
	ByStage       map[string]int `json:"by_stage" bson:"by_stage"`
	OverallRiskScore float64     `json:"overall_risk_score" bson:"overall_risk_score"`
}

func (SituationSnapshot) GetCollectionName() string { return "situation_snapshots" }
```

- [ ] **Step 2: 创建 server/dto/situation.go**

```go
package dto

import "time"

// === 态势概览 ===

type SituationOverviewResponse struct {
	ActiveChains     int     `json:"active_chains"`
	TotalChains24h   int     `json:"total_chains_24h"`
	TotalAttackers24h int    `json:"total_attackers_24h"`
	OverallRiskScore float64 `json:"overall_risk_score"`
	RiskTrend        string  `json:"risk_trend"` // rising / stable / falling
	TopAttackTypes   []CountItem `json:"top_attack_types"`
	TopAttackerIPs   []CountItem `json:"top_attacker_ips"`
	TopTargetSites   []CountItem `json:"top_target_sites"`
	ByCountry        []CountItem `json:"by_country"`
}

type CountItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// === 攻击链 ===

type ChainListRequest struct {
	SourceIP  string `form:"source_ip"`
	Stage     string `form:"stage"`
	Severity  string `form:"severity"`
	Active    *bool  `form:"active"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

type ChainListResponse struct {
	Chains     []ChainSummary `json:"chains"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

type ChainSummary struct {
	ID         string    `json:"id"`
	SourceIP   string    `json:"source_ip"`
	GeoCountry string    `json:"geo_country"`
	Stages     []string  `json:"stages"` // 阶段名列表
	RiskScore  int       `json:"risk_score"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	Active     bool      `json:"active"`
}

type ChainDetailResponse struct {
	ID             string                   `json:"id"`
	SourceIP       string                   `json:"source_ip"`
	GeoCountry     string                   `json:"geo_country"`
	Stages         []ChainStageItem         `json:"stages"`
	CorrelationIDs []string                 `json:"correlation_ids"`
	RiskScore      int                      `json:"risk_score"`
	RiskLabel      string                   `json:"risk_label"`
	FirstSeen      time.Time                `json:"first_seen"`
	LastSeen       time.Time                `json:"last_seen"`
	Active         bool                     `json:"active"`
	AttackerProfile *AttackerProfileDetail  `json:"attacker_profile,omitempty"`
}

type ChainStageItem struct {
	Stage      string    `json:"stage"`
	Technique  string    `json:"technique"`
	DetectedAt time.Time `json:"detected_at"`
	Confidence float64   `json:"confidence"`
	Evidence   []string  `json:"evidence"`
}

// === 攻击者 ===

type AttackerListRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	SortBy     string `form:"sort_by"` // risk_score, total_attacks, last_seen
	RiskLabel  string `form:"risk_label"`
}

type AttackerListResponse struct {
	Attackers []AttackerProfileSummary `json:"attackers"`
	Total     int64                    `json:"total"`
}

type AttackerProfileSummary struct {
	SourceIP          string    `json:"source_ip"`
	GeoCountry        string    `json:"geo_country"`
	TotalAttacks      int       `json:"total_attacks"`
	UniqueAttackTypes  int      `json:"unique_attack_types"`
	TopAttackType     string    `json:"top_attack_type"`
	AttackPhase       string    `json:"attack_phase"`
	RiskScore         int       `json:"risk_score"`
	RiskLabel         string    `json:"risk_label"`
	LastSeen          time.Time `json:"last_seen"`
}

type AttackerProfileDetail struct {
	SourceIP          string    `json:"source_ip"`
	GeoCountry        string    `json:"geo_country"`
	GeoCity           string    `json:"geo_city,omitempty"`
	TotalAttacks      int       `json:"total_attacks"`
	UniqueAttackTypes  int      `json:"unique_attack_types"`
	TopAttackType     string    `json:"top_attack_type"`
	UniqueTargetSites  int      `json:"unique_target_sites"`
	ActiveHours       []int     `json:"active_hours"`
	BurstIntervals    []string  `json:"burst_intervals"`
	AttackPhase       string    `json:"attack_phase"`
	ToolsIdentified   string    `json:"tools_identified"`
	IsAutomated       bool      `json:"is_automated"`
	IsPersistent      bool      `json:"is_persistent"`
	RiskScore         int       `json:"risk_score"`
	RiskLabel         string    `json:"risk_label"`
	FirstSeen         time.Time `json:"first_seen"`
	LastSeen          time.Time `json:"last_seen"`
	RecentEvents      []LogEventItem `json:"recent_events"`
}

type LogEventItem struct {
	ID            string    `json:"id"`
	AttackType    string    `json:"attack_type"`
	Severity      string    `json:"severity"`
	Action        string    `json:"action"`
	SiteDomain    string    `json:"site_domain"`
	CorrelationID string    `json:"correlation_id"`
	Timestamp     time.Time `json:"timestamp"`
}

// === 趋势 ===

type TrendRequest struct {
	Duration string `form:"duration"` // 1h / 6h / 24h / 7d
}

type TrendResponse struct {
	Timeline       []TrendPoint `json:"timeline"`
	FrequentTypes  []CountItem  `json:"frequent_types"`
	ActiveAttackers int         `json:"active_attackers"`
	NewChains24h    int         `json:"new_chains_24h"`
}

type TrendPoint struct {
	Timestamp    int64 `json:"timestamp"`
	TotalEvents  int   `json:"total_events"`
	BlockedCount int   `json:"blocked_count"`
	DetectCount  int   `json:"detect_count"`
	UniqueIPs    int   `json:"unique_ips"`
}

// === 规则管理 ===

type SituationRuleRequest struct {
	Name           string `json:"name" binding:"required"`
	Stage          string `json:"stage" binding:"required"`
	LogQL          string `json:"logql" binding:"required"`
	Interval       int    `json:"interval_seconds" binding:"required"`
	Threshold      int    `json:"threshold" binding:"required"`
	Severity       string `json:"severity" binding:"required"`
	MITRETactic    string `json:"mitre_tactic"`
	MITRETechnique string `json:"mitre_technique"`
	Enabled        bool   `json:"enabled"`
}

// === 快速处置 ===

type QuickActionRequest struct {
	SourceIP      string `json:"source_ip" binding:"required"`
	Action        string `json:"action" binding:"required"` // block / blacklist / both
	DurationHours int    `json:"duration_hours"`            // 0 = 永久
	Reason        string `json:"reason" binding:"required"`
	CorrelationID string `json:"correlation_id"`
}

type QuickActionResponse struct {
	Success     bool   `json:"success"`
	BlockRecordID string `json:"block_record_id,omitempty"`
	AuditLogID    string `json:"audit_log_id,omitempty"`
	Note          string `json:"note"`
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/model/situation.go server/dto/situation.go
git commit -m "feat(situation): add situation awareness data models and DTOs"
```

---

### Task P1-2: 态势检测规则引擎

**Files:**
- Create: `server/service/situation/logql_builder.go`
- Create: `server/service/situation/rule.go`
- Create: `server/repository/situation.go`

**Interfaces:**
- Consumes: `model.SituationRule` (from P1-1), `service.LokiLogService` (已有)
- Produces:
  - `situation.LogQLBuilder` — 动态 LogQL 构造器
  - `situation.RuleEngine` — 规则加载+评估引擎
  - `repository.SituationRepository` — CRUD 接口

**Steps:**

- [ ] **Step 1: 创建 server/service/situation/logql_builder.go**

```go
package situation

import (
	"fmt"
	"strings"
	"time"
)

// LogQLBuilder 动态构造 LogQL 查询
type LogQLBuilder struct {
	selector string
	filters  []string
	groupBy  string
	aggregate string
}

// NewLogQLBuilder 创建新的 LogQL 构造器
func NewLogQLBuilder() *LogQLBuilder {
	return &LogQLBuilder{
		selector: `{container_name="mrya-waf"}`,
		filters:  make([]string, 0),
	}
}

// Filter 添加标签过滤器
func (b *LogQLBuilder) Filter(key, operator, value string) *LogQLBuilder {
	b.filters = append(b.filters, fmt.Sprintf(`%s%s"%s"`, key, operator, value))
	return b
}

// FilterIP 按源 IP 过滤
func (b *LogQLBuilder) FilterIP(ip string) *LogQLBuilder {
	return b.Filter("source_ip", `=`, ip)
}

// FilterAttackType 按攻击类型过滤
func (b *LogQLBuilder) FilterAttackType(attackType string) *LogQLBuilder {
	return b.Filter("attack_type", `=`, attackType)
}

// FilterSeverity 按严重级别过滤
func (b *LogQLBuilder) FilterSeverity(severity string) *LogQLBuilder {
	return b.Filter("severity", `=~`, severity)
}

// CountOverTime 按时间窗口计数
func (b *LogQLBuilder) CountOverTime(duration string) string {
	sel := b.buildSelector()
	return fmt.Sprintf(`sum by (source_ip) (count_over_time(%s[%s]))`, sel, duration)
}

// DistinctCount 去重计数
func (b *LogQLBuilder) DistinctCount(field string, duration string) string {
	sel := b.buildSelector()
	return fmt.Sprintf(`count by (source_ip) (count_over_time(%s | unwrap %s [%s]))`, sel, field, duration)
}

// RawQuery 返回原始 LogQL 查询字符串
func (b *LogQLBuilder) RawQuery() string {
	return b.buildSelector()
}

// AttackChainQuery 构建单个 IP 的攻击链查询
func (b *LogQLBuilder) AttackChainQuery(ip string, duration time.Duration) string {
	end := time.Now().Unix()
	start := time.Now().Add(-duration).Unix()
	return fmt.Sprintf(
		`{container_name="mrya-waf",source_ip="%s"} | json | line_format "{{.attack_type}} {{.severity}} {{.action}} {{.waf_phase}} {{.site_id}}"`,
		ip,
	)
}

// StageCountQuery 构建攻击阶段分布查询
func (b *LogQLBuilder) StageCountQuery(ip string, duration time.Duration) string {
	end := time.Now().Unix()
	start := time.Now().Add(-duration).Unix()
	filters := b.appendFilter(`source_ip="%s"`, ip)
	_ = start // 用于构造完整查询
	_ = end
	return fmt.Sprintf(
		`sum by (attack_type) (count_over_time({%s}[%dm]))`,
		strings.Join(filters, ","),
		int(duration.Minutes()),
	)
}

func (b *LogQLBuilder) buildSelector() string {
	filters := make([]string, 0, len(b.filters)+1)
	filters = append(filters, `container_name="mrya-waf"`)
	for _, f := range b.filters {
		filters = append(filters, f)
	}
	return "{" + strings.Join(filters, ",") + "}"
}

func (b *LogQLBuilder) appendFilter(keyValue string, args ...interface{}) []string {
	filters := make([]string, 0, len(b.filters)+1)
	filters = append(filters, fmt.Sprintf(keyValue, args...))
	for _, f := range b.filters {
		filters = append(filters, f)
	}
	return filters
}
```

- [ ] **Step 2: 创建 server/service/situation/rule.go**

```go
package situation

import (
	"context"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/rs/zerolog"
)

// DefaultRules 内置默认检测规则
var DefaultRules = []model.SituationRule{
	{
		Name:           "高频端口扫描检测",
		Stage:          string(model.StageScanning),
		LogQL:          `sum by (source_ip)(count_over_time({container_name="mrya-waf",attack_type="scanner"}[5m])) > 100`,
		Interval:       30,
		Threshold:      100,
		Severity:       "medium",
		MITRETactic:    "TA0043",
		MITRETechnique: "T1595",
		Enabled:        true,
	},
	{
		Name:           "SQL注入高频利用",
		Stage:          string(model.StageExploitation),
		LogQL:          `sum by (source_ip)(count_over_time({container_name="mrya-waf",attack_type="sql_injection",severity=~"critical|high"}[5m])) > 10`,
		Interval:       30,
		Threshold:      10,
		Severity:       "critical",
		MITRETactic:    "TA0001",
		MITRETechnique: "T1190",
		Enabled:        true,
	},
	{
		Name:           "多攻击面探测",
		Stage:          string(model.StageScanning),
		LogQL:          `count by (source_ip)(count_over_time({container_name="mrya-waf",attack_type!=""}[10m])) > 50`,
		Interval:       30,
		Threshold:      3,
		Severity:       "high",
		MITRETactic:    "TA0043",
		MITRETechnique: "T1046",
		Enabled:        true,
	},
	{
		Name:           "漏洞扫描器指纹识别",
		Stage:          string(model.StageScanning),
		LogQL:          `{container_name="mrya-waf",attack_type="vulnerability_scanner"} |~ "(?i)(nuclei|nikto|nessus|openvas|acunetix|burpsuite)"`,
		Interval:       30,
		Threshold:      1,
		Severity:       "high",
		MITRETactic:    "TA0043",
		MITRETechnique: "T1595",
		Enabled:        true,
	},
	{
		Name:           "RCE攻击检测",
		Stage:          string(model.StageExploitation),
		LogQL:          `sum by (source_ip)(count_over_time({container_name="mrya-waf",attack_type=~"rce|command_injection"}[5m])) > 5`,
		Interval:       30,
		Threshold:      5,
		Severity:       "critical",
		MITRETactic:    "TA0002",
		MITRETechnique: "T1059",
		Enabled:        true,
	},
}

// RuleEngine LogQL 规则评估引擎
type RuleEngine struct {
	loki   service.LokiLogService
	repo   repository.SituationRepository
	logger zerolog.Logger
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(loki service.LokiLogService, repo repository.SituationRepository) *RuleEngine {
	return &RuleEngine{
		loki:   loki,
		repo:   repo,
		logger: config.GetServiceLogger("situation-rule"),
	}
}

// InitializeDefaults 写入默认规则（已存在的跳过，按 name 去重）
func (e *RuleEngine) InitializeDefaults(ctx context.Context) error {
	for _, rule := range DefaultRules {
		existing, err := e.repo.FindRuleByName(ctx, rule.Name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue // 已存在，跳过
		}
		if err := e.repo.CreateRule(ctx, &rule); err != nil {
			return err
		}
	}
	return nil
}

// Evaluate 执行单条规则评估
func (e *RuleEngine) Evaluate(ctx context.Context, rule model.SituationRule) (*RuleResult, error) {
	resp, err := e.loki.QueryLogs(ctx, service.LokiQueryRequest{
		Query: rule.LogQL,
		Limit: 100,
		Start: fmtDuration(time.Duration(rule.Interval) * time.Second),
	})
	if err != nil {
		e.logger.Warn().Err(err).Str("rule_id", rule.ID).Msg("LogQL rule evaluation failed")
		return nil, err
	}

	result := &RuleResult{
		Rule:      rule,
		HitCount:  len(resp.Data.Result),
		HitIPs:    extractSourceIPs(resp),
		Timestamp: time.Now(),
	}
	result.Triggered = result.HitCount > rule.Threshold
	return result, nil
}

// EvaluateAll 评估所有启用的规则
func (e *RuleEngine) EvaluateAll(ctx context.Context) ([]RuleResult, error) {
	rules, err := e.repo.ListEnabledRules(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]RuleResult, 0, len(rules))
	for _, rule := range rules {
		result, err := e.Evaluate(ctx, rule)
		if err != nil {
			continue // 单条规则失败不影响其他规则
		}
		if result.Triggered {
			results = append(results, *result)
		}
	}
	return results, nil
}

// RuleResult 单条规则评估结果
type RuleResult struct {
	Rule      model.SituationRule `json:"rule"`
	HitCount  int                 `json:"hit_count"`
	HitIPs    []string            `json:"hit_ips"`
	Triggered bool                `json:"triggered"`
	Timestamp time.Time           `json:"timestamp"`
}

func extractSourceIPs(resp *service.LokiQueryResponse) []string {
	seen := make(map[string]bool)
	var ips []string
	for _, stream := range resp.Data.Result {
		if ip, ok := stream.Stream["source_ip"]; ok && ip != "" && !seen[ip] {
			seen[ip] = true
			ips = append(ips, ip)
		}
	}
	return ips
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	s := d.String()
	if s == "0s" {
		return "1m"
	}
	return s
}
```

- [ ] **Step 3: 创建 server/repository/situation.go**

```go
package repository

import (
	"context"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SituationRepository interface {
	// 规则管理
	ListRules(ctx context.Context) ([]model.SituationRule, error)
	ListEnabledRules(ctx context.Context) ([]model.SituationRule, error)
	FindRuleByName(ctx context.Context, name string) (*model.SituationRule, error)
	GetRuleByID(ctx context.Context, id string) (*model.SituationRule, error)
	CreateRule(ctx context.Context, rule *model.SituationRule) error
	UpdateRule(ctx context.Context, id string, rule *model.SituationRule) error
	DeleteRule(ctx context.Context, id string) error

	// 攻击链
	ListChains(ctx context.Context, filter bson.M, skip, limit int64) ([]model.AttackChain, int64, error)
	GetChainByID(ctx context.Context, id string) (*model.AttackChain, error)
	UpsertChain(ctx context.Context, chain *model.AttackChain) error
	GetChainByIP(ctx context.Context, ip string) (*model.AttackChain, error)

	// 攻击者画像
	UpsertProfile(ctx context.Context, profile *model.AttackerProfile) error
	GetProfile(ctx context.Context, ip string) (*model.AttackerProfile, error)
	ListProfiles(ctx context.Context, sortBy string, skip, limit int64) ([]model.AttackerProfile, int64, error)
}

type MongoSituationRepository struct {
	ruleCollection     *mongo.Collection
	chainCollection    *mongo.Collection
	profileCollection  *mongo.Collection
	logger             zerolog.Logger
}

func NewSituationRepository(db *mongo.Database) SituationRepository {
	return &MongoSituationRepository{
		ruleCollection:    db.Collection("situation_rules"),
		chainCollection:   db.Collection("attack_chains"),
		profileCollection: db.Collection("attacker_profiles"),
		logger:            config.GetRepositoryLogger("situation"),
	}
}

func (r *MongoSituationRepository) ListRules(ctx context.Context) ([]model.SituationRule, error) {
	cursor, err := r.ruleCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rules []model.SituationRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *MongoSituationRepository) ListEnabledRules(ctx context.Context) ([]model.SituationRule, error) {
	cursor, err := r.ruleCollection.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rules []model.SituationRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *MongoSituationRepository) FindRuleByName(ctx context.Context, name string) (*model.SituationRule, error) {
	var rule model.SituationRule
	err := r.ruleCollection.FindOne(ctx, bson.M{"name": name}).Decode(&rule)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *MongoSituationRepository) GetRuleByID(ctx context.Context, id string) (*model.SituationRule, error) {
	var rule model.SituationRule
	err := r.ruleCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *MongoSituationRepository) CreateRule(ctx context.Context, rule *model.SituationRule) error {
	_, err := r.ruleCollection.InsertOne(ctx, rule)
	return err
}

func (r *MongoSituationRepository) UpdateRule(ctx context.Context, id string, rule *model.SituationRule) error {
	_, err := r.ruleCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": rule})
	return err
}

func (r *MongoSituationRepository) DeleteRule(ctx context.Context, id string) error {
	_, err := r.ruleCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MongoSituationRepository) ListChains(ctx context.Context, filter bson.M, skip, limit int64) ([]model.AttackChain, int64, error) {
	total, err := r.chainCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"last_seen": -1})
	cursor, err := r.chainCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var chains []model.AttackChain
	if err := cursor.All(ctx, &chains); err != nil {
		return nil, 0, err
	}
	return chains, total, nil
}

func (r *MongoSituationRepository) GetChainByID(ctx context.Context, id string) (*model.AttackChain, error) {
	var chain model.AttackChain
	err := r.chainCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&chain)
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

func (r *MongoSituationRepository) UpsertChain(ctx context.Context, chain *model.AttackChain) error {
	filter := bson.M{"source_ip": chain.SourceIP}
	update := bson.M{"$set": chain}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.chainCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoSituationRepository) GetChainByIP(ctx context.Context, ip string) (*model.AttackChain, error) {
	var chain model.AttackChain
	err := r.chainCollection.FindOne(ctx, bson.M{"source_ip": ip}).Decode(&chain)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

func (r *MongoSituationRepository) UpsertProfile(ctx context.Context, profile *model.AttackerProfile) error {
	filter := bson.M{"source_ip": profile.SourceIP}
	update := bson.M{"$set": profile}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.profileCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoSituationRepository) GetProfile(ctx context.Context, ip string) (*model.AttackerProfile, error) {
	var profile model.AttackerProfile
	err := r.profileCollection.FindOne(ctx, bson.M{"source_ip": ip}).Decode(&profile)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *MongoSituationRepository) ListProfiles(ctx context.Context, sortBy string, skip, limit int64) ([]model.AttackerProfile, int64, error) {
	total, err := r.profileCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	sortField := "risk_score"
	if sortBy != "" {
		sortField = sortBy
	}
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{sortField: -1})
	cursor, err := r.profileCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var profiles []model.AttackerProfile
	if err := cursor.All(ctx, &profiles); err != nil {
		return nil, 0, err
	}
	return profiles, total, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add server/service/situation/ server/repository/situation.go
git commit -m "feat(situation): add LogQL builder, rule engine, and situation repository"
```

---

### Task P1-3: 态势感知引擎（攻击链 + 画像 + 风险评分）

**Files:**
- Create: `server/service/situation/engine.go`
- Create: `server/service/situation/attack_chain.go`
- Create: `server/service/situation/profiler.go`
- Create: `server/service/situation/risk.go`
- Create: `server/service/situation/publisher.go`

**Interfaces:**
- Consumes: All types from P1-1 and P1-2, `LokiLogService`, Redis client
- Produces:
  - `situation.Engine` — 态势聚合引擎
  - `situation.AttackChainBuilder` — 攻击链构建器
  - `situation.Profiler` — 攻击者画像
  - `situation.RiskScorer` — 风险评分
  - `situation.Publisher` — Redis Pub/Sub 发布

**Steps:**

- [ ] **Step 1: 创建 server/service/situation/attack_chain.go**

```go
package situation

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/rs/zerolog"
)

type AttackChainBuilder struct {
	loki service.LokiLogService
	repo repository.SituationRepository
	logger zerolog.Logger
}

func NewAttackChainBuilder(loki service.LokiLogService, repo repository.SituationRepository) *AttackChainBuilder {
	return &AttackChainBuilder{
		loki:   loki,
		repo:   repo,
		logger: config.GetServiceLogger("attack-chain"),
	}
}

// BuildChain 构建指定 IP 的完整攻击链
func (b *AttackChainBuilder) BuildChain(ctx context.Context, ip string) (*model.AttackChain, error) {
	// 1. 查 Loki 获取该 IP 最近 24h 的日志
	resp, err := b.loki.QueryRange(ctx, service.LokiRangeRequest{
		Query: fmt.Sprintf(`{container_name="mrya-waf",source_ip="%s"} | json`, ip),
		Start: fmt.Sprintf("%d", time.Now().Add(-24*time.Hour).Unix()),
		End:   fmt.Sprintf("%d", time.Now().Unix()),
		Step:  "1m",
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}

	entries := service.ToLogEntries(resp)
	if entries.TotalHits == 0 {
		return nil, nil
	}

	// 2. 构建攻击阶段列表
	stages := b.buildStages(entries)

	// 3. 提取 correlation_ids
	correlationIDs := extractCorrelationIDs(entries)

	// 4. 计算阶段中的最高置信度
	var maxConfidence float64
	for _, s := range stages {
		if s.Confidence > maxConfidence {
			maxConfidence = s.Confidence
		}
	}

	chain := &model.AttackChain{
		SourceIP:       ip,
		GeoCountry:     extractFirstGeo(entries),
		Stages:         stages,
		CorrelationIDs: correlationIDs,
		FirstSeen:      parseFirstTimestamp(entries),
		LastSeen:       parseLastTimestamp(entries),
		Active:         true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return chain, nil
}

// UpdateChainFromRuleResult 根据规则命中更新攻击链
func (b *AttackChainBuilder) UpdateChainFromRuleResult(ctx context.Context, result RuleResult) error {
	rule := result.Rule
	stage := model.AttackStage(rule.Stage)
	if _, ok := model.StageMapping[rule.Stage]; ok {
		stage = model.StageMapping[rule.Stage]
	}

	for _, ip := range result.HitIPs {
		existing, _ := b.repo.GetChainByIP(ctx, ip)

		if existing == nil {
			// 创建新攻击链
			chain := &model.AttackChain{
				SourceIP:   ip,
				Stages: []model.ChainStage{
					{
						Stage:      stage,
						Technique:  rule.MITRETechnique,
						DetectedAt: result.Timestamp,
						Confidence: float64(result.HitCount) / float64(rule.Threshold+1),
					},
				},
				FirstSeen: result.Timestamp,
				LastSeen:  result.Timestamp,
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := b.repo.UpsertChain(ctx, chain); err != nil {
				b.logger.Error().Err(err).Str("ip", ip).Msg("Failed to create attack chain")
			}
		} else {
			// 更新已有攻击链：检查阶段迁移
			newStage := b.shouldAdvanceStage(existing, stage, result)
			existing.Stages = append(existing.Stages, ChainStage{
				Stage:      newStage,
				Technique:  rule.MITRETechnique,
				DetectedAt: result.Timestamp,
				Confidence: float64(result.HitCount) / float64(rule.Threshold+1),
			})
			existing.LastSeen = result.Timestamp
			existing.UpdatedAt = time.Now()

			if err := b.repo.UpsertChain(ctx, existing); err != nil {
				b.logger.Error().Err(err).Str("ip", ip).Msg("Failed to update attack chain")
			}
		}
	}
	return nil
}

func (b *AttackChainBuilder) shouldAdvanceStage(existing *model.AttackChain, detectedStage model.AttackStage, result RuleResult) model.AttackStage {
	if detectedStage.Order() > b.highestStage(existing).Order() {
		return detectedStage
	}
	// 即使没有推进阶段，也记录当前阶段
	return b.highestStage(existing)
}

func (b *AttackChainBuilder) highestStage(chain *model.AttackChain) model.AttackStage {
	var highest model.AttackStage
	for _, s := range chain.Stages {
		if s.Stage.Order() > highest.Order() {
			highest = s.Stage
		}
	}
	return highest
}

func (b *AttackChainBuilder) buildStages(entries *service.LokiLogQueryResponse) []model.ChainStage {
	stageMap := make(map[model.AttackStage]*model.ChainStage)
	for _, entry := range entries.Results {
		attackType := entry.Labels["attack_type"]
		stage, ok := model.StageMapping[attackType]
		if !ok {
			stage = model.StageUnknown
		}
		if existing, ok := stageMap[stage]; ok {
			existing.Confidence += 0.1
		} else {
			stageMap[stage] = &model.ChainStage{
				Stage:      stage,
				Confidence: 0.5,
			}
		}
	}

	stages := make([]model.ChainStage, 0, len(stageMap))
	for _, s := range stageMap {
		if s.Confidence > 1.0 {
			s.Confidence = 1.0
		}
		stages = append(stages, *s)
	}
	return stages
}

func extractCorrelationIDs(entries *service.LokiLogQueryResponse) []string {
	seen := make(map[string]bool)
	var ids []string
	for _, e := range entries.Results {
		if cid, ok := e.Labels["correlation_id"]; ok && cid != "" && !seen[cid] {
			seen[cid] = true
			ids = append(ids, cid)
		}
	}
	return ids
}

func extractFirstGeo(entries *service.LokiLogQueryResponse) string {
	for _, e := range entries.Results {
		if geo, ok := e.Labels["geo_country"]; ok && geo != "" {
			return geo
		}
	}
	return "unknown"
}

func parseFirstTimestamp(entries *service.LokiLogQueryResponse) time.Time {
	if len(entries.Results) > 0 {
		return time.Now().Add(-24 * time.Hour) // 简化：精确解析后续可优化
	}
	return time.Now()
}

func parseLastTimestamp(entries *service.LokiLogQueryResponse) time.Time {
	if len(entries.Results) > 0 {
		return time.Now()
	}
	return time.Now()
}
```

- [ ] **Step 2: 创建 server/service/situation/profiler.go**

```go
package situation

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/rs/zerolog"
)

type Profiler struct {
	loki   service.LokiLogService
	repo   repository.SituationRepository
	logger zerolog.Logger
}

func NewProfiler(loki service.LokiLogService, repo repository.SituationRepository) *Profiler {
	return &Profiler{
		loki:   loki,
		repo:   repo,
		logger: config.GetServiceLogger("profiler"),
	}
}

// BuildProfile 构建攻击者画像
func (p *Profiler) BuildProfile(ctx context.Context, ip string) (*model.AttackerProfile, error) {
	now := time.Now()
	start := now.Add(-24 * time.Hour)

	// 查询该 IP 24h 的攻击日志
	resp, err := p.loki.QueryRange(ctx, service.LokiRangeRequest{
		Query: fmt.Sprintf(`{container_name="mrya-waf",source_ip="%s"} | json`, ip),
		Start: fmt.Sprintf("%d", start.Unix()),
		End:   fmt.Sprintf("%d", now.Unix()),
		Step:  "5m",
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}

	entries := service.ToLogEntries(resp)

	profile := &model.AttackerProfile{
		SourceIP:         ip,
		FirstSeen:        p.findFirstTime(entries),
		LastSeen:         p.findLastTime(entries),
		UpdatedAt:        now,
	}

	// 攻击类型统计
	typeCount := make(map[string]int)
	siteCount := make(map[string]bool)
	hourCount := make([]int, 24)
	for _, e := range entries.Results {
		if at, ok := e.Labels["attack_type"]; ok && at != "" {
			typeCount[at]++
		}
		if site, ok := e.Labels["site_id"]; ok && site != "" {
			siteCount[site] = true
		}
		// 提取活跃时段
		if ts, err := strconv.ParseInt(e.Timestamp, 10, 64); err == nil {
			h := time.Unix(0, ts).Hour()
			if h >= 0 && h < 24 {
				hourCount[h]++
			}
		}
	}
	profile.TotalAttacks = entries.TotalHits
	profile.UniqueAttackTypes = len(typeCount)
	profile.UniqueTargetSites = len(siteCount)
	profile.GeoCountry = p.extractFirstLabel(entries, "geo_country")

	// Top attack type
	maxCount := 0
	for at, c := range typeCount {
		if c > maxCount {
			maxCount = c
			profile.TopAttackType = at
		}
	}

	// 活跃时段
	profile.ActiveHours = make([]int, 0)
	for h, c := range hourCount {
		if c > 0 {
			profile.ActiveHours = append(profile.ActiveHours, h)
		}
	}

	// 行为特征
	profile.IsAutomated = profile.TotalAttacks > 50 && profile.UniqueAttackTypes > 2
	profile.IsPersistent = profile.LastSeen.Sub(profile.FirstSeen) > 24*time.Hour

	// 攻击阶段
	chain, _ := p.repo.GetChainByIP(ctx, ip)
	if chain != nil {
		var highest model.AttackStage
		for _, s := range chain.Stages {
			if s.Stage.Order() > highest.Order() {
				highest = s.Stage
			}
		}
		profile.AttackPhase = string(highest)
	} else {
		profile.AttackPhase = string(model.StageUnknown)
	}

	// 工具识别
	profile.ToolsIdentified = p.detectTools(entries)

	return profile, nil
}

// SaveProfile 构建并持久化画像
func (p *Profiler) SaveProfile(ctx context.Context, ip string) error {
	profile, err := p.BuildProfile(ctx, ip)
	if err != nil {
		return err
	}
	profile.UpdatedAt = time.Now()
	return p.repo.UpsertProfile(ctx, profile)
}

func (p *Profiler) findFirstTime(entries *service.LokiLogQueryResponse) time.Time {
	if len(entries.Results) > 0 {
		ts := entries.Results[len(entries.Results)-1].Timestamp
		if t, err := strconv.ParseInt(ts, 10, 64); err == nil {
			return time.Unix(0, t)
		}
	}
	return time.Now().Add(-24 * time.Hour)
}

func (p *Profiler) findLastTime(entries *service.LokiLogQueryResponse) time.Time {
	if len(entries.Results) > 0 {
		ts := entries.Results[0].Timestamp
		if t, err := strconv.ParseInt(ts, 10, 64); err == nil {
			return time.Unix(0, t)
		}
	}
	return time.Now()
}

func (p *Profiler) extractFirstLabel(entries *service.LokiLogQueryResponse, key string) string {
	for _, e := range entries.Results {
		if v, ok := e.Labels[key]; ok && v != "" {
			return v
		}
	}
	return ""
}

func (p *Profiler) detectTools(entries *service.LokiLogQueryResponse) string {
	tools := []string{"nuclei", "nikto", "nessus", "openvas", "acunetix", "burpsuite", "sqlmap", "nmap"}
	found := make([]string, 0)
	for _, e := range entries.Results {
		msg := strings.ToLower(e.Message)
		for _, tool := range tools {
			if strings.Contains(msg, tool) {
				found = append(found, tool)
			}
		}
	}
	if len(found) > 0 {
		return strings.Join(uniqueStrings(found), ", ")
	}
	return ""
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
```

- [ ] **Step 3: 创建 server/service/situation/risk.go**

```go
package situation

import (
	"github.com/mingrenya/AI-Waf/pkg/model"
)

var attackStageWeight = map[string]int{
	string(model.StageReconnaissance):   5,
	string(model.StageScanning):        10,
	string(model.StageExploitation):    20,
	string(model.StageLateralMovement): 30,
	string(model.StageC2):              30,
	string(model.StageExfiltration):    40,
}

// CalculateRisk 计算风险评分 0-100
func CalculateRisk(profile *model.AttackerProfile) int {
	score := 0

	// 攻击频率 (0-30)
	freqScore := profile.TotalAttacks * 3
	if freqScore > 30 {
		freqScore = 30
	}
	score += freqScore

	// 攻击多样性 (0-20)
	divScore := profile.UniqueAttackTypes * 5
	if divScore > 20 {
		divScore = 20
	}
	score += divScore

	// 攻击阶段 (0-20)
	if w, ok := attackStageWeight[profile.AttackPhase]; ok {
		score += w
	}

	// 持续性 (0-15)
	if profile.IsPersistent {
		score += 15
	}

	// 自动化 (0-15)
	if profile.IsAutomated {
		score += 15
	}

	if score > 100 {
		score = 100
	}
	return score
}

// RiskLabel 根据评分返回风险标签
func RiskLabel(score int) string {
	switch {
	case score >= 80:
		return "critical"
	case score >= 60:
		return "high"
	case score >= 30:
		return "medium"
	default:
		return "low"
	}
}
```

- [ ] **Step 4: 创建 server/service/situation/publisher.go**

```go
package situation

import (
	"context"
	"encoding/json"

	"github.com/mingrenya/AI-Waf/server/config"
	ws "github.com/mingrenya/AI-Waf/server/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Publisher struct {
	redis  *redis.Client
	logger zerolog.Logger
}

func NewPublisher(redis *redis.Client) *Publisher {
	return &Publisher{
		redis:  redis,
		logger: config.GetServiceLogger("situation-publisher"),
	}
}

// PublishAlert 发布攻击链告警
func (p *Publisher) PublishAlert(chain interface{}) {
	data, _ := json.Marshal(chain)
	p.publishRedis(context.Background(), "situation:alert", data)
	ws.GetHub().BroadcastJSON("situation:alert", chain)
}

// PublishUpdate 发布态势数据更新
func (p *Publisher) PublishUpdate(update interface{}) {
	data, _ := json.Marshal(update)
	p.publishRedis(context.Background(), "situation:update", data)
	ws.GetHub().BroadcastJSON("situation:update", update)
}

// PublishAttack 发布实时攻击事件
func (p *Publisher) PublishAttack(event interface{}) {
	data, _ := json.Marshal(event)
	p.publishRedis(context.Background(), "situation:attack", data)
	ws.GetHub().BroadcastJSON("situation:attack", event)
}

func (p *Publisher) publishRedis(ctx context.Context, channel string, data []byte) {
	if p.redis == nil {
		return
	}
	if err := p.redis.Publish(ctx, channel, string(data)).Err(); err != nil {
		p.logger.Warn().Err(err).Str("channel", channel).Msg("Redis publish failed")
	}
}
```

- [ ] **Step 5: 创建 server/service/situation/engine.go**

```go
package situation

import (
	"context"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
)

// Engine 态势感知聚合引擎
type Engine struct {
	ruleEngine    *RuleEngine
	chainBuilder  *AttackChainBuilder
	profiler      *Profiler
	publisher     *Publisher
	logger        zerolog.Logger
	running       bool
	stopCh        chan struct{}
}

// NewEngine 创建态势感知引擎
func NewEngine(
	ruleEngine *RuleEngine,
	chainBuilder *AttackChainBuilder,
	profiler *Profiler,
	publisher *Publisher,
) *Engine {
	return &Engine{
		ruleEngine:   ruleEngine,
		chainBuilder: chainBuilder,
		profiler:     profiler,
		publisher:    publisher,
		logger:       config.GetServiceLogger("situation-engine"),
		stopCh:       make(chan struct{}),
	}
}

// Start 启动引擎（后台 goroutine）
func (e *Engine) Start(ctx context.Context, interval time.Duration) {
	if e.running {
		return
	}
	e.running = true
	e.logger.Info().Dur("interval", interval).Msg("Situation engine started")

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-e.stopCh:
				e.logger.Info().Msg("Situation engine stopped")
				return
			case <-ticker.C:
				e.runCycle(ctx)
			}
		}
	}()
}

// Stop 停止引擎
func (e *Engine) Stop() {
	if e.running {
		e.stopCh <- struct{}{}
		e.running = false
	}
}

// runCycle 执行一轮检测
func (e *Engine) runCycle(ctx context.Context) {
	// 1. 执行所有规则
	results, err := e.ruleEngine.EvaluateAll(ctx)
	if err != nil {
		e.logger.Error().Err(err).Msg("Rule evaluation failed")
		return
	}

	if len(results) == 0 {
		return
	}

	e.logger.Info().Int("triggered_rules", len(results)).Msg("Rules triggered")

	// 2. 对每个命中结果，构建/更新攻击链
	for _, result := range results {
		if err := e.chainBuilder.UpdateChainFromRuleResult(ctx, result); err != nil {
			e.logger.Warn().Err(err).Str("rule", result.Rule.Name).Msg("Chain update failed")
		}
	}

	// 3. 对命中 IP 构建/刷新画像 + 计算风险
	processedIPs := make(map[string]bool)
	for _, result := range results {
		for _, ip := range result.HitIPs {
			if processedIPs[ip] {
				continue
			}
			processedIPs[ip] = true

			if err := e.profiler.SaveProfile(ctx, ip); err != nil {
				e.logger.Warn().Err(err).Str("ip", ip).Msg("Profile build failed")
				continue
			}

			profile, err := e.profiler.BuildProfile(ctx, ip)
			if err != nil {
				continue
			}

			// 计算风险评分
			score := CalculateRisk(profile)
			profile.RiskScore = score
			profile.RiskLabel = RiskLabel(score)

			if score >= 60 {
				e.publisher.PublishAlert(map[string]interface{}{
					"ip":        ip,
					"score":     score,
					"label":     profile.RiskLabel,
					"phase":     profile.AttackPhase,
					"country":   profile.GeoCountry,
					"attack_type": profile.TopAttackType,
				})
			}

			e.publisher.PublishUpdate(map[string]interface{}{
				"profile": profile,
			})
		}
	}
}
```

- [ ] **Step 6: Commit**

```bash
git add server/service/situation/
git commit -m "feat(situation): add engine, attack chain builder, profiler, risk scorer, publisher"
```

---

### Task P1-4: 态势感知控制器 + 路由 + 启动

**Files:**
- Create: `server/controller/situation.go`
- Modify: `server/router/router.go`
- Modify: `server/main.go`

**Interfaces:**
- Consumes: All situation engine, LokiLogService
- Produces: REST API endpoints (10 routes) + Engine startup in main.go

**Steps:**

- [ ] **Step 1: 创建 server/controller/situation.go**

文件内容包含所有接口实现，因篇幅限制此处简述结构。完整代码遵循 controller 模式：

```go
package controller

import (
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service/situation"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type SituationController interface {
	GetOverview(ctx *gin.Context)
	ListChains(ctx *gin.Context)
	GetChainDetail(ctx *gin.Context)
	ListAttackers(ctx *gin.Context)
	GetAttackerProfile(ctx *gin.Context)
	GetTrends(ctx *gin.Context)
	ListRules(ctx *gin.Context)
	CreateRule(ctx *gin.Context)
	UpdateRule(ctx *gin.Context)
	DeleteRule(ctx *gin.Context)
	QuickAction(ctx *gin.Context)
}

type SituationControllerImpl struct {
	engine        *situation.Engine
	repo          repository.SituationRepository
	logger        zerolog.Logger
}

// ... 所有接口方法实现（每个 GET 参数解析 → repo 查询 → 返回统一响应）
```

_(完整代码太长，记录在 plan 的 controller 细节中，这里省略写完整代码以节省上下文。实际执行时每个方法按 repository 模式实现。)_

- [ ] **Step 2: 在 router.go 中添加态势路由**

在 `server/router/router.go` 中：

插入 import：
```go
situationSvc "github.com/mingrenya/AI-Waf/server/service/situation"
```

在 `repo` 创建处添加：
```go
situationRepo := repository.NewSituationRepository(db)
```

在 `service` 创建处添加：
```go
lokiLogService := service.NewLokiLogService()
redisClient, _ := redisPkg.GetClient(context.Background())
situationRuleEngine := situationSvc.NewRuleEngine(lokiLogService, situationRepo)
situationChainBuilder := situationSvc.NewAttackChainBuilder(lokiLogService, situationRepo)
situationProfiler := situationSvc.NewProfiler(lokiLogService, situationRepo)
situationPublisher := situationSvc.NewPublisher(redisClient)
situationEngine := situationSvc.NewEngine(situationRuleEngine, situationChainBuilder, situationProfiler, situationPublisher)
```

在 `authenticated` 组内添加路由：
```go
	// 态势感知
	situationRoutes := authenticated.Group("/situation")
	{
		situationRoutes.GET("/overview", middleware.HasPermission(model.PermWAFLogRead), situationController.GetOverview)
		situationRoutes.GET("/chains", middleware.HasPermission(model.PermWAFLogRead), situationController.ListChains)
		situationRoutes.GET("/chains/:id", middleware.HasPermission(model.PermWAFLogRead), situationController.GetChainDetail)
		situationRoutes.GET("/attackers", middleware.HasPermission(model.PermWAFLogRead), situationController.ListAttackers)
		situationRoutes.GET("/attackers/:ip", middleware.HasPermission(model.PermWAFLogRead), situationController.GetAttackerProfile)
		situationRoutes.GET("/trends", middleware.HasPermission(model.PermWAFLogRead), situationController.GetTrends)
		situationRoutes.GET("/rules", middleware.HasPermission(model.PermConfigRead), situationController.ListRules)
		situationRoutes.POST("/rules", middleware.HasPermission(model.PermConfigUpdate), situationController.CreateRule)
		situationRoutes.PUT("/rules/:id", middleware.HasPermission(model.PermConfigUpdate), situationController.UpdateRule)
		situationRoutes.DELETE("/rules/:id", middleware.HasPermission(model.PermConfigUpdate), situationController.DeleteRule)
		situationRoutes.POST("/quick-action", middleware.HasPermission(model.PermConfigUpdate), situationController.QuickAction)
	}
```

- [ ] **Step 3: 在 main.go 中启动态势引擎并初始化默认规则**

在 `server/main.go` 中，在 `aiTask.Start()` 之后添加：

```go
// 启动态势感知引擎（间隔 30s 执行检测周期）
go func() {
	situationRepo := repository.NewSituationRepository(db)
	lokiLogService := service.NewLokiLogService()
	situationRuleEngine := situationSvc.NewRuleEngine(lokiLogService, situationRepo)
	situationChainBuilder := situationSvc.NewAttackChainBuilder(lokiLogService, situationRepo)
	situationProfiler := situationSvc.NewProfiler(lokiLogService, situationRepo)
	redisClient, redisErr := redisPkg.GetClient(bgCtx)
	situationPublisher := situationSvc.NewPublisher(redisClient)
	situationEngine := situationSvc.NewEngine(situationRuleEngine, situationChainBuilder, situationProfiler, situationPublisher)
	
	// 初始化默认规则
	if err := situationRuleEngine.InitializeDefaults(bgCtx); err != nil {
		config.Logger.Warn().Err(err).Msg("Failed to initialize situation rules")
	}

	if redisErr != nil {
		config.Logger.Warn().Err(redisErr).Msg("Redis unavailable, situation engine running without real-time push")
	}

	config.Logger.Info().Msg("Situation awareness engine started (interval=30s)")
}()

// Start situation engine
situationEngine.Start(bgCtx, 30*time.Second)
defer situationEngine.Stop()
```

- [ ] **Step 4: Commit**

```bash
git add server/controller/situation.go server/router/router.go server/main.go
git commit -m "feat(situation): add controller, routes, and engine startup"
```

---

### Task P1-5: 快速处置聚合 API

**Files:**
- Create: `server/service/situation/quick_action.go`
- Modify: `server/controller/situation.go` — 添加 QuickAction 方法

**Interfaces:**
- Consumes: `repository.BlockedIPRepository`, `repository.IPGroupRepository`, Redis Pub/Sub
- Produces: `POST /api/v1/situation/quick-action`

**Steps:**

- [ ] **Step 1: 创建 server/service/situation/quick_action.go**

```go
package situation

import (
	"context"
	"fmt"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/repository"
	ws "github.com/mingrenya/AI-Waf/server/websocket"
	"github.com/rs/zerolog"
)

type QuickActionService struct {
	blockedIPRepo repository.BlockedIPRepository
	ipGroupRepo   repository.IPGroupRepository
	publisher     *Publisher
	logger        zerolog.Logger
}

// NewQuickActionService 创建快速处置服务
func NewQuickActionService(
	blockedIPRepo repository.BlockedIPRepository,
	ipGroupRepo repository.IPGroupRepository,
	publisher *Publisher,
) *QuickActionService {
	return &QuickActionService{
		blockedIPRepo: blockedIPRepo,
		ipGroupRepo:   ipGroupRepo,
		publisher:     publisher,
		logger:        config.GetServiceLogger("quick-action"),
	}
}

// ExecuteQuickAction 执行一键处置
func (s *QuickActionService) ExecuteQuickAction(ctx context.Context, req QuickActionRequest) (*QuickActionResult, error) {
	result := &QuickActionResult{Success: true}

	switch req.Action {
	case "block", "both":
		duration := req.DurationHours
		blockRecord := &model.BlockedIPRecord{
			IP:         req.SourceIP,
			Reason:     req.Reason,
			RequestURI: "",
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Time{},
		}
		if duration > 0 {
			blockRecord.ExpiresAt = time.Now().Add(time.Duration(duration) * time.Hour)
		}
		if err := s.blockedIPRepo.CreateBlockedIP(ctx, blockRecord); err != nil {
			return nil, fmt.Errorf("封禁IP失败: %w", err)
		}
		result.Blocked = true
		fallthrough
	}

	if req.Action == "blacklist", req.Action == "both":
		// ... 调用 ipGroupRepo 添加到系统默认黑名单
		result.Blacklisted = true
	}

	result.Action = req.Action
	result.SourceIP = req.SourceIP
	result.Note = fmt.Sprintf("已对 IP %s 执行 [%s] 处置，原因: %s", req.SourceIP, req.Action, req.Reason)

	// 通知前端刷新
	ws.GetHub().BroadcastJSON("situation:quick_action", result)

	s.logger.Info().
		Str("ip", req.SourceIP).
		Str("action", req.Action).
		Str("reason", req.Reason).
		Msg("Quick action executed")

	return result, nil
}

type QuickActionRequest struct {
	SourceIP      string `json:"source_ip"`
	Action        string `json:"action"`
	DurationHours int    `json:"duration_hours"`
	Reason        string `json:"reason"`
	CorrelationID string `json:"correlation_id"`
}

type QuickActionResult struct {
	Success     bool   `json:"success"`
	SourceIP    string `json:"source_ip"`
	Action      string `json:"action"`
	Blocked     bool   `json:"blocked"`
	Blacklisted bool   `json:"blacklisted"`
	Note        string `json:"note"`
}
```

- [ ] **Step 2: 在 controller/situation.go 中添加 QuickAction 方法**

```go
func (c *SituationControllerImpl) QuickAction(ctx *gin.Context) {
	var req situation.QuickActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误", err)
		return
	}

	result, err := c.quickActionSvc.ExecuteQuickAction(ctx.Request.Context(), req)
	if err != nil {
		response.InternalServerError(ctx, "处置失败", err)
		return
	}
	response.Success(ctx, result.Note, result)
}
```

- [ ] **Step 3: Commit**

```bash
git add server/service/situation/quick_action.go server/controller/situation.go
git commit -m "feat(situation): add quick action service and API for one-click IP blocking"
```

---

## Phase 2: 前端态势感知 (P0-P1)

### Task P2-1: 前端 TypeScript 类型 + API 接口

**Files:**
- Create: `web/src/types/situation.ts`
- Create: `web/src/api/situation.ts`

**Interfaces:**
- Consumes: None (new files)
- Produces: TypeScript types mirroring Go DTOs + API client functions

**Steps:**

- [ ] **Step 1: 创建 web/src/types/situation.ts**

```typescript
// 攻击阶段
export type AttackStage =
  | 'unknown'
  | 'reconnaissance'
  | 'scanning'
  | 'exploitation'
  | 'lateral_movement'
  | 'command_and_control'
  | 'exfiltration';

// 态势概览
export interface SituationOverview {
  active_chains: number;
  total_chains_24h: number;
  total_attackers_24h: number;
  overall_risk_score: number;
  risk_trend: 'rising' | 'stable' | 'falling';
  top_attack_types: CountItem[];
  top_attacker_ips: CountItem[];
  top_target_sites: CountItem[];
  by_country: CountItem[];
}

export interface CountItem {
  label: string;
  count: number;
}

// 攻击链
export interface ChainSummary {
  id: string;
  source_ip: string;
  geo_country: string;
  stages: string[];
  risk_score: number;
  first_seen: string;
  last_seen: string;
  active: boolean;
}

export interface ChainListResponse {
  chains: ChainSummary[];
  total: number;
  page: number;
  page_size: number;
}

export interface ChainStageItem {
  stage: string;
  technique: string;
  detected_at: string;
  confidence: number;
  evidence: string[];
}

export interface AttackerProfileDetail {
  source_ip: string;
  geo_country: string;
  geo_city?: string;
  total_attacks: number;
  unique_attack_types: number;
  top_attack_type: string;
  unique_target_sites: number;
  active_hours: number[];
  attack_phase: string;
  tools_identified: string;
  is_automated: boolean;
  is_persistent: boolean;
  risk_score: number;
  risk_label: string;
  first_seen: string;
  last_seen: string;
  recent_events: LogEventItem[];
}

export interface LogEventItem {
  id: string;
  attack_type: string;
  severity: string;
  action: string;
  site_domain: string;
  correlation_id: string;
  timestamp: string;
}

export interface ChainDetail {
  id: string;
  source_ip: string;
  geo_country: string;
  stages: ChainStageItem[];
  correlation_ids: string[];
  risk_score: number;
  risk_label: string;
  first_seen: string;
  last_seen: string;
  active: boolean;
  attacker_profile?: AttackerProfileDetail;
}

// 攻击者
export interface AttackerSummary {
  source_ip: string;
  geo_country: string;
  total_attacks: number;
  unique_attack_types: number;
  top_attack_type: string;
  attack_phase: string;
  risk_score: number;
  risk_label: string;
  last_seen: string;
}

// 趋势
export interface TrendPoint {
  timestamp: number;
  total_events: number;
  blocked_count: number;
  detect_count: number;
  unique_ips: number;
}

export interface TrendResponse {
  timeline: TrendPoint[];
  frequent_types: CountItem[];
  active_attackers: number;
  new_chains_24h: number;
}

// 规则
export interface SituationRule {
  id: string;
  name: string;
  stage: string;
  logql: string;
  interval_seconds: number;
  threshold: number;
  severity: string;
  mitre_tactic: string;
  mitre_technique: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

// 快速处置
export interface QuickActionRequest {
  source_ip: string;
  action: 'block' | 'blacklist' | 'both';
  duration_hours: number;
  reason: string;
  correlation_id?: string;
}

export interface QuickActionResponse {
  success: boolean;
  source_ip: string;
  action: string;
  blocked: boolean;
  blacklisted: boolean;
  note: string;
}

// WebSocket 消息
export interface WSSituationMessage {
  type: 'situation:alert' | 'situation:update' | 'situation:attack' | 'situation:quick_action';
  payload: unknown;
  time: string;
}
```

- [ ] **Step 2: 创建 web/src/api/situation.ts**

```typescript
import api from './index';
import type {
  SituationOverview,
  ChainListResponse,
  ChainDetail,
  AttackerSummary,
  AttackerProfileDetail,
  TrendResponse,
  SituationRule,
  QuickActionRequest,
  QuickActionResponse,
} from '@/types/situation';

const BASE = '/situation';

// 态势概览
export const getOverview = () =>
  api.get<{ data: SituationOverview }>(`${BASE}/overview`);

// 攻击链
export const listChains = (params: {
  source_ip?: string;
  stage?: string;
  active?: boolean;
  page?: number;
  page_size?: number;
}) => api.get<{ data: ChainListResponse }>(`${BASE}/chains`, { params });

export const getChainDetail = (id: string) =>
  api.get<{ data: ChainDetail }>(`${BASE}/chains/${id}`);

// 攻击者
export const listAttackers = (params: {
  page?: number;
  page_size?: number;
  sort_by?: string;
  risk_label?: string;
}) => api.get<{ data: { attackers: AttackerSummary[]; total: number } }>(
  `${BASE}/attackers`,
  { params },
);

export const getAttackerProfile = (ip: string) =>
  api.get<{ data: AttackerProfileDetail }>(`${BASE}/attackers/${ip}`);

// 趋势
export const getTrends = (duration = '24h') =>
  api.get<{ data: TrendResponse }>(`${BASE}/trends`, { params: { duration } });

// 规则
export const listRules = () =>
  api.get<{ data: SituationRule[] }>(`${BASE}/rules`);

export const createRule = (rule: Omit<SituationRule, 'id' | 'created_at' | 'updated_at'>) =>
  api.post(`${BASE}/rules`, rule);

export const updateRule = (id: string, rule: Partial<SituationRule>) =>
  api.put(`${BASE}/rules/${id}`, rule);

export const deleteRule = (id: string) =>
  api.delete(`${BASE}/rules/${id}`);

// 快速处置
export const quickAction = (req: QuickActionRequest) =>
  api.post<{ data: QuickActionResponse }>(`${BASE}/quick-action`, req);
```

- [ ] **Step 3: Commit**

```bash
git add web/src/types/situation.ts web/src/api/situation.ts
git commit -m "feat(web): add situation awareness TypeScript types and API client"
```

---

### Task P2-2: 前端态势大屏页面

**Files:**
- Create: `web/src/pages/situation/layout.tsx`
- Create: `web/src/pages/situation/page.tsx`
- Create: `web/src/feature/situation/hooks/useSituationData.ts`
- Create: `web/src/feature/situation/hooks/useSituationWebSocket.ts`
- Create: `web/src/feature/situation/components/SituationDashboard.tsx`
- Create: `web/src/feature/situation/components/AttackChainTimeline.tsx`
- Create: `web/src/feature/situation/components/AttackerRankingChart.tsx`
- Create: `web/src/feature/situation/components/AttackerDrawer.tsx`
- Create: `web/src/feature/situation/components/QuickActionToolbar.tsx`
- Modify: `web/src/routes/config.tsx` — 添加态势感知路由
- Modify: `web/src/components/common/sidebar-nav.tsx` — 添加态势入口

本任务的 Steps 因前端代码量较大，主要在子代理执行时按顺序开发。

- [ ] **Step 1: 创建 useSituationWebSocket hook**
- [ ] **Step 2: 创建 useSituationData hook**
- [ ] **Step 3: 创建 SituationDashboard 组件（复用已有 SecurityDashboard 组件）**
- [ ] **Step 4: 创建 AttackChainTimeline 组件**
- [ ] **Step 5: 创建 AttackerRankingChart 组件**
- [ ] **Step 6: 创建 AttackerDrawer 组件**
- [ ] **Step 7: 创建 QuickActionToolbar 组件**
- [ ] **Step 8: 创建路由页面（layout + page）**
- [ ] **Step 9: 注册路由 + 侧边栏入口**

---

### Task P2-3: 日志界面增强 — 攻击详情 + 快捷处置

**Files:**
- Modify: `web/src/feature/log/components/AttackDetailDialog.tsx`
- Create: `web/src/feature/flow-control/components/QuickBlockDialog.tsx`

---

## Phase 3: Nuclei 集成 (P1)

### Task P3-1: Nuclei SDK 集成 + 扫描服务

### Task P3-2: Nuclei 控制器 + 路由

### Task P3-3: 前端 Nuclei 管理页面

---

## Phase 4: 收尾 (P2)

### Task P4-1: Nuclei AI 闭环事件触发

### Task P4-2: 态势告警通道打通

### Task P4-3: 国际化翻译更新
