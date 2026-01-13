# 告警系统 API 使用指南

## 概述

MRYa WAF 告警系统提供完整的告警管理功能，支持多种通知渠道和灵活的规则配置。

## 快速开始

### 1. 创建告警渠道

#### Webhook 渠道
```bash
curl -X POST http://localhost:2333/api/v1/alerts/channels \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的 Webhook",
    "type": "webhook",
    "enabled": true,
    "config": {
      "url": "https://your-server.com/webhook",
      "method": "POST",
      "headers": {
        "X-API-Key": "your-api-key"
      },
      "timeout": 30
    }
  }'
```

#### Slack 渠道
```bash
curl -X POST http://localhost:2333/api/v1/alerts/channels \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Slack 通知",
    "type": "slack",
    "enabled": true,
    "config": {
      "webhookUrl": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "channel": "#security-alerts",
      "username": "WAF Alert Bot",
      "iconEmoji": ":shield:"
    }
  }'
```

#### 钉钉渠道
```bash
curl -X POST http://localhost:2333/api/v1/alerts/channels \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "钉钉群通知",
    "type": "dingtalk",
    "enabled": true,
    "config": {
      "webhookUrl": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN",
      "secret": "YOUR_SECRET",
      "isAtAll": false,
      "atMobiles": ["13800138000"]
    }
  }'
```

### 2. 测试告警渠道

```bash
curl -X POST http://localhost:2333/api/v1/alerts/channels/{channelId}/test \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "这是一条测试消息"
  }'
```

### 3. 创建告警规则

#### 高 QPS 告警
```bash
curl -X POST http://localhost:2333/api/v1/alerts/rules \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "高 QPS 告警",
    "description": "当 QPS 超过 1000 时触发告警",
    "enabled": true,
    "severity": "high",
    "logic": "AND",
    "cooldown": 5,
    "channels": ["channel_id_1", "channel_id_2"],
    "conditions": [
      {
        "metric": "qps",
        "operator": ">",
        "threshold": 1000,
        "duration": 2
      }
    ],
    "template": "⚠️ 高 QPS 告警\n当前 QPS: {{.qps}}\n拦截率: {{.block_rate}}%\n时间: {{.timestamp}}"
  }'
```

#### 攻击拦截告警
```bash
curl -X POST http://localhost:2333/api/v1/alerts/rules \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "攻击拦截告警",
    "description": "当拦截数量超过 100 时触发告警",
    "enabled": true,
    "severity": "critical",
    "logic": "OR",
    "cooldown": 10,
    "channels": ["channel_id_1"],
    "conditions": [
      {
        "metric": "attack_count",
        "operator": ">",
        "threshold": 100,
        "duration": 1
      },
      {
        "metric": "block_rate",
        "operator": ">",
        "threshold": 50,
        "duration": 5
      }
    ],
    "template": "🚨 检测到大量攻击\n拦截数量: {{.attack_count}}\n拦截率: {{.block_rate}}%\n4xx 错误率: {{.error_4xx_rate}}%"
  }'
```

### 4. 查询告警历史

```bash
# 获取所有告警历史
curl -X GET "http://localhost:2333/api/v1/alerts/history?page=1&pageSize=20" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 按规则 ID 查询
curl -X GET "http://localhost:2333/api/v1/alerts/history?ruleId=rule_id_here" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 按严重级别查询
curl -X GET "http://localhost:2333/api/v1/alerts/history?severity=critical" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 按时间范围查询
curl -X GET "http://localhost:2333/api/v1/alerts/history?startTime=2026-01-01T00:00:00Z&endTime=2026-01-13T23:59:59Z" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. 确认告警

```bash
curl -X POST http://localhost:2333/api/v1/alerts/history/{historyId}/acknowledge \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "comment": "已处理，误报"
  }'
```

### 6. 获取告警统计

```bash
curl -X GET "http://localhost:2333/api/v1/alerts/statistics?startTime=2026-01-01T00:00:00Z" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 支持的指标

告警规则可以基于以下指标：

| 指标 | 说明 | 单位 |
|-----|------|-----|
| `qps` | 每秒请求数 | 次/秒 |
| `block_rate` | 拦截率 | 百分比 |
| `error_rate` | 总错误率 | 百分比 |
| `error_4xx_rate` | 4xx 错误率 | 百分比 |
| `error_5xx_rate` | 5xx 错误率 | 百分比 |
| `attack_count` | 攻击拦截数量 | 次 |
| `traffic` | 总流量 | 字节 |

## 支持的运算符

