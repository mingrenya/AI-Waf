# AI-Waf 功能增强设计文档

> 日期: 2026-07-02 | 状态: 设计评审通过

## 概述

本次设计覆盖三个功能的增强：
1. **Nuclei 漏洞扫描集成** — 主动安全扫描 + WAF 规则闭环验证
2. **态势感知日志系统** — 基于 Loki/LogQL 的攻击链追踪与实时风险评估
3. **日志界面快速处置联动** — 攻击事件详情中一键封禁/加黑名单

---

## 一、Nuclei 漏洞扫描集成

### 1.1 方案选择

**在 server 进程内直接 import nuclei Go SDK**，不设独立容器。

- 依赖：`github.com/projectdiscovery/nuclei/v3/lib`
- 使用 `ThreadSafeNucleiEngine` 支持并发扫描
- 受已有 JWT + RBAC 保护，不额外暴露端口
- 模板目录 volume mount 到 `~/.config/nuclei/nuclei-templates`

### 1.2 新增模块

```
server/
├── service/nuclei/
│   ├── scanner.go          # ThreadSafeNucleiEngine 封装，任务生命周期
│   ├── result_handler.go   # 扫描结果回调 → MongoDB + 审计日志
│   ├── template_mgr.go     # 模板加载/过滤/热更新
│   └── event_trigger.go    # AI 分析器事件驱动的针对性扫描
├── controller/nuclei.go    # REST API + WebSocket 进度推送
├── repository/nuclei.go    # 扫描任务/报告持久化
├── dto/nuclei.go           # 请求/响应 DTO
└── cornjob/nuclei/
    └── scheduler.go        # 定时扫描任务
```

### 1.3 REST API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/nuclei/scan` | POST | 发起扫描（指定站点/模板/参数） |
| `/api/v1/nuclei/scan/:task_id` | GET | 查询任务状态（含进度） |
| `/api/v1/nuclei/scan/:task_id/cancel` | POST | 取消扫描 |
| `/api/v1/nuclei/templates` | GET | 列出可用模板 |
| `/api/v1/nuclei/templates/reload` | POST | 热加载模板目录 |
| `/api/v1/nuclei/config` | GET/PUT | 扫描配置管理 |
| `/api/v1/nuclei/reports` | GET | 历史报告列表 |
| `/api/v1/nuclei/reports/:id` | GET | 报告详情 |
| `/api/v1/nuclei/schedule` | GET/PUT | 定时扫描调度配置 |
| `/api/v1/nuclei/event-trigger` | POST | AI 事件触发的验证扫描（内部调用） |

### 1.4 触发来源

| 触发源 | 实现位置 | 说明 |
|--------|----------|------|
| 定时任务 | `cornjob/nuclei/scheduler.go` | 按站点 cron 定时执行，支持全量/增量 |
| AI 分析器事件 | `service/nuclei/event_trigger.go` | AI 检测到新攻击模式 → 生成模板 → 验证 WAF  |
| 手动触发 | `controller/nuclei.go POST /scan` | 前端控制台发起 |
| MCP 工具 | `mcp-server/tools/nuclei.go` | AI 助手通过 MCP 协议发起 |

### 1.5 AI 闭环：检测 → 验证 → 加固

```
AI Analyzer 检测到新攻击模式
    │
    ├── 生成 Nuclei 验证模板
    ├── 对受保护站点运行模板
    ├── 结果：
    │   ├── WAF 拦截失败 → 生成紧急规则建议 → 推送管理员
    │   └── WAF 拦截成功 → 确认规则有效 → 记录审计
    └── 前端展示闭环状态
```

### 1.6 MongoDB 集合

```json
// nuclei_tasks
{ "_id": "uuid", "site_id": "...", "trigger": "scheduled|manual|ai_analyzer|mcp",
  "config": {...}, "status": "pending|running|completed|failed|cancelled",
  "progress": { "total": 100, "completed": 45 }, "created_at": "...", "completed_at": "..." }

// nuclei_reports (沿用 nuclei_tasks 中 result 字段)
{ "_id": "uuid", "task_id": "...", "findings": [...], "summary": {...},
  "duration_seconds": 245, "created_at": "..." }
```

---

## 二、态势感知日志系统

### 2.1 核心理念

**LogQL 就是规则匹配引擎。** 不引入额外第三方规则引擎，所有检测/关联/聚合通过 LogQL 查询实现。

### 2.2 数据流架构

```
HAProxy/Coraza/WAF 日志
    │
    ▼
Promtail ──添加结构化 labels──► Loki (已有)
    │   correlation_id           │
    │   source_ip                │
    │   attack_type              │ LogQL 查询
    │   action                    │
    │   severity                  ▼
    │   geo_country        态势感知引擎
    │   rule_id             (situation/engine.go)
    │   waf_phase                │
    │   site_id          ┌───────┼───────┐
    │                    ▼       ▼       ▼
    │              Redis Stream  MongoDB  WebSocket
    │              (实时事件)   (持久化)   (推送前端)
    ▼
   Loki 存储
```

