# WAF MCP 工具快速参考卡

## 🔒 IP黑名单管理 (6个工具)

### waf_block_ip
添加单个IP到黑名单
```json
{
  "ip_address": "192.168.1.1",
  "reason": "Detected attack",
  "duration_seconds": 3600,  // 0=永久
  "severity": "high"
}
```

### waf_batch_block_ips
批量添加IP（最多1000条）
```json
{
  "ips": ["192.168.1.1", "10.0.0.1"],
  "reason": "Botnet",
  "duration_seconds": 86400
}
```

### waf_unblock_ip
移除单个IP
```json
{
  "ip_address": "192.168.1.1",
  "reason": "False positive"
}
```

### waf_batch_unblock_ips
批量移除IP
```json
{
  "ips": ["192.168.1.1", "10.0.0.1"],
  "reason": "Whitelist"
}
```

### waf_list_blocked_ips
查询黑名单（分页）
```json
{
  "page": 1,
  "page_size": 20,
  "severity_filter": "high",
  "sort_by": "created_at"
}
```

### waf_get_blocked_ip_stats
获取黑名单统计
```json
{
  // 无参数，返回:
  // - total_blocked_count
  // - active_blocks_count
  // - stats_by_severity
  // - recent_blocks
}
```

---

## 📋 规则管理 (9个工具)

### waf_create_rule
创建单条防护规则
```json
{
  "rule_name": "Block SQL Injection",
  "rule_type": "pattern",
  "action": "block",
  "priority": 100,
  "path": "/api/*",
  "method": ["POST"],
  "enabled": true
}
```

**rule_type选项**: blacklist, whitelist, pattern, rate_limit, geo_block
**action选项**: block, allow, log, challenge

### waf_batch_create_rules
批量创建规则（最多100条）
```json
{
  "rules": [
    {
      "rule_name": "Rule1",
      "rule_type": "pattern",
      "action": "block",
      "priority": 100
    },
    // ... more rules
  ]
}
```

### waf_update_rule
更新规则
```json
{
  "rule_id": "rule_12345",
  "action": "block",
  "priority": 200,
  "enabled": false
}
```

### waf_delete_rule
删除单条规则
```json
{
  "rule_id": "rule_12345",
  "force": false
}
```

### waf_batch_delete_rules
批量删除规则
```json
{
  "rule_ids": ["rule_1", "rule_2"],
  "force": false
}
```

### waf_list_rules
查询规则列表
```json
{
  "page": 1,
  "page_size": 20,
  "rule_type": "pattern",
  "enabled_only": true,
  "sort_by": "priority"
}
```

### waf_get_rule_details
获取规则详情
```json
{
  "rule_id": "rule_12345"
}
```

### waf_export_rules
导出规则
```json
{
  "format": "json",  // 或 yaml
  "include_disabled": true
}
```

### waf_import_rules
导入规则
```json
{
  "rules_content": "{ JSON内容 }",
  "format": "json",
  "merge_mode": "merge"  // merge, replace, skip_duplicate
}
```

---

## 🔍 攻击分析 (5个工具)

### waf_list_attack_logs
查询攻击日志
```json
{
  "page": 1,
  "page_size": 50,
  "attack_type": "sql_injection",
  "severity_filter": "high",
  "source_ip": "192.168.1.1",
  "hours": 24
}
```

**attack_type选项**: sql_injection, xss, path_traversal, cmd_injection, xxe, csrf

### waf_analyze_attack_patterns
分析攻击模式（AI驱动）
```json
{
  "hours": 24,
  "clustering_method": "kmeans",  // 或 dbscan
  "min_samples": 10,
  "anomaly_threshold": 2.0
}
```

返回: patterns, anomalies, recommendations, suggested_rules

### waf_generate_rule_from_pattern
根据模式生成规则
```json
{
  "pattern_id": "pattern_123",
  "action": "block",
  "priority": 100,
  "auto_review": false
}
```

### waf_review_generated_rule
审核AI生成的规则
```json
{
  "rule_id": "rule_123",
  "action": "approve",  // 或 reject
  "comment": "Looks good"
}
```

### waf_deploy_generated_rule
部署已审核的规则
```json
{
  "rule_id": "rule_123",
  "test_first": true,
  "deployment_strategy": "gradual"  // immediate, gradual, scheduled
}
```

---

## 📊 系统监控 (5个工具)

### waf_get_stats_overview
获取统计概览
```json
{
  "time_range": "24h"  // 1h, 6h, 24h, 7d, 30d
}
```

返回: total_requests, blocked_requests, block_rate, top_attack_types

### waf_get_time_series_data
获取时间序列数据
```json
{
  "metric_type": "requests",  // errors, response_time, blocked_rate
  "time_range": "24h",
  "interval": "1h"  // 自动推荐
}
```

### waf_get_security_metrics
获取安全指标
```json
{
  "time_range": "7d"
}
```

返回: threat_level, detected_threats, prevented_breaches, vulnerability_score

### waf_get_realtime_qps
实时QPS监控
```json
{
  "limit": 30  // 最多60
}
```

返回: current_qps, peak_qps, average_qps, data_points

### waf_compare_rules
对比两条规则
```json
{
  "rule_id_1": "rule_1",
  "rule_id_2": "rule_2",
  "time_range": "7d"
}
```

---

## ⚙️ 规则优化 (4个工具)

### waf_evaluate_rule_effectiveness
评估规则有效性
```json
{
  "rule_id": "rule_123",
  "time_range": "24h"
}
```

返回: true_positive_rate, false_positive_rate, block_rate, improvement_suggestions

### waf_optimize_rule
自动优化规则
```json
{
  "rule_id": "rule_123",
  "optimize_for": "accuracy",  // performance, both
  "keep_history": true
}
```