| 运算符 | 说明 |
|-------|------|
| `>` | 大于 |
| `<` | 小于 |
| `>=` | 大于等于 |
| `<=` | 小于等于 |
| `==` | 等于 |
| `!=` | 不等于 |

## 严重级别

| 级别 | 说明 |
|-----|------|
| `low` | 低 |
| `medium` | 中 |
| `high` | 高 |
| `critical` | 严重 |

## 渠道类型

| 类型 | 说明 |
|-----|------|
| `webhook` | 通用 Webhook |
| `slack` | Slack |
| `discord` | Discord |
| `dingtalk` | 钉钉 |
| `wecom` | 企业微信 |

## 模板变量

告警消息模板支持以下变量：

| 变量 | 说明 |
|-----|------|
| `{{.qps}}` | 当前 QPS |
| `{{.block_rate}}` | 拦截率 |
| `{{.error_4xx_rate}}` | 4xx 错误率 |
| `{{.error_5xx_rate}}` | 5xx 错误率 |
| `{{.attack_count}}` | 攻击数量 |
| `{{.traffic}}` | 流量大小 |

## 权限要求

| 操作 | 所需权限 |
|-----|---------|
| 创建渠道 | `alert:channel:create` |
| 查看渠道 | `alert:channel:read` |
| 更新渠道 | `alert:channel:update` |
| 删除渠道 | `alert:channel:delete` |
| 创建规则 | `alert:rule:create` |
| 查看规则 | `alert:rule:read` |
| 更新规则 | `alert:rule:update` |
| 删除规则 | `alert:rule:delete` |
| 查看历史 | `alert:history:read` |

## 最佳实践

### 1. 合理设置冷却时间
避免告警风暴，建议设置 5-15 分钟的冷却时间。

### 2. 使用多条件组合
使用 AND/OR 逻辑组合多个条件，提高告警准确性。

### 3. 配置多个渠道
为重要告警配置多个通知渠道，确保不遗漏。

### 4. 定期审查告警历史
定期查看告警历史，优化告警规则和阈值。

### 5. 测试告警渠道
创建渠道后立即测试，确保配置正确。

## 故障排查

### 告警未发送
1. 检查告警规则是否启用
2. 检查告警渠道是否启用
3. 检查是否在冷却时间内
4. 查看告警历史中的错误信息

### 渠道测试失败
1. 检查 webhook URL 是否正确
2. 检查网络连接
3. 检查 token/secret 是否有效
4. 查看错误日志获取详细信息

## 完整示例

```bash
#!/bin/bash

# 设置变量
API_BASE="http://localhost:2333/api/v1"
TOKEN="YOUR_JWT_TOKEN"

# 1. 创建 Webhook 渠道
CHANNEL_RESPONSE=$(curl -s -X POST "$API_BASE/alerts/channels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "告警 Webhook",
    "type": "webhook",
    "enabled": true,
    "config": {
      "url": "https://your-server.com/webhook",
      "method": "POST"
    }
  }')

CHANNEL_ID=$(echo $CHANNEL_RESPONSE | jq -r '.data.id')
echo "创建渠道成功，ID: $CHANNEL_ID"

# 2. 测试渠道
curl -X POST "$API_BASE/alerts/channels/$CHANNEL_ID/test" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message": "测试消息"}'

# 3. 创建告警规则
RULE_RESPONSE=$(curl -s -X POST "$API_BASE/alerts/rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"高 QPS 告警\",
    \"description\": \"QPS 超过 1000 时触发\",
    \"enabled\": true,
    \"severity\": \"high\",
    \"logic\": \"AND\",
    \"cooldown\": 5,
    \"channels\": [\"$CHANNEL_ID\"],
    \"conditions\": [{
      \"metric\": \"qps\",
      \"operator\": \">\",
      \"threshold\": 1000,
      \"duration\": 2
    }],
    \"template\": \"⚠️ 高 QPS 告警\\n当前 QPS: {{.qps}}\"
  }")

RULE_ID=$(echo $RULE_RESPONSE | jq -r '.data.id')
echo "创建规则成功，ID: $RULE_ID"

# 4. 查询告警历史
curl -X GET "$API_BASE/alerts/history?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN" | jq

echo "告警系统配置完成！"
```

## 注意事项

1. **Token 安全**: 妥善保管 JWT Token，不要在代码中硬编码
2. **速率限制**: API 可能有速率限制，避免频繁请求
3. **数据保留**: 告警历史数据会定期清理，建议及时导出重要数据
4. **权限管理**: 按需分配告警管理权限，遵循最小权限原则
