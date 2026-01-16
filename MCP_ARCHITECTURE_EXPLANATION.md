# MCP Server 架构说明

## 当前架构（正确）

```
┌──────────────────┐
│  AnythingLLM /   │  (MCP Client)
│  Claude Desktop  │
└────────┬─────────┘
         │ stdio (Standard Input/Output)
         ▼
┌────────────────────┐
│  MCP Server        │  运行 /Users/duheling/Downloads/AI-Waf/mcp-server/ai-waf-mcp
│  (ai-waf-mcp)      │  实现：基于官方 go-sdk
└────────┬───────────┘
         │ HTTP API 调用
         ▼
┌────────────────────┐
│  后端 API          │  http://localhost:2333 或 http://host.docker.internal:2333
│  (mrya server)     │
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│  MongoDB           │
└────────────────────┘

同时：

┌────────────────────┐
│  前端 Web UI       │  http://localhost:2333
└────────┬───────────┘
         │ HTTP API 调用
         ▼
┌────────────────────┐
│  后端 API          │  /api/v1/mcp/status
└────────────────────┘
```

## 关键理解

### 1. MCP Server 的运行方式

**MCP Server 不是一个常驻后台服务！**

根据官方 SDK 示例（`examples/server/hello/main.go`）：

```go
func main() {
    server := mcp.NewServer(&mcp.Implementation{Name: "greeter"}, nil)
    mcp.AddTool(server, &mcp.Tool{Name: "greet"}, SayHi)
    
    // 通过 stdio 运行，等待客户端连接
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Printf("Server failed: %v", err)
    }
}
```

**MCP Server 只在以下情况运行**：
1. AnythingLLM/Claude Desktop 启动它作为子进程
2. 通过 stdio 与客户端通信
3. 客户端关闭连接后，MCP Server 退出

### 2. 前端"未连接"状态是正常的

**问题**：前端调用 `/api/v1/mcp/status` 显示 "connected: false"

**原因**：
- 后端尝试检测 MCP Server 是否在运行
- 但 MCP Server 是 **stdio 进程**，不是 HTTP 服务
- 后端无法通过网络连接检测 stdio 进程

**正确的状态含义**：
- `connected: true` 应该表示：**MCP 功能可用**（后端 API 正常）
- `connected: false` 应该表示：**MCP 功能不可用**（后端 API 故障）

### 3. 实际的连接状态

**真实情况**：
```
AnythingLLM → ai-waf-mcp (stdio) → 后端API ✅ 正常工作
前端 → 后端API → 检测 ai-waf-mcp ❌ 无法检测（stdio进程）
```

**AnythingLLM 配置示例**：
```json
{
  "command": "/Users/duheling/Downloads/AI-Waf/mcp-server/ai-waf-mcp",
  "env": {
    "WAF_BACKEND_URL": "http://localhost:2333",
    "WAF_API_TOKEN": "eyJhbGci..."
  }
}
```

当 AnythingLLM 调用工具时：
1. AnythingLLM 发送 JSON-RPC 请求到 stdio
2. MCP Server 接收请求，调用对应的工具函数
3. 工具函数调用后端 HTTP API
4. 后端API 返回结果
5. MCP Server 将结果返回给 AnythingLLM

## 修正方案

### 方案1：改变"连接"的含义（推荐）

修改 `checkMCPServerConnection()` 让它检测 **MCP 功能是否可用**：

```go
func (s *MCPService) checkMCPServerConnection() bool {
    // MCP Server 是 stdio 进程，无法直接检测
    // 这里返回 true 表示后端API（MCP功能实现）正常运行
    return true
}
```

**前端显示**：
- `connected: true` - MCP 功能可用（后端API运行中）
- `totalTools: 31` - 可用工具数量

### 方案2：检测最近的工具调用（可选）

如果想知道 AnythingLLM 是否在使用 MCP Server：

