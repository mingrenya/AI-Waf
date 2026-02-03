# AI-WAF MCP 服务器工具清单 - 50个工具完整列表

## 📋 按功能分类（共50个工具）

### 🔍 日志查询工具 (2个)

| # | 工具名 | API端点 | 方法 | 输入参数 | 权限 | 幂等 |
|---|--------|---------|------|---------|------|------|
| 1 | list_attack_logs | `/api/v1/waf/logs` | GET | page, size, filter (time_range, attack_type, severity, blocked) | 读 | ✅ |
| 2 | get_log_stats | `/api/v1/waf/logs/events` | GET | time_range, metric_type | 读 | ✅ |

---

### 📋 规则管理工具 (8个)

| # | 工具名 | API端点 | 方法 | 备注 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 3 | list_micro_rules | `/api/v1/micro-rules` | GET | page, size, filter, sort | 读 | ✅ |
| 4 | create_micro_rule | `/api/v1/micro-rules` | POST | **name, type, status, priority, condition** ✅已修复 | 写 | ❌ |
| 5 | update_micro_rule | `/api/v1/micro-rules/{id}` | PUT | name, type, status, priority, condition (可选) | 写 | ❌ |
| 6 | delete_micro_rule | `/api/v1/micro-rules/{id}` | DELETE | ruleId, force (可选) | 写 | ❌ |
| 7 | export_rules | `/api/v1/micro-rules` | GET | format (json/yaml), filter, include_disabled | 读 | ✅ |
| 8 | import_rules | `/api/v1/micro-rules` | POST | rules_content, format, merge_mode, dry_run | 写 | ❌ |
| 9 | batch_update_rules | `/api/v1/micro-rules/{id}` | PUT | rule_ids[], updates{}, rollback | 写 | ❌ |
| 10 | test_rule | 本地 | - | rule_id/rule_config, test_cases[] | 读 | ✅ |

---

### 🚫 IP封禁管理工具 (2个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 11 | list_blocked_ips | `/api/v1/blocked-ips` | GET | page, size, filter (severity, tags, created_after) | 读 | ✅ |
| 12 | get_blocked_ip_stats | `/api/v1/blocked-ips/stats` | GET | 无参数 | 读 | ✅ |

---

### 🏢 站点管理工具 (2个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 13 | list_sites | `/api/v1/site` | GET | page, size, filter | 读 | ✅ |
| 14 | get_site_details | `/api/v1/site/{id}` | GET | site_id | 读 | ✅ |

---

### ⚙️ 配置管理工具 (3个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 15 | get_waf_config | `/api/v1/config` | GET | 无参数 | 读 | ✅ |
| 16 | update_waf_config | `/api/v1/config` | PATCH | config_updates{} | 写 | ❌ |
| 17 | get_stats_overview | `/api/v1/stats/overview` | GET | 无参数 | 读 | ✅ |

---

### 📦 批量操作工具 (4个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 18 | batch_block_ips | `/api/v1/blocked-ips` | POST | ips[], reason, duration, severity | 写 | ❌ |
| 19 | batch_unblock_ips | `/api/v1/blocked-ips/{ip}` | DELETE | ips[], reason | 写 | ❌ |
| 20 | batch_create_rules | `/api/v1/micro-rules` | POST | rules[] (RuleCreateRequest[]) | 写 | ❌ |
| 21 | batch_delete_rules | `/api/v1/micro-rules/{id}` | DELETE | rule_ids[], force | 写 | ❌ |

---

### 📊 实时监控工具 (4个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 22 | get_realtime_qps | `/api/v1/stats/realtime-qps` | GET | 无参数 | 读 | ✅ |
| 23 | get_time_series_data | `/api/v1/stats/time-series` | GET | metric_type, time_range, granularity | 读 | ✅ |
| 24 | get_security_metrics | `/api/v1/stats/security-metrics` | GET | 无参数 | 读 | ✅ |
| 25 | get_system_health | `/health` | GET | 无参数 | 读 | ✅ |

---

### 🤖 AI分析工具 (5个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 26 | list_attack_patterns | `/api/v1/ai-analyzer/patterns` | GET | page, size, filter | 读 | ✅ |
| 27 | list_generated_rules | `/api/v1/ai-analyzer/rules` | GET | page, size, status | 读 | ✅ |
| 28 | trigger_ai_analysis | `/api/v1/ai-analyzer/trigger` | POST | 无参数 | 写 | ❌ |
| 29 | review_rule | `/api/v1/ai-analyzer/rules/review` | POST | rule_id, action, comment | 写 | ❌ |
| 30 | deploy_rule | `/api/v1/ai-analyzer/rules/{id}/deploy` | POST | rule_id | 写 | ❌ |

---

### 🧠 高级AI分析工具 (5个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 31 | analyze_attack_patterns | `/api/v1/ai-analyzer/patterns/analyze` | POST | time_range, clustering_method, min_samples | 读 | ✅ |
| 32 | generate_rule_from_pattern | `/api/v1/ai-analyzer/rules` | POST | pattern_id, action, priority, auto_review | 写 | ❌ |
| 33 | evaluate_rule_effectiveness | `/api/v1/rules/effectiveness/{id}` | GET | rule_id, time_range | 读 | ✅ |
| 34 | optimize_rule | 本地ML | - | rule_id/rule_config, optimization_strategy | 读 | ✅ |
| 35 | compare_rules | 本地ML | - | rule_id1, rule_id2, metrics[] | 读 | ✅ |

---

