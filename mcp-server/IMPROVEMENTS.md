# MCP Server 改进说明

基于官方 go-sdk examples 的最佳实践，AI-Waf MCP Server 已完成以下改进。

## 📋 改进内容

### 1. **添加日志中间件** (参考 `examples/server/middleware/main.go`)

**功能**:
- ✅ 记录每个 MCP 方法调用
- ✅ 记录工具调用的参数和结果
- ✅ 记录执行时间
- ✅ 区分成功/失败状态

**实现**:
```go
server.AddReceivingMiddleware(createLoggingMiddleware())
```

**日志示例**:
```
[MCP] Method: tools/call | Session: abc123
[MCP] Tool Call: list_attack_logs | Args: map[limit:10]
[MCP] ✅ Success: tools/call | Duration: 150ms
[MCP] Tool Result: IsError=false | ContentCount=1
```

### 2. **添加工具调用追踪中间件**

**功能**:
- ✅ 异步记录每次工具调用到后端数据库
- ✅ 记录调用参数、执行时间、成功状态
- ✅ 不阻塞工具执行
- ✅ 失败时只记录警告，不影响工具正常使用

**实现**:
```go
server.AddReceivingMiddleware(createTrackingMiddleware(client))
```

**记录到**:
- 后端接口: `POST /api/v1/mcp/tool-calls/record`
- 数据库集合: `mcp_tool_calls`

**数据结构**:
```json
{
  "toolName": "list_attack_logs",
  "arguments": {"limit": 10},
  "duration": 150,
  "success": true,
  "timestamp": "2026-01-16T10:30:00Z"
}
```

### 3. **改进的 MCP Server 版本**

#### **main.go** - stdio 版本（生产环境）

**特点**:
- ✅ 保持原有 stdio 传输方式（AnythingLLM/Claude Desktop）
- ✅ 可选的日志中间件（通过环境变量启用）
- ✅ 可选的追踪中间件（通过环境变量启用）
- ✅ 共享的中间件实现（middleware.go）

**使用方式**:
```bash
# 不启用中间件（生产环境）
./ai-waf-mcp

# 启用调试日志
MCP_DEBUG=1 ./ai-waf-mcp

# 启用工具调用追踪
MCP_TRACK=1 ./ai-waf-mcp

# 同时启用
MCP_DEBUG=1 MCP_TRACK=1 ./ai-waf-mcp
```

**注意**: stdio 模式下，日志输出到 stderr，不会干扰 JSON-RPC 通信。

#### **cmd/server-http/main.go** - HTTP 版本（监控/测试）

**特点**:
- ✅ 使用 `StreamableHTTPHandler` 提供 HTTP 接口
- ✅ 后端可以通过 HTTP 检测服务状态
- ✅ 默认日志和追踪中间件
- ✅ 支持自定义端口

**使用方式**:
```bash
cd mcp-server/cmd/server-http
export WAF_BACKEND_URL=http://localhost:2333
export WAF_API_TOKEN=your-token
go run main.go -addr localhost:8080
```

#### **cmd/client-test/main.go** - 测试客户端

**功能**:
- ✅ 连接到 HTTP MCP Server
- ✅ 列出所有可用工具（按类别分组）
- ✅ 调用指定工具并显示结果
- ✅ 显示执行时间和格式化输出

**使用方式**:
```bash
cd mcp-server/cmd/client-test

# 列出所有工具
go run main.go -server http://localhost:8080

# 调用特定工具
go run main.go -server http://localhost:8080 \
  -tool list_attack_logs \
  -args '{"limit":5}'
```

#### **middleware.go** - 共享中间件（新增）

**功能**:
- ✅ 统一的日志中间件实现
- ✅ 统一的追踪中间件实现
- ✅ 避免代码重复
- ✅ 符合官方最佳实践

### 4. **自动化测试脚本** - test-mcp.sh

**功能**:
- ✅ 自动启动 HTTP MCP Server
- ✅ 测试工具列表和调用
- ✅ 验证工具调用记录
- ✅ 检查 MCP 连接状态
- ✅ 自动清理测试环境

**使用方式**:
```bash
cd mcp-server
./test-mcp.sh
```

## 📁 项目结构

```
mcp-server/
├── main.go                    # stdio 版本（生产环境，AnythingLLM）
├── middleware.go              # 共享中间件实现（新增）
├── go.mod
├── go.sum
├── test-mcp.sh               # 自动化测试脚本
├── IMPROVEMENTS.md           # 改进说明文档
├── cmd/
│   ├── server-http/
│   │   └── main.go          # HTTP 版本（监控/测试）
│   └── client-test/
│       └── main.go          # 测试客户端
└── tools/
    ├── api_client.go
    └── *.go                  # 31 个工具实现
```

### 改进前

```
┌──────────────────────┐
│ AnythingLLM/Claude   │
└──────────┬───────────┘
           │ stdio
           ▼
┌──────────────────────┐
│ MCP Server (stdio)   │  ❌ 无日志
└──────────┬───────────┘  ❌ 无追踪
           │              ❌ 后端无法检测
           ▼
┌──────────────────────┐
│ 后端 API             │
└──────────────────────┘
```

### 改进后

