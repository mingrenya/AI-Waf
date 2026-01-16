## MCP 集成后端 API 端点

为了支持前端的 MCP 功能，需要在后端实现以下 API 端点：

### 1. MCP 连接状态

**GET /api/v1/mcp/status**

返回 MCP 服务器的连接状态。

响应示例：
```json
{
  "data": {
    "connected": true,
    "lastConnectedAt": "2026-01-15T12:00:00Z",
    "serverVersion": "v1.0.0",
    "totalTools": 31,
    "availableTools": [
      "list_attack_logs",
      "get_log_stats",
      "list_micro_rules",
      ...
    ]
  }
}
```

### 2. MCP 工具列表

**GET /api/v1/mcp/tools**

返回所有可用的 MCP 工具列表。

响应示例：
```json
{
  "data": {
    "tools": [
      "list_attack_logs",
      "get_log_stats",
      "list_micro_rules",
      "create_micro_rule",
      ...
    ]
  }
}
```

### 3. MCP 工具调用历史

**GET /api/v1/mcp/tool-calls**

查询参数：
- `limit` (int): 返回记录数量限制，默认 50
- `offset` (int): 偏移量，默认 0

返回 MCP 工具的调用历史记录。

响应示例：
```json
{
  "data": {
    "data": [
      {
        "id": "call_123",
        "toolName": "analyze_attack_patterns",
        "timestamp": "2026-01-15T12:00:00Z",
        "duration": 1250,
        "success": true
      }
    ],
    "total": 150
  }
}
```

### 4. AI 规则建议列表

**GET /api/v1/ai-analyzer/suggestions**

查询参数：
- `status` (string): 筛选状态 - pending, approved, rejected, deployed
- `severity` (string): 筛选严重程度 - low, medium, high, critical
- `limit` (int): 返回记录数量限制，默认 20
- `offset` (int): 偏移量，默认 0

返回 AI 生成的规则建议列表。

响应示例：
```json
{
  "data": {
    "data": [
      {
        "id": "sugg_001",
        "patternId": "pattern_123",
        "patternName": "高频 SQL 注入攻击",
        "ruleName": "阻止 UNION SELECT 注入",
        "ruleType": "micro_rule",
        "confidence": 0.92,
        "severity": "high",
        "description": "检测到大量针对用户表的 UNION SELECT 注入尝试",
        "recommendation": "建议立即部署此规则以阻止该攻击模式",
        "ruleContent": {
          "conditions": [...],
          "action": "block"
        },
        "status": "pending",
        "createdAt": "2026-01-15T10:30:00Z"
      }
    ],
    "total": 12
  }
}
```

### 5. 批准规则建议

**POST /api/v1/ai-analyzer/suggestions/:id/approve**

批准指定的规则建议。

响应示例：
```json
{
  "message": "规则建议已批准",
  "data": {
    "id": "sugg_001",
    "status": "approved",
    "reviewedAt": "2026-01-15T12:10:00Z"
  }
}
```

### 6. 拒绝规则建议

**POST /api/v1/ai-analyzer/suggestions/:id/reject**

请求体：
```json
{
  "reason": "误报率过高"
}
```

拒绝指定的规则建议。

响应示例：
```json
{
  "message": "规则建议已拒绝",
  "data": {
    "id": "sugg_001",
    "status": "rejected",
    "reviewedAt": "2026-01-15T12:10:00Z"
  }
}
```

### 7. 部署规则建议

**POST /api/v1/ai-analyzer/suggestions/:id/deploy**

部署已批准的规则建议到生产环境。

响应示例：
```json
{
  "message": "规则已成功部署",
  "data": {
    "id": "sugg_001",
    "status": "deployed",
    "deployedAt": "2026-01-15T12:15:00Z",
    "ruleId": "rule_456"
  }
}
```

### 8. 获取 AI 分析结果

**GET /api/v1/ai-analyzer/analysis/result**

查询参数：
- `timeRange` (string): 时间范围 - 1h, 6h, 24h, 7d，默认 24h

返回最近的 AI 分析结果统计。

响应示例：
```json
{
  "data": {
    "totalPatterns": 15,
    "highSeverityPatterns": 5,
    "suggestedRules": 12,
    "processingTime": 3.5,
    "timestamp": "2026-01-15T12:00:00Z"
  }
}
```

### 9. 触发 AI 分析

**POST /api/v1/ai-analyzer/analyze/patterns**

请求体：
```json
{
  "timeRange": "24h",
  "minSamples": 10,
  "anomalyThreshold": 2.0,
  "clusteringMethod": "kmeans"
}
```

触发新的攻击模式分析任务。

响应示例：
```json
{
  "message": "分析任务已启动",
  "data": {
    "taskId": "task_789",
    "status": "running",
    "estimatedTime": 60
  }
}
```

## 实现建议

1. **MCP 状态管理**: 在后端维护一个 MCP 连接状态的缓存，定期（如每 10 秒）检查 MCP 服务器状态。

2. **工具调用日志**: 将所有通过 MCP 执行的工具调用记录到 MongoDB，便于审计和分析。

3. **规则建议存储**: 创建一个新的 MongoDB collection 存储 AI 生成的规则建议，支持状态转换（pending → approved → deployed）。

4. **权限控制**: 部署规则建议应该需要管理员权限，使用现有的 JWT 中间件进行鉴权。

5. **异步处理**: 对于耗时的 AI 分析任务，使用异步任务队列（如 goroutines + channels）处理，避免阻塞 HTTP 请求。

## 数据库 Schema

### ai_rule_suggestions Collection

```go
type AIRuleSuggestion struct {
    ID              primitive.ObjectID `bson:"_id,omitempty"`
    PatternID       string             `bson:"pattern_id,omitempty"`
    PatternName     string             `bson:"pattern_name"`
    RuleName        string             `bson:"rule_name"`
    RuleType        string             `bson:"rule_type"` // micro_rule, modsecurity
    Confidence      float64            `bson:"confidence"`
    Severity        string             `bson:"severity"` // low, medium, high, critical
    Description     string             `bson:"description"`
    Recommendation  string             `bson:"recommendation"`
    RuleContent     interface{}        `bson:"rule_content"`
    Status          string             `bson:"status"` // pending, approved, rejected, deployed
    CreatedAt       time.Time          `bson:"created_at"`
    ReviewedAt      *time.Time         `bson:"reviewed_at,omitempty"`
    DeployedAt      *time.Time         `bson:"deployed_at,omitempty"`
    ReviewedBy      string             `bson:"reviewed_by,omitempty"`
    DeployedRuleID  string             `bson:"deployed_rule_id,omitempty"`
}
```

### mcp_tool_calls Collection

```go
type MCPToolCall struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    ToolName  string             `bson:"tool_name"`
    Timestamp time.Time          `bson:"timestamp"`
    Duration  int64              `bson:"duration"` // milliseconds
    Success   bool               `bson:"success"`
    Error     string             `bson:"error,omitempty"`
    UserID    string             `bson:"user_id,omitempty"`
    RequestID string             `bson:"request_id,omitempty"`
}
```
