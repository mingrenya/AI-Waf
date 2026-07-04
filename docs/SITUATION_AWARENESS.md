# AI-Waf 态势感知实现分析

## 一、态势感知实现原理

### 1. 整体架构

```
┌──────────────────────────────────────────────────────────────┐
│                    态势感知引擎 (30s 轮询)                     │
│                                                              │
│   main.go:118  situationEngine.Start(bgCtx, 30*time.Second) │
│                                                              │
│   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   │
│   │  RuleEngine  │──▶│ ChainBuilder │──▶│   Profiler   │   │
│   │  规则评估     │   │  攻击链构建   │   │  攻击者画像   │   │
│   │              │   │              │   │              │   │
│   │ 5条默认规则   │   │ stage 进阶   │   │ 攻击频率     │   │
│   │ LogQL 查询   │   │ 关联ID聚合   │   │ 攻击阶段     │   │
│   └──────┬───────┘   └──────┬───────┘   └──────┬───────┘   │
│          │                  │                  │            │
│          ▼                  ▼                  ▼            │
│     Loki 日志源         MongoDB              MongoDB         │
│  (LogQL 实时查询)   (attack_chains)    (attacker_profiles)  │
│                                                              │
│   ┌──────────────────────────────────────────────────────┐   │
│   │  Publisher ──▶ Redis Pub/Sub + WebSocket Broadcast   │   │
│   │  实时推送告警/更新到前端                                │   │
│   └──────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
```

### 2. 数据流

```
1. WAF引擎拦截攻击 → 写入 MongoDB (waf_log collection)
                           ↓ (同步写入程)
                     Loki 通过 Promtail 采集 Docker 日志
                           ↓
2. RuleEngine 每30s 通过 LogQL 查询 Loki
   - 查询触发的 IP 列表
   - 返回 RuleResult{HitIPs, HitCount, Rule}

3. ChainBuilder.UpdateChainFromRuleResult()
   - 对每个 HitIP，查询是否存在已有的 AttackChain (MongoDB)
   - 不存在 → 创建新链 (Upsert to attack_chains)
   - 已存在 → 追加新 stage，判定 stage 是否进阶

4. Profiler.BuildProfile()
   - 查询 Loki 获取该 IP 24h 内的全部攻击日志
   - 统计：攻击次数、攻击类型分布、攻击时段、目标站点数
   - 判定：是否自动化(>50次 & >2类型)、是否持续(>24h)
   - 写入 MongoDB (attacker_profiles)

5. CalculateRisk() 计算风险评分
   - 频率分数 (max 30) + 多样性分数 (max 20)
   - + 攻击阶段权重 (侦查5/扫描10/利用20/横移30/C2 30/外传40)
   - + 持续性加分(15) + 自动化加分(15)
   - = 0~100 分

6. Publisher 实时推送
   - 评分 >= 60 → PublishAlert (WebSocket + Redis)
   - 每轮 → PublishUpdate
```

### 3. 5条内置检测规则

| 规则名 | LogQL | 阶段 | 严重度 |
|--------|-------|------|--------|
| 高频端口扫描检测 | `sum by (source_ip)(count_over_time({...,attack_type="scanner"}[5m])) > 100` | 扫描 | medium |
| SQL注入高频利用 | `sum by (source_ip)(count_over_time({...,severity=~"critical\|high"}[5m])) > 10` | 利用 | critical |
| 多攻击面探测 | `count by (source_ip)(count_over_time({...,attack_type!=""}[10m])) > 50` | 扫描 | high |
| 漏洞扫描器指纹 | `{...} \|~ "(?i)(nuclei\|nikto\|nessus\|...)"` | 扫描 | high |
| RCE攻击检测 | `sum by (source_ip)(count_over_time({...,attack_type=~"rce"...}[5m])) > 5` | 利用 | critical |

### 4. 攻击阶段映射 (MITRE ATT&CK)

```
侦查(reconnaissance) → 扫描(scanning) → 漏洞利用(exploitation)
                                            ↓
                                      横向移动(lateral_movement)
                                            ↓
                                   C2控制(command_and_control)
                                            ↓
                                      数据外传(exfiltration)
```

---

## 二、功能覆盖状态

### 已实现

| 组件 | 文件 | 状态 |
|------|------|------|
| 引擎定时轮询 | `engine.go` | ✅ 正常 |
| LogQL 规则评估 | `rule.go` | ✅ 正常 (5条默认规则) |
| 攻击链构建 | `attack_chain.go` | ✅ 正常 |
| 攻击者画像 | `profiler.go` | ✅ 正常 |
| 风险评分 | `risk.go` | ✅ 正常 (0-100) |
| 实时推送 | `publisher.go` | ✅ WebSocket + Redis Pub/Sub |
| 一键处置 | `quick_action.go` | ✅ 封禁/黑名单 |
| 数据持久化 | `repository/situation.go` | ✅ 3表 (rules/chains/profiles) |
| REST API | `controller/situation.go` | ✅ 11个端点 |
| 前端仪表盘 | `SituationDashboard.tsx` | ✅ |
| 前端时间线 | `AttackChainTimeline.tsx` | ✅ |
| 前端排名 | `AttackerRankingChart.tsx` | ✅ |
| 前端攻击者详情 | `AttackerDrawer.tsx` | ✅ |
| WebSocket 客户端 | `useSituationWebSocket.ts` | ✅ |