```
场景 1: AnythingLLM 使用（stdio）
┌──────────────────────┐
│ AnythingLLM/Claude   │
└──────────┬───────────┘
           │ stdio
           ▼
┌──────────────────────┐
│ MCP Server (stdio)   │  ✅ 可选日志 (MCP_DEBUG=1)
│ + 日志中间件          │  ✅ 可选追踪 (MCP_TRACK=1)
│ + 追踪中间件          │
└──────────┬───────────┘
           │ HTTP
           ▼
┌──────────────────────┐
│ 后端 API             │
│ /tool-calls/record   │  ✅ 记录工具调用
└──────────────────────┘

场景 2: HTTP MCP Client 使用（新增）
┌──────────────────────┐
│ MCP Client           │  ✅ 可以是任何 HTTP 客户端
└──────────┬───────────┘  ✅ 测试客户端
           │ HTTP
           ▼
┌──────────────────────┐
│ MCP Server (HTTP)    │  ✅ 默认日志
│ + 日志中间件          │  ✅ 默认追踪
│ + 追踪中间件          │  ✅ 可检测连接状态
└──────────┬───────────┘
           │ HTTP
           ▼
┌──────────────────────┐
│ 后端 API             │
│ /tool-calls/record   │  ✅ 记录工具调用
└──────────────────────┘
```

## 📊 对比官方示例

| 功能 | 官方示例 | AI-Waf 实现 | 状态 |
|------|---------|------------|------|
| 基本 Server 创建 | `examples/server/hello` | ✅ main.go | ✅ 完成 |
| HTTP 传输 | `examples/http` | ✅ server-http.go | ✅ 完成 |
| 日志中间件 | `examples/server/middleware` | ✅ createLoggingMiddleware | ✅ 完成 |
| 工具调用追踪 | - | ✅ createTrackingMiddleware | ✅ 新增 |
| 测试客户端 | `examples/client/listfeatures` | ✅ client-test.go | ✅ 完成 |
| 认证中间件 | `examples/server/auth-middleware` | ⏳ 待实现 | 可选 |
| 限流中间件 | `examples/server/rate-limiting` | ⏳ 待实现 | 可选 |

## 🚀 使用指南

### 开发环境测试

```bash
# 1. 启动后端
cd /Users/duheling/Downloads/AI-Waf
docker compose up -d mrya

# 2. 启动 HTTP MCP Server
cd mcp-server
export WAF_BACKEND_URL=http://localhost:2333
export WAF_API_TOKEN=test-token
go run server-http.go -addr localhost:8080

# 3. 在另一个终端测试
cd mcp-server
go run client-test.go -server http://localhost:8080
go run client-test.go -server http://localhost:8080 -tool get_stats_overview -args '{}'

# 4. 查看工具调用记录
curl http://localhost:2333/api/v1/mcp/tool-calls?limit=10 | jq '.'
```

### 生产环境（AnythingLLM）

```bash
# 编译 MCP Server
cd mcp-server
go build -o ai-waf-mcp main.go

# AnythingLLM 配置（不启用调试）
{
  "ai-waf": {
    "command": "/path/to/ai-waf-mcp",
    "env": {
      "WAF_BACKEND_URL": "http://localhost:2333",
      "WAF_API_TOKEN": "your-production-token"
    }
  }
}

# 如需启用追踪（推荐）
{
  "ai-waf": {
    "command": "/path/to/ai-waf-mcp",
    "env": {
      "WAF_BACKEND_URL": "http://localhost:2333",
      "WAF_API_TOKEN": "your-production-token",
      "MCP_TRACK": "1"
    }
  }
}
```

## 🔍 调试技巧

### 查看日志

**HTTP 版本**:
```bash
# 日志会直接输出到终端
go run server-http.go -addr localhost:8080
```

**stdio 版本**:
```bash
# 启用调试模式，日志输出到 stderr
MCP_DEBUG=1 ./ai-waf-mcp 2> mcp-debug.log &
tail -f mcp-debug.log
```

### 查看工具调用记录

```bash
# 获取最近的工具调用
curl http://localhost:2333/api/v1/mcp/tool-calls?limit=20 | jq '.data'

# 查看特定工具的调用
curl http://localhost:2333/api/v1/mcp/tool-calls | jq '.data[] | select(.toolName=="list_attack_logs")'

# 统计工具调用次数
curl http://localhost:2333/api/v1/mcp/tool-calls?limit=1000 | jq '.data | group_by(.toolName) | map({tool: .[0].toolName, count: length})'
```

### 检查 MCP 连接状态

```bash
# 前端显示的状态就是基于这个接口
curl http://localhost:2333/api/v1/mcp/status | jq '.'
```

## 📝 总结

基于官方 examples 的改进，AI-Waf MCP Server 现在具备：

1. ✅ **生产级日志**: 完整的方法调用和执行时间记录
2. ✅ **自动追踪**: 工具调用自动记录到数据库
3. ✅ **双模式支持**: stdio（AnythingLLM）和 HTTP（测试/监控）
4. ✅ **灵活配置**: 通过环境变量控制功能开关
5. ✅ **完整测试**: 自动化测试脚本和测试客户端
6. ✅ **符合规范**: 完全遵循官方 SDK 最佳实践

这使得：
- 前端可以准确显示 MCP 使用情况（基于工具调用记录）
- 开发者可以轻松调试和监控
- 系统可以追踪 AI 工具的实际使用情况
