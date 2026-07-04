# Nuclei 漏洞扫描 vs 云WAF扫描防护 — 概念与实现分析

## 一、核心区别

很多人把"Nuclei扫描"和"WAF扫描防护"混为一谈，但它们是两个**完全不同方向**的功能：

| 维度 | Nuclei 漏洞扫描 | 云WAF 扫描防护 |
|------|---------------|---------------|
| **方向** | 主动出击 — 扫描外部目标 | 被动防守 — 拦截外界扫描 |
| **使用者** | 安全人员自查资产 | WAF 自动运行 |
| **触发方式** | 手动/定时触发扫描任务 | 恶意流量经过时自动触发 |
| **目标** | 发现自身漏洞 | 阻止攻击者探测 |
| **比喻** | 你是攻击者，扫描别人 | 你是防守者，阻止别人扫描你 |

---

## 二、AI-Waf 当前状态

### ✅ 已有：Nuclei 漏洞扫描

```
server/service/nuclei/
├── scanner.go       # 扫描核心 (封装 nuclei 引擎)
├── template_mgr.go  # 模板管理
├── paths.go         # 模板路径探测
└── result_handler.go # 结果处理

web/src/pages/nuclei/
├── pages/scan/      # 扫描页面
└── pages/templates/ # 模板管理页面
```

**功能**：
- 选择目标站点 + Nuclei 模板 → 发起扫描
- YAML 模板匹配漏洞特征 → 产生漏洞报告
- 可从 WAF 管理后台直接发起扫描

**这就是你之前修复的那个模块**（模板路径 500 错误）。

### ❌ 缺失：WAF 扫描防护

云WAF的扫描防护是指**WAF 自动识别并拦截外界扫描器**对受保护站点的探测。AI-Waf 现在没有这个能力。

```
外界扫描器 (Nuclei/sqlmap/Nessus/Nikto...)
      │
      ▼
   HAProxy 端口 (80/443)
      │
      ▼
   Coraza WAF ─── 能检测攻击payload ✓
      │           不能识别扫描器指纹 ✗
      │           不能按频率自动封禁 ✗
      ▼
   MicroEngine ─── 有自定义规则能力 ✓
                   但没配扫描防护规则 ✗
```

---

## 三、云WAF扫描防护的三大机制

参考阿里云WAF扫描防护模块：

### 3.1 高频扫描封禁

```
原理: 统计单IP在时间窗口内触发不同规则的次数和种类
触发条件 (可配置):
  - 检测窗口: 60秒
  - 触发规则次数 > 20次
  - 触发不同子规则种类 > 2种
动作:
  - 将该IP自动加入临时黑名单, 封禁 1800秒 (30分钟)
  - 或切换为"观察模式", 仅记录日志不拦截
```

### 3.2 目录遍历封禁

```
原理: 监控单IP请求大量不存在的目录
触发条件 (可配置):
  - 检测窗口: 10秒
  - 请求次数 > 50次
  - 不同目录数 > 50个
  - 404响应码占比 >= 70%
动作:
  - 自动封禁
  - 排除 .js / .png 等静态文件, 降低误报
```

### 3.3 扫描工具封禁

```
原理: User-Agent / 请求特征匹配
覆盖工具:
  Sqlmap, AWVS, Nessus, Appscan, Webinspect,
  Netsparker, Nikto, RSAS, Nuclei, BurpSuite ...等

实现: 维护扫描器指纹库, 匹配即拦截
```

---

## 四、AI-Waf 如何实现扫描防护

由于 MicroEngine 已经有 AND/OR 复合条件 + IP组 + 优先级排序能力，扫描防护可以直接用 MicroEngine规则实现：

### 4.1 高频扫描封禁 → MicroEngine规则

```json
{
  "name": "高频扫描自动封禁",
  "type": "blacklist",
  "status": "enabled",
  "priority": 100,
  "condition": {
    "type": "composite",
    "operator": "AND",
    "conditions": [
      {
        "type": "simple",
        "target": "source_ip",
        "match_type": "in_ipgroup",
        "match_value": "system_auto_blocked_scanners"
      }
    ]
  }
}
```
这段规则的意思是：IP如果在`system_auto_blocked_scanners`组中 → 直接封禁。关键是把IP实时加入这个组。

**需要新增的服务**：
```go
// server/service/cornjob/scan_protector.go
type ScanProtector struct {
    counter   map[string]*IPScanCounter  // IP → 计数
}
type IPScanCounter struct {
    triggerCount int       // 触发规则次数
    ruleTypes    map[int]bool  // 触发的不同规则ID
    windowStart  time.Time
}
```

### 4.2 扫描器指纹 → 规则中加 User-Agent 匹配

当前 MicroEngine 的 `MatchType` 只有 `equal/not_equal/fuzzy/in_cidr/not_in_cidr/in_ipgroup/not_in_ipgroup/include/contains/not_contains/prefix_keyword/regex`。但**缺少 Header 级别的 Target**。

需要扩展：
```go
// micro_engine.go 新增
const (
    SourceIP     TargetType = "source_ip"
    TargetURL    TargetType = "url"
    TargetPath   TargetType = "path"
    TargetHeader TargetType = "header"  // ← 新增
)
```
这样就能写：
```json
{
  "type": "simple",
  "target": "header",
  "match_type": "contains",
  "match_value": "User-Agent: sqlmap"
}
```

### 4.3 目录遍历封禁 → 需要 404 计数

MicroEngine 规则匹配发生在 **请求阶段**（HAProxy SPOE → HandleRequest），此时还不知道响应码。需要改为**依赖 WAF 日志统计 + 定时任务** 来实现：

```
每30s: 查询 MongoDB waf_log
    → 聚合: ip, 404_count, unique_paths, total_requests
    → 判定: 404占比 >= 70% && 不同目录 > N
    → 动作: 添加到 system_auto_blocked_scanners IP组
```

---

## 五、实现工作量估算

| 功能 | 实现方式 | 工作量 |
|------|----------|--------|
| 高频扫描封禁 | MicroEngine IP组 + 定时任务 IP计数器 | 小 (~2h) |
| 扫描器指纹封禁 | MicroEngine 扩展 Header Target + 指纹库 | 中 (~4h) |
| 目录遍历封禁 | 定时任务查询WAF日志统计 + 自动封禁 | 中 (~3h) |
| 管理界面 | 前端配置页 (阈值/开关/白名单) | 中 (~3h) |
| **合计** | | **~12h** |

---

## 六、总结

- **Nuclei** = AI-Waf 已有的主动漏洞扫描功能，是"攻击视角"自查工具
- **扫描防护** = AI-Waf 缺失的被动拦截功能，是 WAF 拦截外界扫描器的防御能力
- 扫描防护可以通过扩展现有 MicroEngine + 新增定时任务实现，不需要引入新组件
- 当前态势感知中已经有扫描器指纹检测规则（`漏洞扫描器指纹识别`），但只是检测+告警，没有自动封禁能力