```go
func (s *MCPService) checkMCPServerConnection() bool {
    // 检查最近5分钟是否有工具调用记录
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    lastCall, err := s.mcpRepo.GetLastToolCall(ctx)
    if err != nil || lastCall == nil {
        return false  // 没有调用记录
    }

    // 如果最近5分钟内有工具调用，说明 AnythingLLM 正在使用
    return time.Since(lastCall.Timestamp) < 5*time.Minute
}
```

**但这需要 MCP Server 记录工具调用到数据库！**

### 方案3：添加工具调用记录功能（完整解决）

在 MCP Server 中，每次工具调用后记录到后端：

**1. 后端添加记录端点**：
```go
// POST /api/v1/mcp/tool-calls/record
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

**2. MCP Server 工具调用后记录**：
```go
func CreateListAttackLogs(client *APIClient) func(...) {
    return func(ctx context.Context, req *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, Output, error) {
        start := time.Now()
        
        // 执行工具逻辑
        result, output, err := doWork()
        
        // 记录到后端
        duration := time.Since(start).Milliseconds()
        _ = client.Post("/api/v1/mcp/tool-calls/record", map[string]interface{}{
            "toolName": "list_attack_logs",
            "duration": duration,
            "success":  err == nil,
            "error":    errMsg,
        })
        
        return result, output, err
    }
}
```

## 测试 MCP Server

### 1. 测试 AnythingLLM 连接

配置 AnythingLLM：
```json
{
  "ai-waf": {
    "command": "/Users/duheling/Downloads/AI-Waf/mcp-server/ai-waf-mcp",
    "env": {
      "WAF_BACKEND_URL": "http://localhost:2333",
      "WAF_API_TOKEN": "your-token-here"
    }
  }
}
```

在 AnythingLLM 中询问：
- "列出最近的攻击日志"
- "显示WAF统计信息"
- "创建一条新的MicroRule规则"

### 2. 测试本地运行

```bash
cd /Users/duheling/Downloads/AI-Waf/mcp-server

# 设置环境变量
export WAF_BACKEND_URL=http://localhost:2333
export WAF_API_TOKEN=your-token-here

# 运行（等待 stdin 输入）
./ai-waf-mcp

# 发送测试请求（JSON-RPC格式）
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./ai-waf-mcp
```

### 3. 检查日志

MCP Server 启动日志：
```
2026-01-16T09:20:00Z [INFO] 注册MCP工具...
2026-01-16T09:20:00Z [INFO] ================================
2026-01-16T09:20:00Z [INFO] AI-Waf MCP Server 启动成功
2026-01-16T09:20:00Z [INFO] 后端URL: http://localhost:2333
2026-01-16T09:20:00Z [INFO] 已注册31个MCP工具
2026-01-16T09:20:00Z [INFO] 等待MCP客户端连接...
```

当 AnythingLLM 连接时：
```
2026-01-16T09:20:05Z [INFO] 收到客户端连接
2026-01-16T09:20:05Z [INFO] 执行工具: list_attack_logs
2026-01-16T09:20:05Z [INFO] API请求: GET /api/v1/waf/logs
2026-01-16T09:20:05Z [INFO] 工具执行成功，耗时: 150ms
```

## 总结

### 当前状态
✅ MCP Server 实现正确（基于官方 SDK）
✅ 工具注册正确（31个工具）
✅ stdio 传输正确
✅ HTTP API 调用正确

### 前端显示问题
❌ "未连接"不准确
✅ 应该显示"功能可用"

### 建议的修改

1. **立即修改**：将 `checkMCPServerConnection()` 改为返回 `true`
2. **可选增强**：添加工具调用记录功能
3. **文案优化**：前端显示"MCP 功能状态"而不是"MCP 连接状态"

### 正确的理解

**MCP Server 不是后台服务**，而是：
- 一个命令行工具
- 由 AI 客户端（AnythingLLM/Claude Desktop）按需启动
- 通过 stdin/stdout 通信
- 调用后端 HTTP API 实现功能
- 客户端断开后自动退出

**前端 Web UI** 和 **MCP Server** 是两个独立的通道访问后端API，不存在直接连接关系。
