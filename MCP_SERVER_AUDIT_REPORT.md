# MCP Server 实现审计报告

## 审计基准
**官方示例**: https://github.com/modelcontextprotocol/go-sdk/tree/main/examples/http/main.go

## 官方 HTTP 示例分析

### 服务端实现 (runServer)

```go
func runServer(url string) {
    // 1. 创建 MCP Server
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "time-server",
        Version: "1.0.0",
    }, nil)

    // 2. 添加中间件（可选）
    server.AddReceivingMiddleware(createLoggingMiddleware())

    // 3. 注册工具
    mcp.AddTool(server, &mcp.Tool{
        Name:        "cityTime",
        Description: "Get the current time in NYC, San Francisco, or Boston",
    }, getTime)

    // 4. 创建 StreamableHTTPHandler
    handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
        return server
    }, nil)

    // 5. 启动 HTTP 服务器
    log.Fatal(http.ListenAndServe(url, handler))
}
```

### 客户端实现 (runClient)

```go
func runClient(url string) {
    ctx := context.Background()

    // 1. 创建 MCP Client
    client := mcp.NewClient(&mcp.Implementation{
        Name:    "time-client",
        Version: "1.0.0",
    }, nil)

    // 2. 使用 StreamableClientTransport 连接
    session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
        Endpoint: url,
    }, nil)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer session.Close()

    // 3. 调用工具
    result, err := session.CallTool(ctx, &mcp.CallToolParams{
        Name:      "cityTime",
        Arguments: map[string]any{"city": "nyc"},
    })
}
```

## 本项目 MCP Server 实现分析

### 当前实现 (mcp-server/main.go)

```go
func main() {
    // 1. 获取环境变量配置 ✅
    backendURL := os.Getenv("WAF_BACKEND_URL")
    if backendURL == "" {
        backendURL = "http://localhost:2333"
    }
    apiToken := os.Getenv("WAF_API_TOKEN")

    // 2. 创建后端 API 客户端 ✅
    client := tools.NewAPIClient(backendURL, apiToken)

    // 3. 创建 MCP Server ✅
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "ai-waf",
        Version: "v1.0.0",
    }, nil)

    // 4. 注册 31 个工具 ✅
    mcp.AddTool(server, &mcp.Tool{
        Name:        "list_attack_logs",
        Description: "查询WAF攻击日志...",
    }, tools.CreateListAttackLogs(client))
    // ... 其他 30 个工具

    // 5. 使用 stdio 传输 ✅
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

## 对比结果

### ✅ 正确的部分

1. **Server 创建**: 使用 `mcp.NewServer()` 正确初始化
2. **工具注册**: 使用 `mcp.AddTool()` 正确注册工具
3. **Stdio 传输**: 使用 `&mcp.StdioTransport{}` 符合官方模式
4. **工具处理函数**: 返回 `(*mcp.CallToolResult, any, error)` 签名正确
5. **环境变量配置**: 正确读取配置并传递给工具

### 📋 与官方示例的差异

#### 1. 传输协议不同（正常）

**官方示例**: HTTP (`StreamableHTTPHandler`)
```go
handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
    return server
}, nil)
http.ListenAndServe(url, handler)
```

**本项目**: stdio (`StdioTransport`)
```go
server.Run(context.Background(), &mcp.StdioTransport{})
```

**结论**: ✅ 这是正常的差异，stdio 用于本地客户端（AnythingLLM/Claude Desktop），HTTP 用于网络客户端

#### 2. 缺少中间件（可选）

**官方示例**: 包含日志中间件
```go
server.AddReceivingMiddleware(createLoggingMiddleware())
```

**本项目**: 无中间件

**建议**: ⚠️ 可以添加中间件用于：
- 日志记录（调试）
- 工具调用追踪（记录到数据库）
- 性能监控

#### 3. 工具调用未记录到后端

**问题**: 工具函数调用后端 API，但没有记录调用历史到数据库

**影响**: 
- 后端 `/api/v1/mcp/tool-calls/history` 返回空数组
- 前端无法显示工具调用历史
- `checkMCPServerConnection()` 无法检测真实使用情况

## 架构说明

### 当前架构是正确的！

```
┌──────────────────────┐
│ AnythingLLM/Claude   │  MCP Client
└──────────┬───────────┘
           │ stdio (Standard Input/Output)
           ▼
┌──────────────────────┐
│ MCP Server           │  ai-waf-mcp (本地二进制)
│ (mcp-server/main.go) │  使用 stdio 传输
└──────────┬───────────┘
           │ HTTP API 调用
           ▼
