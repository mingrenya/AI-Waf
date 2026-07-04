# AI-Waf 下一步开发计划

## 当前状态

| 类别 | 已实现 | 缺失 |
|------|--------|------|
| WAF 核心能力 | 27 项已实现 | 17 项待补 |

---

## 下一步任务（按优先级排列）

### 🔴 P0 — 必须补

| # | 功能 | 说明 | 工作量 | 实现方式 |
|---|------|------|--------|----------|
| 1 | **取证捕获 WAF 侧对接** | `forensics.go` 已写好，需在 coraza-spoa 攻击事件处调用 `ForensicsCapture.CaptureAttackTraffic` | 1h | 攻击检测时异步触发 |
| 2 | **Logger.Panic() 修复** | `agent.go` 5 处 `Logger.Panic()` → 优雅 Error 返回，防止 SPOE 连接断开 | 2h | 改 Panic 为 Error 日志 + 返回 nil |
| 3 | **Bot 管理** | UA 指纹 + 行为频率统计 → 爬虫/自动化工具识别 | 8h | `coraza-spoa/internal/bot_detector.go` |

### 🟡 P1 — 安全增强

| # | 功能 | 说明 | 工作量 |
|---|------|------|--------|
| 4 | **Anti-逃逸编码** | 在 Coraza 检测前预处理：url_decode → unicode → base64 → hex → 大小写还原 → 同形字符映射 | 8h |
| 5 | **网页防篡改** | 页面哈希快照 → 响应比对 → 被篡改时返回缓存副本 | 6h |
| 6 | **DDoS 协同 IP 黑名单** | 封禁 IP 上报 + 拉取共享黑名单（对接外部威胁情报 API） | 5h |

### 🟢 P2 — 运维与体验

| # | 功能 | 说明 | 工作量 |
|---|------|------|--------|
| 7 | **API 资产自动发现** | 流量采样 → OpenAPI 推断 → 未授权接口检测 | 10h |
| 8 | **规则命中统计面板** | Coraza/MicroEngine 命中次数 Top N → ECharts 图表 | 4h |
| 9 | **防护事件误报反馈** | 攻击日志详情新增"标记误报"→ 自动加入白名单 | 3h |
| 10 | **定时扫描任务** | Nuclei Cron 调度 → 扫描结果 → 邮件/钉钉通知 | 3h |
| 11 | **TLS 版本/加密套件 UI** | HAProxy ssl-min-ver / ciphers 可视化配置 | 3h |
| 12 | **非标端口支持 UI** | 前端 Sites 增加自定义端口配置 | 2h |
| 13 | **GeoIP 配置持久化** | `os.Setenv` → MongoDB 存储 | 2h |
| 14 | **前端证捕获会话列表增强** | 增加"攻击事件关联"列 + 一键下载按钮 + 自动捕获状态标签 | 3h |

### ⚪ P3 — 质量补强

| # | 功能 | 说明 | 工作量 |
|---|------|------|--------|
| 15 | Controller 单元测试 | 24 个 controller，当前只有 1 个有测试 | 分散进行 |
| 16 | Repository 单元测试 | 17 个 repo，当前 0 个有测试 | 分散进行 |
| 17 | Integration 测试 | WAF 全链路：HAProxy → SPOE → Coraza → LogStore → MongoDB | 8h |

---

## 总计

| 优先级 | 项数 | 总工作量 |
|--------|------|----------|
| P0 | 3 | ~11h |
| P1 | 3 | ~19h |
| P2 | 8 | ~30h |
| P3 | 3 | ~16h+ |
| **合计** | **17** | **~76h** |

---

## 建议执行顺序

```
1. P0 取证捕获对接（1h） — 打通自动抓包全链路
2. P0 Logger.Panic() 修复（2h） — 消除生产稳定性隐患
3. P0 Bot 管理（8h） — 企业 WAF 必备能力
4. P2 GeoIP 持久化（2h） — 快速修复已识别问题
5. P1 Anti-逃逸编码（8h） — 提升绕过防御难度
6. P2 规则命中面板（4h） — 让运维看见 WAF 效果
7. 剩余按需推进
```

---

## 环境变量开关参考

```sh
# ── WAF 高级防护开关 ──
SENSITIVE_DATA_FILTER=true  # 信息泄露脱敏
SCAN_PROTECTION=true        # 扫描防护
FORENSICS_CAPTURE=true      # 取证式自动流量捕获
```