### 依赖链完整性

```
可运行需要：
✅ Loki (docker-compose 已配置)
✅ Promtail (采集 Docker 日志到 Loki)
✅ MongoDB (situation_rules / attack_chains / attcker_profiles)
✅ Redis (Pub/Sub 实时推送, 可选)
⚠️ 数据源 -- 见下方问题
```

---

## 三、发现的关键问题

### 🔴 P0: WAF日志未以结构化方式进入Loki

**问题**: 态势感知引擎的所有 LogQL 查询都依赖 Loki 中存在带 `source_ip`、`attack_type`、`severity` 等 **label** 的日志流。但当前架构中：

1. WAF 攻击日志直接写入 **MongoDB** (`waf_log` collection)，通过异步批量写入实现
2. Loki 通过 **Promtail** 采集 **Docker 容器 stdout/stderr 日志**
3. 容器的 stdout 日志是 zerolog 的 **非结构化文本**，没有 `source_ip`、`attack_type` 等 label

**实际效果**: LogQL 查询 `{container_name="mrya-waf",source_ip="x.x.x.x"}` 永远匹配不到日志，因为 `source_ip` 不存在于 Loki 的 stream labels 中。

**日志链路**:
```
应用层写 MongoDB(log_store.go:206)
    → zerolog 也输出到 stdout
    → Docker json-file logging driver
    → Promtail 采集并加容器标签
    → Loki 只有 container_name 等基础标签
    → LogQL 查询字段不匹配 ❌
```

### 🟡 P2: Loki URL 硬编码

```go
// server/service/loki_log.go:92
baseURL: "http://loki:3100", // Docker Compose 内部 DNS
```
无环境变量覆盖，非 Docker 环境或自定义端口无法使用。

### 🟡 P3: Profiler 时间戳获取不精确

```go
// profiler.go:126-137
func (p *Profiler) findFirstTime(entries *service.LokiLogQueryResponse) time.Time {
    if len(entries.Results) > 0 {
        return time.Now().Add(-24 * time.Hour)  // 固定24h前，非真实首次时间
    }
    return time.Now()
}
```

### 🟢 P4: Trend 接口未实现

```go
// controller/situation.go:236
func (c *SituationControllerImpl) GetTrends(ctx *gin.Context) {
    // 简化的趋势统计
    response.Success(ctx, "获取趋势数据成功", dto.TrendResponse{
        Timeline: []dto.TrendPoint{},  // 始终返回空
        ...
    })
}
```
前端虽然有 `useTrends` hook，但 `TrendChart` 暂未接入页面。

---

## 四、与云WAF功能对齐

### 扫描防护 vs Nuclei

| 维度 | 云WAF扫描防护 | AI-Waf Nuclei |
|------|-------------|---------------|
| 方向 | 防御：检测+拦截外界扫描 | 检测：主动扫描自身站点 |
| 触发 | 被动，流量经过时自动触发 | 主动，用户手动/定时触发 |
| 目标 | 阻止攻击者探测 | 发现自身漏洞 |
| 使用者 | 防守方 | 防守方(自查) |

**Nuclei = 漏洞扫描 ≠ 扫描防护**

当前 AI-Waf 有 Nuclei（主动扫描），但缺云WAF的扫描防护（被动拦截）：
- 高频扫描封禁：某IP 60s内触发规则>20次+>2种规则 → 自动封禁
- 目录遍历封禁：某IP 10s内大量404>70% → 自动封禁
- 扫描器指纹：User-Agent 识别 sqlmap/Nessus/Nuclei 等 → 拦截

### 功能对齐矩阵

| 功能 | 实现状态 |
|------|----------|
| Web基础防护 | ✅ Coraza + OWASP CRS |
| 自定义规则 | ✅ MicroEngine (AND/OR复合) |
| IP黑白名单 | ✅ IP组 + CIDR前缀树 |
| CC防护 | ✅ Alibaba Sentinel |
| GeoIP封禁 | ✅ MaxMind GeoLite2 |
| 告警通知 | ✅ 钉钉/企微/邮件/Slack/Discord |
| RBAC+审计 | ✅ |
| 态势感知 | ⚠️ 架构完整但数据链路不可用 |
| Bot管理 | ❌ 无爬虫/自动化流量识别 |
| 扫描防护 | ❌ 无被动扫描拦截 |
| 信息泄露/脱敏 | ❌ |
| API安全 | ❌ |
| 网页防篡改 | ❌ |

---

## 五、修复优先级

| 优先级 | 问题 | 工作量 |
|--------|------|--------|
| **P0** | 态势感知数据链路打通 | 中（需LogQL查询方式调整或日志格式改造） |
| **P0** | 扫描防护实现 | 小（MicroEngine规则 + Counter统计） |
| **P1** | Loki URL 可配置化 | 极小 |
| **P1** | 信息泄露防护 | 小（响应体正则过滤） |
| **P2** | Bot 管理基础版 | 中（UA统计 + JS Challenge） |
| **P2** | Profiler/Trend 数据精度修复 | 小 |
| **P3** | API安全、防篡改 | 大 |