### waf_test_rule
测试规则
```json
{
  "rule_id": "rule_123",
  "test_payload": "SELECT * FROM users",
  "expected_action": "block"
}
```

### waf_get_rule_logs
获取规则执行日志
```json
{
  "rule_id": "rule_123",
  "action": "triggered",  // passed, error
  "limit": 100
}
```

---

## 🔧 配置管理 (3个工具)

### waf_get_config
获取WAF配置
```json
{
  // 无参数
}
```

返回: 完整的WAF配置对象

### waf_update_config
更新WAF配置
```json
{
  "enable_ai_analysis": true,
  "enable_rate_limit": true,
  "rate_limit_rpm": 1000,
  "max_request_body_size": 10485760,
  "log_level": "info"
}
```

### waf_get_system_health
获取系统健康状态
```json
{
  // 无参数
}
```

返回: status, uptime, cpu_usage, memory_usage, services_status

---

## 🎯 常见使用场景

### 场景1: 应急响应
```
1. waf_analyze_attack_patterns(hours=1)
   └─> 识别最近的攻击

2. waf_batch_block_ips(ips=[...])
   └─> 封禁恶意IP

3. waf_generate_rule_from_pattern(pattern_id=X)
   └─> 生成防护规则

4. waf_deploy_generated_rule(rule_id=Y)
   └─> 部署规则
```

### 场景2: 规则优化
```
1. waf_list_rules(rule_type="pattern")
   └─> 找出性能问题规则

2. waf_evaluate_rule_effectiveness(rule_id=X)
   └─> 评估规则效果

3. waf_optimize_rule(rule_id=X, optimize_for="accuracy")
   └─> 优化规则参数

4. waf_test_rule(rule_id=X, test_payload="...")
   └─> 验证优化效果
```

### 场景3: 安全审计
```
1. waf_get_stats_overview(time_range="7d")
   └─> 获取周报数据

2. waf_list_attack_logs(hours=168)
   └─> 查询一周的攻击

3. waf_analyze_attack_patterns(hours=168)
   └─> 分析攻击趋势

4. waf_get_security_metrics(time_range="7d")
   └─> 生成安全评分
```

---

## ⏱️ 参数速查表

| 参数 | 类型 | 范围 | 示例 |
|------|------|------|------|
| page | int | ≥1 | 1 |
| page_size | int | 1-100 | 20 |
| priority | int | 1-10000 | 100 |
| duration_seconds | int | ≥0 | 3600 |
| severity | enum | critical\|high\|medium\|low | "high" |
| rule_type | enum | blacklist\|whitelist\|pattern\|rate_limit\|geo_block | "pattern" |
| action | enum | block\|allow\|log\|challenge | "block" |
| time_range | enum | 1h\|6h\|24h\|7d\|30d | "24h" |
| hours | int | 1-8760 | 24 |

---

## 📝 响应格式

### 成功响应
```json
{
  "success": true,
  "data": { /* 具体数据 */ },
  "message": "Operation completed successfully",
  "timestamp": "2024-02-02T10:30:00Z",
  "request_id": "req_12345"
}
```

### 错误响应
```json
{
  "success": false,
  "error": {
    "code": "WAF_BLOCK_IP_FAILED",
    "message": "Error description",
    "details": { /* 错误详情 */ }
  },
  "timestamp": "2024-02-02T10:30:00Z",
  "request_id": "req_12345"
}
```

---

## 🚨 错误代码

| 错误代码 | 说明 | 处理建议 |
|---------|------|---------|
| WAF_BLOCK_IP_FAILED | IP封禁失败 | 检查IP格式，重试或使用规则方案 |
| WAF_RULE_CREATE_FAILED | 规则创建失败 | 验证规则参数，检查重复名称 |
| WAF_INVALID_PARAMETER | 参数验证失败 | 检查参数类型和范围 |
| WAF_RULE_NOT_FOUND | 规则不存在 | 检查rule_id是否正确 |
| WAF_PERMISSION_DENIED | 权限不足 | 检查API Key和权限 |
| WAF_OPERATION_TIMEOUT | 操作超时 | 增加timeout配置，重试 |
| WAF_BACKEND_ERROR | 后端错误 | 检查WAF服务状态 |
| WAF_CONFLICT | 冲突（如IP已存在） | 检查是否重复操作 |

---

## 💻 代码示例

### Python调用示例
```python
import httpx

async with httpx.AsyncClient() as client:
    response = await client.post(
        "http://localhost:8000/waf_block_ip",
        json={
            "ip_address": "192.168.1.1",
            "reason": "SQL injection detected",
            "duration_seconds": 86400,
            "severity": "high"
        }
    )
    result = response.json()
    print(f"Success: {result['success']}")
```

### JavaScript/Node.js调用示例
```javascript
const response = await fetch('http://localhost:8000/waf_block_ip', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    ip_address: '192.168.1.1',
    reason: 'SQL injection detected',
    duration_seconds: 86400,
    severity: 'high'
  })
});
const result = await response.json();
console.log(result.success);
```

### cURL示例
```bash
curl -X POST http://localhost:8000/waf_block_ip \
  -H "Content-Type: application/json" \
  -d '{
    "ip_address": "192.168.1.1",
    "reason": "SQL injection detected",
    "duration_seconds": 86400,
    "severity": "high"
  }'
```

---

## 🔗 相关文档

- **完整设计**: WAF_MCP_Design.md
- **实现代码**: waf_mcp_server.py
- **API客户端**: waf_api_client.py
- **部署指南**: WAF_MCP_Testing_Deployment.md

---

**更新日期**: 2024年2月2日
**版本**: 1.0.0
