# Go MCP Server 开发最佳实践

基于 [Model Context Protocol](https://modelcontextprotocol.io/) 和 MCP Builder Skill 的 Go 语言实现指南。

## 🎯 核心原则

### 1. 工具设计哲学

**API 覆盖度 vs 工作流工具**
- 优先考虑全面的 API 端点覆盖
- 为特定任务提供专门的工作流工具
- 平衡灵活性和便利性

**工具命名和可发现性**
- 使用一致的前缀（例如：`ai_waf_list_logs`, `ai_waf_create_rule`）
- 采用动作导向的命名（list、create、update、delete、get）
- 清晰描述工具功能，避免歧义

**上下文管理**
- 工具描述简洁明了
- 支持结果过滤和分页
- 返回聚焦、相关的数据

**可操作的错误消息**
- 提供具体的解决建议
- 包含下一步操作指导
- 区分客户端错误和服务器错误

---

## 🏗️ 项目结构

```
mcp-server/
├── main.go                      # MCP Server 入口
├── go.mod                       # Go 模块定义
├── tools/                       # 工具实现
│   ├── client.go               # API 客户端封装
│   ├── types.go                # 共享类型定义
│   ├── helpers.go              # 工具辅助函数
│   ├── logs.go                 # 日志相关工具
│   ├── rules.go                # 规则管理工具
│   └── ...                     # 其他功能模块
├── middleware.go               # 中间件（日志、跟踪）
└── README.md                   # 用户文档
```

---

## 🔧 实现模式

### 1. 工具注册

```go
// main.go
import (
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "ai-waf",
        Version: "v1.0.0",
    }, nil)

    // 注册工具，使用描述性的 Tool 结构
    mcp.AddTool(server, &mcp.Tool{
        Name:        "ai_waf_list_logs",
        Description: "Query WAF attack logs with filtering and pagination",
        Annotations: map[string]interface{}{
            "readOnlyHint": true,
            "idempotentHint": true,
        },
    }, tools.CreateListLogs(client))

    // 启动 stdio transport
    server.Run(context.Background(), &mcp.StdioTransport{})
}
```

### 2. 工具实现模式

```go
// tools/logs.go

// 输入结构 - 使用 JSON Schema 标签
type ListLogsInput struct {
    Hours      int    `json:"hours,omitempty" jsonschema:"description=查询最近N小时的日志,minimum=1,maximum=168,default=1"`
    AttackType string `json:"attack_type,omitempty" jsonschema:"description=攻击类型过滤,enum=sql_injection|xss|path_traversal"`
    Limit      int    `json:"limit,omitempty" jsonschema:"description=返回结果数量限制,minimum=1,maximum=1000,default=100"`
    Offset     int    `json:"offset,omitempty" jsonschema:"description=分页偏移量,minimum=0,default=0"`
}

// 输出结构 - 结构化数据
type ListLogsOutput struct {
    Logs       []AttackLog `json:"logs"`
    Total      int         `json:"total"`
    HasMore    bool        `json:"has_more"`
    Summary    LogSummary  `json:"summary"`
}

type AttackLog struct {
    ID         string    `json:"id"`
    Timestamp  time.Time `json:"timestamp"`
    SourceIP   string    `json:"source_ip"`
    AttackType string    `json:"attack_type"`
    Severity   string    `json:"severity"`
    Blocked    bool      `json:"blocked"`
}

// 工具函数 - 返回结构化结果
func CreateListLogs(client *APIClient) func(context.Context, *mcp.CallToolRequest, ListLogsInput) (*mcp.CallToolResult, ListLogsOutput, error) {
    return func(ctx context.Context, req *mcp.CallToolRequest, input ListLogsInput) (*mcp.CallToolResult, ListLogsOutput, error) {
        // 1. 参数验证
        if input.Hours < 1 || input.Hours > 168 {
            return nil, ListLogsOutput{}, fmt.Errorf("hours must be between 1 and 168")
        }

        // 2. API 调用
        data, err := client.GetWithContext(ctx, fmt.Sprintf("/api/logs?hours=%d", input.Hours))
        if err != nil {
            return nil, ListLogsOutput{}, formatError("Failed to fetch logs", err)
        }

        // 3. 解析响应
        var logs []AttackLog
        if err := json.Unmarshal(data, &logs); err != nil {
            return nil, ListLogsOutput{}, formatError("Failed to parse logs", err)
        }

        // 4. 构建输出
        output := ListLogsOutput{
            Logs:    logs,
            Total:   len(logs),
            HasMore: len(logs) >= input.Limit,
        }

        return nil, output, nil
    }
}
```

### 3. 错误处理

```go
// tools/helpers.go

// 格式化用户友好的错误消息
func formatError(context string, err error) error {
    switch {
    case isAuthError(err):
        return fmt.Errorf("%s: Authentication failed. Please check your WAF_API_TOKEN environment variable and ensure it's a valid JWT token", context)
    
    case isNotFoundError(err):
        return fmt.Errorf("%s: Resource not found. Verify the ID exists using list commands", context)
    
    case isValidationError(err):
        return fmt.Errorf("%s: Invalid input parameters. %v", context, err)
    
    case isNetworkError(err):
        return fmt.Errorf("%s: Network error. Check if WAF backend is accessible at WAF_BACKEND_URL", context)
    
    default:
        return fmt.Errorf("%s: %v", context, err)
    }
}

func isAuthError(err error) bool {
    return strings.Contains(err.Error(), "401") || 
           strings.Contains(err.Error(), "403")
}

func isNotFoundError(err error) bool {
    return strings.Contains(err.Error(), "404")
}

func isValidationError(err error) bool {
    return strings.Contains(err.Error(), "400")
}

func isNetworkError(err error) bool {
    return strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "timeout")
}
```

### 4. API 客户端

```go
// tools/client.go

type APIClient struct {
    BaseURL    string
    Token      string
    HTTPClient *http.Client
}

func NewAPIClient(baseURL, token string) *APIClient {
    return &APIClient{
        BaseURL: baseURL,
        Token:   token,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
    }
}

// 带 context 的请求方法
func (c *APIClient) GetWithContext(ctx context.Context, path string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+path, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
    }
    
    return body, nil
}
```

---

## 📊 工具注解

使用工具注解帮助 AI 理解工具特性：

```go
mcp.AddTool(server, &mcp.Tool{
    Name:        "ai_waf_list_logs",
    Description: "Query WAF attack logs with filtering and pagination",
    Annotations: map[string]interface{}{
        "readOnlyHint":   true,  // 只读操作
        "idempotentHint": true,  // 幂等操作
    },
}, tools.CreateListLogs(client))

mcp.AddTool(server, &mcp.Tool{
    Name:        "ai_waf_delete_rule",
    Description: "Delete a WAF rule by ID",
    Annotations: map[string]interface{}{
        "destructiveHint": true, // 破坏性操作
        "idempotentHint":  true, // 幂等操作
    },
}, tools.CreateDeleteRule(client))
```

**注解说明：**
- `readOnlyHint`: 工具不修改任何数据
- `destructiveHint`: 工具会删除或修改重要数据
- `idempotentHint`: 多次调用产生相同结果
- `openWorldHint`: 工具可能访问外部/动态资源

---

## 🧪 测试策略

### 1. 单元测试

```go
// tools/logs_test.go

func TestListLogs(t *testing.T) {
    // 创建模拟客户端
    client := &MockAPIClient{
        Response: `[{"id":"1","attack_type":"sql_injection"}]`,
    }
    
    // 创建工具函数
    handler := CreateListLogs(client)
    
    // 测试调用
    result, output, err := handler(
        context.Background(),
        &mcp.CallToolRequest{},
        ListLogsInput{Hours: 1},
    )
    
    assert.NoError(t, err)
    assert.Equal(t, 1, output.Total)
}
```

### 2. MCP Inspector 测试

```bash
# 使用 MCP Inspector 测试工具
npx @modelcontextprotocol/inspector go run main.go
```

---

## 🎨 响应格式

### JSON 响应（推荐用于结构化数据）

```go
output := ListLogsOutput{
    Logs: []AttackLog{
        {ID: "1", AttackType: "sql_injection", Severity: "high"},
    },
    Total: 1,
    Summary: LogSummary{
        TotalAttacks: 137,
        ByType: map[string]int{
            "sql_injection": 45,
            "xss": 23,
        },
    },
}
return nil, output, nil
```

### Markdown 响应（适合报告和摘要）

```go
content := []mcp.Content{
    {
        Type: "text",
        Text: fmt.Sprintf(`# WAF Attack Report

**Time Range:** Last %d hours
**Total Attacks:** %d

## Attack Breakdown
- SQL Injection: %d (%.1f%%)
- XSS: %d (%.1f%%)
- Path Traversal: %d (%.1f%%)

## Top Source IPs
1. %s - %d attacks
2. %s - %d attacks
`, hours, total, sqlCount, sqlPercent, xssCount, xssPercent, ...),
    },
}

return &mcp.CallToolResult{
    Content: content,
}, ListLogsOutput{...}, nil
```

---

## 🔐 安全最佳实践

1. **认证**
   - 始终使用环境变量存储敏感信息
   - 验证 JWT token 格式
   - 不在日志中输出完整 token

2. **输入验证**
   - 验证所有用户输入
   - 使用 JSON Schema 约束
   - 防止注入攻击

3. **错误处理**
   - 不暴露内部实现细节
   - 提供安全的错误消息
   - 记录详细错误到服务端日志

---

## 📈 性能优化

1. **连接池**
   ```go
   Transport: &http.Transport{
       MaxIdleConns:        100,
       MaxIdleConnsPerHost: 10,
       IdleConnTimeout:     90 * time.Second,
   }
   ```

2. **超时控制**
   ```go
   ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
   defer cancel()
   ```

3. **分页支持**
   - 默认限制返回数量
   - 提供 offset/limit 参数
   - 返回 has_more 标志

---

## 📚 参考资源

- [MCP Protocol Specification](https://modelcontextprotocol.io/specification/draft)
- [Go SDK Documentation](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Best Practices](https://modelcontextprotocol.io/docs/best-practices)

---

## ✅ 检查清单

在发布 MCP Server 前检查：

- [ ] 所有工具使用一致的命名前缀
- [ ] 工具描述清晰、简洁
- [ ] 输入参数有完整的 JSON Schema 约束
- [ ] 错误消息提供可操作的建议
- [ ] 工具有适当的注解（readOnly/destructive/idempotent）
- [ ] 支持分页的工具实现了 limit/offset
- [ ] 所有 HTTP 请求使用 context 控制超时
- [ ] 敏感信息通过环境变量配置
- [ ] 日志输出到 stderr（不干扰 JSON-RPC）
- [ ] 使用 MCP Inspector 测试过所有工具
- [ ] 创建了评估问题集