### 2.3 日志标签体系

`correlation_id` 是整个追踪系统的主键。`server/middleware/request_id.go` 生成后贯穿整条链路：

```
HAProxy → SPOE → Coraza → server API → WAF log → Loki
     (所有环节携带同一个 correlation_id)
```

Promtail 配置 `pipeline_stages` 解析 JSON 提取字段为 label。

### 2.4 LogQL 检测规则库

```go
type SituationRule struct {
    ID             string        // rule_id
    Name           string        // "高频SQL注入扫描"
    Stage          AttackStage   // reconnaissance / scanning / exploitation / ...
    LogQL          string        // 预定义 LogQL 查询
    Interval       time.Duration // 执行频率
    Threshold      int           // 触发阈值
    Severity       string        // critical / high / medium / low
    MITRETactic    string        // MITRE ATT&CK 战术 ID
    MITRETechnique string        // MITRE ATT&CK 技术 ID
    Enabled        bool
}
```

**示例规则：**

| 规则名 | LogQL | 检测目标 |
|--------|-------|----------|
| 端口扫描 | `sum by(source_ip)(count_over_time({attack_type="scanner"}[5m])) > 100` | Active Reconnaissance |
| SQL注入利用 | `sum by(source_ip)(count_over_time({attack_type="sql_injection",severity=~"critical|high"}[5m])) > 10` | Exploit Public-Facing App |
| 多面探测 | `count(distinct attack_type) by(source_ip) > 3` over 10m window | Active Scanning |
| 漏洞扫描器指纹 | `{attack_type="vulnerability_scanner"} |~ "(?i)(nuclei|nikto|nessus|openvas)"` | Vulnerability Scanning |

规则保存在 MongoDB `situation_rules` 集合，支持前端 CRUD + 启用/禁用。

### 2.5 攻击链追踪

```go
type AttackChain struct {
    IP             string          // 攻击源 IP
    GeoCountry     string          // 地理位置
    Stages         []ChainStage    // 攻击链路
    CorrelationIDs []string        // 链路关联的请求 ID
    RiskScore      int             // 实时风险评分 0-100
    FirstSeen      time.Time
    LastSeen       time.Time
    Active         bool
}

type ChainStage struct {
    Stage      AttackStage // reconnaissance / scanning / exploitation / lateral_movement / c2 / exfiltration
    Technique  string      // MITRE ATT&CK T-code (T1595, T1046, etc.)
    DetectedAt time.Time
    Evidence   []LogEvent  // 证据事件
    Confidence float64     // 置信度 0-1
}
```

**追踪流程：**

1. 态势引擎定时（30s）执行检测规则
2. LogQL 命中 → 提取 source_ip / attack_type
3. 追加 LogQL 查询该 IP 历史事件 → 构建完整攻击链时间线
4. 攻击阶段迁移（scanning → exploitation）触发告警

**Redis 存储（实时，TTL 24h）：**
```
chain:ip:{source_ip}     → Hash  (当前状态)
chain:events:{source_ip} → Stream (事件流)
chain:stages:{source_ip} → ZSet  (阶段时间线)
```

**MongoDB 存储（持久化）：**
```
attack_chains 集合 — 完整攻击链记录 + 分析摘要
```

### 2.6 攻击者画像

```go
type AttackerProfile struct {
    IP             string `json:"ip"`
    GeoCountry     string `json:"geo_country"`
    GeoCity        string `json:"geo_city,omitempty"`

    // 攻击统计 (24h)
    TotalAttacks      int      `json:"total_attacks"`
    UniqueAttackTypes int      `json:"unique_attack_types"`
    TopAttackType     string   `json:"top_attack_type"`
    UniqueTargetSites int      `json:"unique_target_sites"`

    // 时间特征
    ActiveHours    []int    `json:"active_hours"`     // 活跃时段 0-23
    BurstIntervals []string `json:"burst_intervals"`  // 爆发时间窗

    // 行为特征
    AttackPhase       string `json:"attack_phase"`         // 当前阶段
    ToolsIdentified   string `json:"tools_identified"`     // 识别到的工具
    IsAutomated       bool   `json:"is_automated"`         // 自动化工具
    IsPersistent      bool   `json:"is_persistent"`        // 持续 > 24h

    // 风险评估
    RiskScore int    `json:"risk_score"`    // 0-100
    RiskLabel string `json:"risk_label"`    // low/medium/high/critical
}
```