### 📈 扩展工具 - 报告与分析 (4个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 36 | generate_security_report | `/api/v1/waf/logs` | GET | time_range, format (json/html/pdf) | 读 | ✅ |
| 37 | aggregate_metrics | `/api/v1/stats/` | GET | dimensions[], time_range, group_by | 读 | ✅ |
| 38 | export_audit_log | 审计日志数据库 | - | start_date, end_date, format, operations[] | 审计 | ✅ |
| 39 | smart_rule_suggestion | AI分析 | - | analyze_period, min_incidents, fp_threshold | 读 | ✅ |

---

### 🔮 扩展工具 - 预测与自动响应 (3个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 40 | predict_threats | ML模型 | - | prediction_window, confidence_threshold | 读 | ✅ |
| 41 | auto_remediate | `/api/v1/` | POST | threat_pattern_id, action, scope, rollback | 写 | ❌ |
| 42 | [预留] auto_scaling | 容量管理 | - | 预留给自动扩容 | 写 | ❌ |

---

### 🔔 扩展工具 - 告警与事件 (2个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 43 | setup_alert_policy | 告警系统 | POST | policy_name, conditions[], actions[], aggregation | 写 | ❌ |
| 44 | get_incident_status | 事件系统 | GET | status_filter, severity_filter, limit | 读 | ✅ |

---

### 📜 扩展工具 - 合规与治理 (2个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 45 | compliance_check | 合规引擎 | - | standards[] (OWASP/PCI-DSS/ISO27001) | 读 | ✅ |
| 46 | audit_trail_validation | 审计系统 | - | date_range, check_integrity, generate_cert | 审计 | ✅ |

---

### 📊 扩展工具 - 容量规划 (1个)

| # | 工具名 | API端点 | 方法 | 功能 | 权限 | 幂等 |
|---|--------|---------|------|------|------|------|
| 47 | capacity_planning | 容量系统 | - | growth_projection, redundancy_factor, scenario | 读 | ✅ |

---

### 🔮 预留扩展工具 (3个)

| # | 工具名 | 功能 | 优先级 |
|---|--------|------|--------|
| 48 | [预留] anomaly_detection | 实时异常检测 | 中 |
| 49 | [预留] performance_tuning | 性能自动优化 | 中 |
| 50 | [预留] threat_intelligence | 威胁情报集成 | 低 |

---

## 🎯 工具使用权限矩阵

### 权限类型
- **读权限** (Read): 仅查询数据，不修改系统状态
- **写权限** (Write): 修改规则、配置、IP黑名单等关键数据
- **审计权限** (Audit): 导出和验证审计日志，用于合规审计

### 幂等性说明
- ✅ **幂等**: 多次调用结果相同（GET请求）
- ❌ **非幂等**: 每次调用可能产生不同结果（POST/DELETE请求）

---

## 📊 覆盖对标

### OWASP Top 10 防护
- A01 Broken Access Control → 规则管理、IP黑名单
- A02 Cryptographic Failures → 加密配置检查
- A03 Injection → 规则库（SQL/命令注入）
- A04 Insecure Design → 合规检查
- ...其他6项

### PCI-DSS 3.2.1 覆盖
- 6.6 Web Application Firewall → 所有规则工具
- 8.2 User Identification → 审计日志导出
- 10.2 Automated Audit Trails → audit_trail_validation

### ISO 27001:2022 覆盖
- A.12.4.1 Event logging → 日志查询、审计导出
- A.12.4.3 Protection of log information → 审计验证
- A.13.1.1 Network architecture → 容量规划

---

## 🔗 实现文件映射

| 工具类别 | 实现文件 | 工具数 |
|---------|---------|--------|
| 日志查询 | logs.go | 2 |
| 规则管理 | rules.go + rules_advanced.go | 8 |
| IP封禁 | blocked_ips.go | 2 |
| 站点管理 | sites.go | 2 |
| 配置管理 | config.go | 3 |
| 批量操作 | batch_operations.go | 4 |
| 监控统计 | monitoring.go | 4 |
| AI分析 | ai_analyzer.go + ai_analysis_advanced.go | 10 |
| 扩展工具 | extended_tools.go | 14 |
| **总计** | **9个Go文件** | **50个工具** |

---

## 🚀 快速参考

### 查询类工具（读权限，幂等）
```
list_attack_logs        get_log_stats
list_micro_rules        list_sites
get_site_details        get_waf_config
get_realtime_qps        get_time_series_data
get_security_metrics    list_attack_patterns
list_generated_rules    generate_security_report
aggregate_metrics       smart_rule_suggestion
predict_threats         get_incident_status
compliance_check        capacity_planning
```

### 修改类工具（写权限，非幂等）
```
create_micro_rule       update_micro_rule
delete_micro_rule       import_rules
batch_update_rules      batch_block_ips
batch_unblock_ips       batch_create_rules
batch_delete_rules      trigger_ai_analysis
review_rule             deploy_rule
generate_rule_from_pattern
auto_remediate          setup_alert_policy
```

### 本地处理工具（无后端API调用）
```
test_rule               optimize_rule
compare_rules           analyze_attack_patterns
smart_rule_suggestion   predict_threats
```

---

## 📝 版本历史

- **v1.0** (2026-01-15): 初版31个核心工具
- **v2.0** (2026-02-02): 
  - 修复规则创建参数
  - 添加12个高级规则工具
  - 添加14个扩展工具
  - **总计50个工具，完整覆盖企业级WAF运维需求**
