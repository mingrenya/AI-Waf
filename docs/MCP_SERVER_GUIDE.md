# AI-WAF MCP Server 配置指南

## 概述

AI-WAF MCP Server 基于官方 [Model Context Protocol (MCP) Go SDK](https://github.com/modelcontextprotocol/go-sdk) 实现,向 LLM(如 Claude Desktop)暴露 WAF 安全分析工具。

### MCP 架构

```
┌─────────────────────┐
│  Claude Desktop     │  (MCP Host/Client)
│  或其他 LLM 客户端   │
└──────────┬──────────┘
           │ JSON-RPC over STDIO
           ▼
┌─────────────────────┐
│  AI-WAF MCP Server  │  (基于官方 Go SDK)
│  - 攻击模式分析     │
│  - 规则生成         │
│  - WAF 统计         │
│  - 日志查询         │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  MongoDB            │
│  - waf_log          │
│  - attack_patterns  │
│  - generated_rules  │
└─────────────────────┘
```

## 提供的工具

MCP Server 暴露以下 6 个工具供 LLM 调用:

### 1. `analyze_attack_patterns`
分析 WAF 日志,使用机器学习检测攻击模式。

**参数:**
- `time_window_hours` (可选): 分析时间窗口,默认 24 小时
- `min_samples` (可选): 最小样本数,默认 100
- `confidence_threshold` (可选): 置信度阈值 0-1,默认 0.7

**示例:**
```
在 Claude Desktop 中询问: "分析过去 24 小时的攻击模式"
```

### 2. `generate_waf_rules`
基于攻击模式生成 ModSecurity 规则。

**参数:**
- `pattern_id` (必需): 攻击模式 ID
- `rule_type` (可选): modsecurity 或 micro_rule,默认 modsecurity

**示例:**
```
"为模式 67890abcdef 生成 ModSecurity 规则"
```

### 3. `get_waf_statistics`
获取 WAF 统计信息。

**参数:**
- `time_range` (可选): 1h, 24h, 7d, 30d,默认 24h

**示例:**
```
"显示过去 7 天的 WAF 统计数据"
```

### 4. `get_attack_pattern`
获取指定攻击模式的详细信息。

**参数:**
- `pattern_id` (必需): 攻击模式 ID

### 5. `list_generated_rules`
列出 AI 生成的防护规则。

**参数:**
- `status` (可选): pending, approved, deployed, rejected
- `limit` (可选): 返回数量,默认 20

**示例:**
```
"列出所有已部署的 AI 规则"
```

### 6. `get_recent_attacks`
获取最近的攻击日志。

**参数:**
- `limit` (可选): 返回数量,默认 50
- `severity` (可选): low, medium, high, critical

**示例:**
```
"显示最近 100 条高危攻击"
```

## 安装配置

### 前提条件

- Go 1.22+
- MongoDB 运行中
- Claude Desktop 或其他 MCP 客户端

### 1. 编译 MCP Server

```bash
cd /Users/duheling/Downloads/AI-Waf/coraza-spoa
go build -o mcp-server ./cmd/mcp-server
```

### 2. 配置 Claude Desktop

编辑 Claude Desktop 配置文件:

**macOS:**
```bash
code ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

**Windows:**
```cmd
notepad %APPDATA%\Claude\claude_desktop_config.json
```

添加 AI-WAF MCP Server:

```json
{
  "mcpServers": {
    "ai-waf": {
      "command": "/Users/duheling/Downloads/AI-Waf/scripts/start-mcp-server.sh",
      "env": {
        "MONGO_URI": "mongodb://localhost:27017",
        "DATABASE": "waf"
      }
    }
  }
}
```

**注意:** 
- 使用绝对路径
- 确保脚本有执行权限: `chmod +x scripts/start-mcp-server.sh`

### 3. 重启 Claude Desktop

关闭并重新打开 Claude Desktop,在 "+" 按钮 → "Connectors" 中应该能看到 `ai-waf` 服务器。

## 使用示例

### 安全分析场景

```
用户: "分析我的 WAF 系统,找出最近的攻击趋势"

Claude 将调用:
1. get_waf_statistics (获取概况)
2. analyze_attack_patterns (检测模式)
3. get_recent_attacks (查看实例)

返回完整的安全分析报告
```

### 规则生成场景

```
用户: "为最近检测到的 SQL 注入攻击生成防护规则"

Claude 将:
1. analyze_attack_patterns (找 SQL 注入模式)
2. generate_waf_rules (生成 ModSecurity 规则)
3. 提供规则建议和部署说明
```

### 实时监控场景

```
用户: "显示当前最严重的攻击"

Claude 将:
1. get_recent_attacks (severity=critical)
2. get_attack_pattern (获取模式详情)
3. 提供威胁评估和建议
```

## 手动测试

不使用 Claude Desktop,可以手动测试 MCP Server:

```bash
# 启动 MCP Server
cd /Users/duheling/Downloads/AI-Waf/coraza-spoa
./mcp-server -mongo mongodb://localhost:27017 -db waf

# 在另一个终端发送 JSON-RPC 请求
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | nc localhost 9999

# 列出工具
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | nc localhost 9999

# 调用工具
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_waf_statistics","arguments":{"time_range":"24h"}}}' | nc localhost 9999
```

## 故障排查

### MCP Server 未出现在 Claude Desktop

1. 检查配置文件路径和 JSON 格式
2. 确保脚本有执行权限
3. 查看日志: `tail -f logs/mcp-server.log`
4. 重启 Claude Desktop

### 工具调用失败

1. 检查 MongoDB 是否运行: `mongosh`
2. 确认数据库中有数据: `db.waf_log.countDocuments()`
3. 查看 MCP Server 日志中的错误信息

### 权限问题

```bash
# 给脚本添加执行权限
chmod +x /Users/duheling/Downloads/AI-Waf/scripts/start-mcp-server.sh

# 检查 MongoDB 连接权限
mongosh "mongodb://localhost:27017/waf" --eval "db.runCommand({ping:1})"
```

## 开发调试

### 启用详细日志

修改 `cmd/mcp-server/main.go`:

```go
logger := zerolog.New(os.Stderr).
    Level(zerolog.DebugLevel).  // 启用 Debug
    With().Timestamp().Logger()
```

### 添加新工具

1. 在 `handleToolsList()` 中添加工具定义
2. 在 `handleToolsCall()` 中添加 case 分支
3. 实现工具处理方法
4. 重新编译并重启

## 最佳实践

1. **安全性**: MCP Server 具有完整的 MongoDB 访问权限,仅在可信环境中使用
2. **性能**: 大数据量查询使用 `limit` 参数限制返回数量
3. **监控**: 定期检查 `logs/mcp-server.log` 了解使用情况
4. **更新**: 修改代码后记得重新编译并重启 Claude Desktop

## 参考资源

- [MCP 官方文档](https://modelcontextprotocol.io/)
- [MCP GitHub](https://github.com/modelcontextprotocol)
- [Claude Desktop 下载](https://claude.ai/download)
- [JSON-RPC 2.0 规范](https://www.jsonrpc.org/specification)