┌──────────────────────┐
│ 后端 API (mrya)      │  http://localhost:2333
│ (server/main.go)     │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ MongoDB              │
└──────────────────────┘
```

**关键理解**:
- MCP Server 不是常驻服务，只在客户端需要时运行
- 后端无法直接检测 stdio MCP Server 是否在运行
- 前端和 MCP Server 是两个独立的通道访问后端 API

## 改进建议

### 建议 1: 添加工具调用追踪中间件（推荐）

在 `mcp-server/main.go` 中添加：

```go
// 在工具注册后，server.Run() 之前添加
server.AddReceivingMiddleware(func(next mcp.Receiver) mcp.Receiver {
    return mcp.ReceiverFunc(func(ctx context.Context, msg jsonrpc.Message) error {
        // 记录工具调用
        if req, ok := msg.(*jsonrpc.Request); ok && req.Method == "tools/call" {
            // 发送异步请求到后端记录调用
            go func() {
                toolName := extractToolName(req) // 从请求中提取工具名
                _ = client.Post("/api/v1/mcp/tool-calls/record", map[string]interface{}{
                    "toolName": toolName,
                    "timestamp": time.Now(),
                })
            }()
        }
        return next.Receive(ctx, msg)
    })
})
```

### 建议 2: 修改 checkMCPServerConnection 逻辑

**当前实现**:
```go
func (s *MCPService) checkMCPServerConnection() bool {
    return true // 默认返回 true
}
```

**改进方案 A**: 检测 MCP 功能可用性（简单）
```go
func (s *MCPService) checkMCPServerConnection() bool {
    // MCP Server 是 stdio 进程，无法直接检测
    // 返回 true 表示后端 API（MCP 功能实现）正常运行
    return true
}
```
**前端显示**: "MCP 功能可用"

**改进方案 B**: 检测最近的工具调用（完整）
```go
func (s *MCPService) checkMCPServerConnection() bool {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    lastCall, err := s.mcpRepo.GetLastToolCall(ctx)
    if err != nil || lastCall == nil {
        return false // 没有调用记录
    }

    // 如果最近 5 分钟内有工具调用，说明 AnythingLLM 正在使用
    return time.Since(lastCall.Timestamp) < 5*time.Minute
}
```
**前端显示**: "最近活跃" / "空闲"

**要求**: 需要中间件或工具函数记录调用到数据库

### 建议 3: 添加后端记录端点

在 `server/controller/mcp.go` 添加：

```go
// RecordToolCall 记录MCP工具调用
func (c *MCPController) RecordToolCall(ctx *gin.Context) {
    var req dto.RecordToolCallRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "参数错误", err), true)
        return
    }
    
    err := c.mcpService.RecordToolCall(ctx, req.ToolName, req.Duration, req.Success, req.Error)
    if err != nil {
        response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "记录失败", err), true)
        return
    }
    
    response.Success(ctx, "记录成功", nil)
}
```

在 `server/router/router.go` 添加路由：
```go
mcp.POST("/tool-calls/record", mcpController.RecordToolCall)
```

### 建议 4: 优化前端文案

将 "MCP 连接状态" 改为 "MCP 功能状态"：

```typescript
// web/src/components/common/mcp-status-indicator.tsx
<div className="font-medium">MCP 功能状态</div>
```

## 测试步骤

### 1. 测试 MCP Server 基本功能

```bash
# 启动后端
cd /Users/duheling/Downloads/AI-Waf
docker compose up -d mrya

# 运行 MCP Server（测试模式）
cd mcp-server
export WAF_BACKEND_URL=http://localhost:2333
export WAF_API_TOKEN=your-token-here
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./ai-waf-mcp
```

**预期输出**: 返回 31 个工具列表的 JSON

### 2. 测试 AnythingLLM 集成

**配置 AnythingLLM**:
```json
{
  "ai-waf": {
    "command": "/Users/duheling/Downloads/AI-Waf/mcp-server/ai-waf-mcp",
    "env": {
      "WAF_BACKEND_URL": "http://localhost:2333",
      "WAF_API_TOKEN": "eyJhbGci..."
    }
  }
}
```

**测试对话**:
- "列出最近的攻击日志"
- "显示 WAF 统计信息"
- "创建一条新的规则"

### 3. 检查工具调用记录（添加中间件后）

```bash
# 连接 MongoDB
docker exec -it AI-Waf-mrya-mongo-1 mongosh

# 查询工具调用记录
use ai-waf
db.mcp_tool_calls.find().sort({timestamp: -1}).limit(10)
```

**预期输出**: 显示最近的工具调用记录

### 4. 测试前端状态显示

访问 http://localhost:2333

**预期显示**:
- MCP 功能状态: ✅ 可用
- 可用工具: 31
- 最近调用: (如果有记录) "2 分钟前"

## 总结

### ✅ 当前实现正确性

| 方面 | 状态 | 说明 |
|------|------|------|
| MCP Server 创建 | ✅ 正确 | 使用官方 SDK 模式 |
| 工具注册 | ✅ 正确 | 31 个工具，签名正确 |
| Stdio 传输 | ✅ 正确 | 符合本地客户端使用场景 |
| 工具实现 | ✅ 正确 | 调用后端 HTTP API |
| 环境变量配置 | ✅ 正确 | 支持自定义后端 URL 和 Token |

### ⚠️ 可改进的部分

| 方面 | 优先级 | 说明 |
|------|--------|------|
| 工具调用追踪 | 高 | 添加中间件记录到数据库 |
| 连接状态检测 | 中 | 改进 checkMCPServerConnection 逻辑 |
| 日志和监控 | 中 | 添加详细的调用日志 |
| 前端文案 | 低 | 将"连接状态"改为"功能状态" |

### 🎯 核心结论

**本项目的 MCP Server 实现与官方示例一致，架构设计正确！**

唯一的"问题"不是实现错误，而是：
1. **架构特性**: stdio MCP Server 无法被后端直接检测（这是正常的）
2. **功能缺失**: 缺少工具调用追踪功能（可以通过中间件补充）

**建议的实施顺序**:
1. 立即：修改前端文案，明确"功能状态"而非"连接状态"
2. 短期：添加工具调用追踪中间件和后端记录端点
3. 中期：改进 checkMCPServerConnection 逻辑，基于工具调用记录
4. 长期：添加详细的监控和分析功能