数据来源：Loki LogQL 聚合查询，实时部分存 Redis（Hash），全量持久化 MongoDB。

### 2.7 风险评分公式

```go
func calculateRisk(p AttackerProfile) int {
    score := 0
    score += min(p.TotalAttacks * 3, 30)           // 攻击频率 (0-30)
    score += min(p.UniqueAttackTypes * 5, 20)      // 攻击多样性 (0-20)
    score += attackStageWeight[p.AttackPhase]      // 攻击阶段 (0-20)
    if p.IsPersistent   { score += 15 }             // 持续性 (0-15)
    if p.IsAutomated    { score += 15 }             // 自动化 (0-15)
    return min(score, 100)
}

var attackStageWeight = map[string]int{
    "reconnaissance":    5,
    "scanning":         10,
    "exploitation":     20,
    "lateral_movement": 30,
    "c2":               30,
    "exfiltration":     40,
}
```

### 2.8 新增模块

```
server/
├── service/situation/
│   ├── engine.go        # 态势聚合引擎主逻辑（定时执行规则/构建链）
│   ├── rule.go          # LogQL 检测规则定义 + 评估
│   ├── attack_chain.go  # 攻击链构建 + 阶段迁移判定
│   ├── profiler.go      # 攻击者画像构建
│   ├── risk.go          # 风险评分 + 风险等级
│   ├── publisher.go     # Redis Pub/Sub 发布态势事件
│   └── logql_builder.go # LogQL 动态查询构造器
├── controller/situation.go
├── repository/situation.go
├── dto/situation.go
└── cornjob/situation/
    └── engine_runner.go  # 态势引擎定时触发
```

### 2.9 REST API

```
GET  /api/v1/situation/overview             态势概览
GET  /api/v1/situation/chains               攻击链列表（分页/筛选）
GET  /api/v1/situation/chains/:id           单链详情
GET  /api/v1/situation/attackers            攻击者排行榜
GET  /api/v1/situation/attackers/:ip        攻击者画像
GET  /api/v1/situation/trends               趋势统计
GET  /api/v1/situation/rules                检测规则列表
POST /api/v1/situation/rules                创建规则
PUT  /api/v1/situation/rules/:id            更新规则
DELETE /api/v1/situation/rules/:id          删除规则
POST /api/v1/situation/quick-action         一键处置（封禁+黑名单+审计）
```

### 2.10 WebSocket 推送

频道设计：

| 频道 | 推送内容 | 频率 |
|------|----------|------|
| `situation:alert` | 新攻击链告警 | 即时 |
| `situation:update` | 风险评分变动 | 分钟级 |
| `situation:attack` | 实时攻击事件推送 | 秒级 |
| `nuclei:progress` | Nuclei 扫描进度 | 秒级 |

### 2.11 前端模块

```
web/src/
├── api/
│   ├── nuclei.ts            # Nuclei API
│   └── situation.ts         # 态势感知 API
├── feature/
│   ├── nuclei/
│   │   ├── components/
│   │   │   ├── ScanTaskTable.tsx        # 扫描任务列表
│   │   │   ├── ScanResultDialog.tsx     # 扫描结果详情
│   │   │   ├── ScanScheduler.tsx        # 定时扫描配置
│   │   │   └── TemplateManager.tsx      # 模板管理
│   │   └── hooks/useNuclei.ts
│   └── situation/
│       ├── components/
│       │   ├── SituationDashboard.tsx       # 态势总览（集成已有 SecurityDashboard）
│       │   ├── AttackChainTimeline.tsx      # 攻击链时间线（MITRE 阶段着色）
│       │   ├── AttackerDrawer.tsx           # 攻击者详情侧边面板
│       │   ├── AttackerRankingChart.tsx      # Top-N 排行
│       │   ├── AttackTypeDistribution.tsx    # 攻击分布饼图
│       │   ├── RiskScoreGauge.tsx            # 风险评分仪表盘
│       │   ├── CorrelationGraph.tsx          # IP→攻击→站点关联图
│       │   └── QuickActionToolbar.tsx        # 快速处置工具栏
│       └── hooks/
│           ├── useSituationWebSocket.ts      # WebSocket 数据
│           └── useSituationData.ts           # REST 数据
├── pages/
│   ├── nuclei/layout.tsx
│   ├── nuclei/pages/
│   │   ├── scan/page.tsx        # 扫描管理
│   │   └── templates/page.tsx   # 模板管理
│   └── situation/
│       ├── layout.tsx           # 态势大屏布局
│       └── page.tsx             # 态势主页面
└── types/
    ├── nuclei.ts
    └── situation.ts
```

---

## 三、日志界面快速处置联动

### 3.1 联动流程

