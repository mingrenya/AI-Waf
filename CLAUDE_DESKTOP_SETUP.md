# Claude Desktop MCP 配置验证指南

## ✅ 端点修复已完成

以下端点路径已修复并验证：

| 功能 | 旧端点 (错误) | 新端点 (正确) | 状态 |
|------|--------------|--------------|------|
| 安全指标 | `/api/v1/stats/security` | `/api/v1/stats/security-metrics` | ✅ 已修复 |
| 系统健康 | `/api/v1/stats/health` | `/api/v1/runner/status` | ✅ 已修复 |
| 攻击日志 | `/api/v1/waf/logs` | `/api/v1/log` | ✅ 已修复 |

---

## 🚀 快速启动步骤

### 1. 重新编译 MCP Server（如未完成）
```bash
cd /Users/duheling/Downloads/AI-Waf/mcp-server
make clean && make build
```

### 2. 验证二进制文件
```bash
ls -lh ai-waf-mcp
# 应该看到新生成的文件
```

### 3. 确保后端运行
```bash
docker compose ps
# mrya 应该是 running 状态

# 测试连通性
curl http://localhost:2333/health
# 应返回: {"status":"ok"}
```

### 4. 配置 Claude Desktop

**配置文件位置**:
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`

**配置内容**:
```json
{
  "mcpServers": {
    "ai-waf": {
      "command": "/Users/duheling/Downloads/AI-Waf/mcp-server/ai-waf-mcp",
      "env": {
        "WAF_BACKEND_URL": "http://localhost:2333",
        "WAF_API_TOKEN": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY2MWM1NDI3YmFlMmZmM2YzN2Q3NzEiLCJ1c2VybmFtZSI6ImFkbWluIiwicm9sZSI6ImFkbWluIiwiaXNzIjoiTVJZYSIsInN1YiI6IjY5NjYxYzU0MjdiYWUyZmYzZjM3ZDc3MSIsImV4cCI6MTc3MDAxMTY2NCwibmJmIjoxNzY5OTI1MjY0LCJpYXQiOjE3Njk5MjUyNjR9.NvoKVMHmq-mBST-WvZ8GfgnFME974USHFsRqKcdN5uo"
      }
    }
  }
}
```

### 5. 重启 Claude Desktop

完全退出 Claude Desktop 后重新打开。

---

## 🧪 测试 MCP 工具

在 Claude Desktop 对话中尝试：

### 测试 1: 获取统计概览
```
请获取 WAF 的统计概览
```

### 测试 2: 查看安全指标
```
显示过去24小时的安全指标
```

### 测试 3: 查看实时 QPS
```
获取实时 QPS 数据
```

### 测试 4: 查看系统状态
```
检查 WAF 系统健康状态
```

### 测试 5: 查询攻击日志
```
列出最近的攻击日志
```

---

## 🔍 故障排查

### 问题 1: Token 已过期（401）

**症状**: 看到 `未授权访问` 错误

**解决**:
```bash
# 1. 登录后端获取新 Token
open http://localhost:2333

# 2. 登录后，在开发者工具中复制新的 JWT Token

# 3. 更新 Claude Desktop 配置中的 WAF_API_TOKEN

# 4. 重启 Claude Desktop
```

**Token 过期时间**: 2025-04-03 (约60天后)

### 问题 2: 404 Not Found

**症状**: 工具调用返回 `404 page not found`

**检查清单**:
- [ ] 已重新编译: `cd mcp-server && make build`
- [ ] 后端运行正常: `curl http://localhost:2333/health`
- [ ] 配置路径正确: 检查 `command` 指向 `ai-waf-mcp` 文件
- [ ] 已重启 Claude Desktop

### 问题 3: Connection Refused

**症状**: 无法连接到 `http://localhost:2333`

**解决**:
```bash
# 检查 Docker 容器状态
docker compose ps

# 如果未运行，启动容器
docker compose up -d

# 查看日志
docker compose logs -f mrya
```

### 问题 4: MCP Server 未显示在 Claude Desktop

**解决**:
1. 检查配置文件语法是否正确（JSON 格式）
2. 检查二进制文件是否存在且可执行
3. 完全退出 Claude Desktop 后重新打开
4. 查看 Claude Desktop 日志（如果可用）

---

## 📊 预期行为

### 成功的工具调用日志

在后端日志中应该看到（使用 `docker compose logs -f mrya`）:

```
mrya-waf  | [ INFO] HTTP Request method:GET path:/api/v1/stats/overview status:200
mrya-waf  | [ INFO] HTTP Request method:GET path:/api/v1/stats/security-metrics status:200
mrya-waf  | [ INFO] HTTP Request method:GET path:/api/v1/runner/status status:200
mrya-waf  | [ INFO] HTTP Request method:GET path:/api/v1/log status:200
```

### 不应该看到的错误

❌ **不应该有这些**:
```
status:404  # 端点不存在
status:401  # Token 过期或未设置
```

---

## 📚 相关文档

- [MCP 端点审计报告](./MCP_ENDPOINT_AUDIT_FIXES.md) - 详细的代码审计
- [环境变量修复指南](./ENV_FIX_VERIFICATION_GUIDE.md) - Docker 环境配置
- [MCP 架构说明](./MCP_ARCHITECTURE_EXPLANATION.md) - 系统架构
- [后端 API 文档](./server/docs/) - API 接口文档

---

## ✨ 成功标志

当所有配置正确时，您将能够：

1. ✅ 在 Claude Desktop 中看到 "ai-waf" MCP 服务器连接
2. ✅ 工具调用返回数据而非错误
3. ✅ 后端日志显示 200 状态码
4. ✅ 无 404 或 401 错误

---

## 🆘 需要帮助？

如果遇到问题：

1. 运行自动验证脚本: `./test-mcp-claude-desktop.sh`
2. 查看详细审计报告: `MCP_ENDPOINT_AUDIT_FIXES.md`
3. 检查后端日志: `docker compose logs -f mrya | grep -E "404|401|ERROR"`

**最后更新**: 2026-02-01