```
日志列表 / 攻击事件详情页
    │
    ├── 点击攻击事件 → 展开详情面板
    │   ├── 攻击者画像（态势数据）
    │   ├── 关联事件列表（同 IP 的其他攻击）
    │   ├── 攻击链时间线
    │   │
    │   └── 一键处置按钮组
    │       ├── 🚫 立即封禁 IP (1h / 24h / 7d / 永久)
    │       ├── ⚠️  加入黑名单 IP 组
    │       ├── 📊 查看攻击链
    │       ├── 🔍 查看引擎日志 (LogQL)
    │       └── 🔗 复制 correlation_id
    │
    └── 处置后效果
        ├── 封禁即时生效 → coraza-spoa IP Radix Tree 热更新
        ├── 审计日志自动记录
        └── 告警通道同步通知
```

### 3.2 增强/新增的前端组件

```
feature/log/components/
├── AttackDetailDialog.tsx    # 增强：攻击者画像面板 + 处置按钮组
├── AttackLogFilter.tsx       # 增强：批量选中操作
└── QuickActionToolbar.tsx    # 新增：选中行浮动工具栏

feature/flow-control/components/
└── QuickBlockDialog.tsx      # 新增：快速封禁确认框（含时长选择+原因预设）
```

### 3.3 聚合处置 API

```json
POST /api/v1/situation/quick-action
{
    "source_ip": "1.2.3.4",
    "action": "block",           // block | blacklist | both
    "duration_hours": 24,        // 0 = 永久
    "reason": "高频SQL注入扫描",
    "correlation_id": "req-abc-123"
}

Response: {
    "success": true,
    "block_record": {...},
    "audit_log_id": "...",
    "note": "封禁已即时同步至 WAF 引擎"
}
```

后端一次调用完成：封禁记录创建 + IP 组同步 + 审计日志写入 + Pub/Sub 推送。

### 3.4 已有 API 复用

无需修改，快速处置直接调用已有接口：

- `POST /api/v1/blocked-ips` — 封禁 IP
- `POST /api/v1/ip-groups/blacklist/add` — 加入黑名单
- `GET /api/v1/situation/chains?source_ip=...` — 查询攻击链
- `POST /api/v1/log/loki-query` — 查询引擎日志

---

## 四、基础设施变更

### 4.1 Docker Compose 新增

```yaml
# 在 docker-compose.yaml 中新增
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

# volumes 新增
redis_data:
  driver: local
```

### 4.2 Redis 数据分层

```
Redis 职责 — 实时数据层
├── stream:attack_events       攻击事件流
├── hash:attacker:{ip}         攻击者画像（实时）
├── zset:attacker_ranking      攻击者排行榜
├── string:risk:{ip}           实时风险评分
├── hash:scan:{task_id}        Nuclei 扫描进度
├── pubsub:situation:alert     态势告警推送
├── pubsub:situation:update    态势更新推送
├── pubsub:situation:attack    攻击事件推送
├── pubsub:nuclei:progress     Nuclei 进度推送
└── ratelimit:*                限流计数器（已有）
```

### 4.3 MongoDB 新增集合

| 集合 | 说明 |
|------|------|
| `nuclei_tasks` | Nuclei 扫描任务 |
| `nuclei_reports` | Nuclei 扫描报告 |
| `attack_chains` | 攻击链持久化 |
| `attacker_profiles` | 攻击者画像持久化 |
| `situation_rules` | 态势检测规则 |
| `situation_snapshots` | 态势快照（定时聚合统计） |

### 4.4 Go 依赖新增

```
server/go.mod:
  github.com/projectdiscovery/nuclei/v3
  github.com/redis/go-redis/v9      (已有，独立容器运行保障)
  github.com/gorilla/websocket
```

---

## 五、实现优先级

| 优先级 | 阶段 | 内容 |
|--------|------|------|
| 🔴 P0 | 基础设施 | Redis 容器 + WebSocket 基础设施 + server 集成 |
| 🔴 P0 | 态势-核心 | 日志标签标准化 + LogQL 规则引擎 + 攻击链追踪 |
| 🔴 P0 | 快速处置 | 聚合 API + 日志界面增强 + 封禁联动 |
| 🟡 P1 | Nuclei | SDK 集成 + 扫描 API + 报告存储 |
| 🟡 P1 | 态势-画像 | 攻击者画像 + 风险评分 + REST API |
| 🟡 P1 | 前端-态势 | 态势大屏 + 攻击链时间线 + 攻击者排行 |
| 🟢 P2 | Nuclei-AI闭环 | AI 事件触发扫描 + 规则验证闭环 |
| 🟢 P2 | 前端-Nuclei | 扫描管理页 + 模板管理页 + 进度推送 |
| 🟢 P2 | 前端-告警 | 态势告警通道打通（钉钉/企微等） |
