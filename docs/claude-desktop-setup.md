# Claude Desktop MCP 配置指南

## 概述

AI-WAF MCP Server 已集成到 Docker Compose 中,可以通过 Claude Desktop 访问以下 WAF 分析工具:

### 可用工具

1. **analyze_attack_patterns** - 分析攻击模式并生成威胁情报
2. **generate_waf_rules** - 基于攻击模式生成 WAF 规则
3. **get_waf_statistics** - 获取 WAF 统计数据
4. **get_attack_pattern** - 获取特定攻击模式详情
5. **list_generated_rules** - 列出所有生成的 WAF 规则
6. **get_recent_attacks** - 获取最近的攻击记录

---

## 方式 1: Docker 容器方式 (推荐)

### 1. 启动服务

```bash
# 构建并启动所有服务
docker compose up -d --build

# 或只启动 MCP Server
docker compose up -d mcp-server

# 验证服务运行状态
docker ps | grep mcp-server
```

### 2. 配置 Claude Desktop

编辑 Claude Desktop 配置文件:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`  
**Linux:** `~/.config/Claude/claude_desktop_config.json`

添加以下配置:

```json
{
  "mcpServers": {
    "ai-waf": {
      "command": "docker",
      "args": [
        "exec",
        "-i",
        "ai-waf-mcp-server",
        "./mcp-server",
        "-mongo",
        "mongodb://root:example@mongodb:27017",
        "-db",
        "waf"
      ],
      "env": {}
    }
  }
}
```

### 3. 重启 Claude Desktop

完全退出并重新启动 Claude Desktop 应用。

### 4. 验证连接

在 Claude Desktop 中询问:

```
请使用 analyze_attack_patterns 工具分析最近的攻击模式
```

或

```
请使用 get_waf_statistics 工具查看 WAF 统计数据
```

---

## 方式 2: 本地直接运行

### 1. 编译并启动

```bash
# 使用管理脚本
./scripts/manage-mcp.sh local

# 或手动编译运行
cd coraza-spoa
go build -o mcp-server ./cmd/mcp-server
./mcp-server -mongo "mongodb://root:example@localhost:27017" -db waf
```

### 2. 配置 Claude Desktop

```json
{
  "mcpServers": {
    "ai-waf": {
      "command": "/Users/duheling/Downloads/AI-Waf/coraza-spoa/mcp-server",
      "args": [
        "-mongo",
        "mongodb://root:example@localhost:27017",
        "-db",
        "waf"
      ],
      "env": {}
    }
  }
}
```

**注意:** 请将路径替换为实际的绝对路径。

---

## 管理脚本使用

项目提供了便捷的管理脚本 `scripts/manage-mcp.sh`:

```bash
# 本地启动
./scripts/manage-mcp.sh local

# Docker 启动
./scripts/manage-mcp.sh docker

# 查看日志
./scripts/manage-mcp.sh logs

# 停止服务
./scripts/manage-mcp.sh stop

# 重启服务
./scripts/manage-mcp.sh restart

# 进入容器
./scripts/manage-mcp.sh shell
```

---

## 故障排查

### 问题 1: Claude Desktop 无法连接

**症状:** 工具未显示或连接失败

**解决方案:**

```bash
# 1. 检查容器状态
docker ps | grep mcp-server

# 2. 查看日志
docker logs ai-waf-mcp-server

# 3. 测试 MCP Server
docker exec -i ai-waf-mcp-server ./mcp-server -mongo mongodb://root:example@mongodb:27017 -db waf
```

### 问题 2: MongoDB 连接失败

**症状:** 日志中显示 "failed to connect to MongoDB"

**解决方案:**

```bash
# 1. 检查 MongoDB 状态
docker compose ps mongodb

# 2. 验证 MongoDB 连接
docker exec -it ai-waf-mongodb mongosh -u root -p example

# 3. 检查网络连接
docker network inspect waf-network
```

### 问题 3: 工具调用失败

**症状:** Claude Desktop 显示工具但执行失败

**解决方案:**

```bash
# 1. 查看详细日志
docker logs -f ai-waf-mcp-server

# 2. 检查数据库数据
docker exec -it ai-waf-mongodb mongosh -u root -p example --eval "use waf; db.attack_patterns.countDocuments()"

# 3. 重启服务
docker compose restart mcp-server
```

### 问题 4: STDIO 通信问题

**症状:** 容器启动但无法通信

**检查配置:**

```yaml
# docker-compose.yaml 中必须包含:
stdin_open: true  # 启用标准输入
tty: true         # 分配伪终端
```

---

## 日志级别配置

MCP Server 使用 zerolog 记录日志,默认输出到 stderr。

### 查看不同级别日志

```bash
# 实时查看所有日志
docker logs -f ai-waf-mcp-server 2>&1 | jq -r '.level + " | " + .message'

# 只看错误日志
docker logs ai-waf-mcp-server 2>&1 | jq -r 'select(.level == "error")'

# 只看调试日志
docker logs ai-waf-mcp-server 2>&1 | jq -r 'select(.level == "debug")'
```

---

## 性能监控

### 检查资源使用

```bash
# CPU 和内存使用
docker stats ai-waf-mcp-server

# 容器详细信息
docker inspect ai-waf-mcp-server
```

### 监控工具调用

MCP Server 会记录每个工具调用的日志:

```bash
docker logs ai-waf-mcp-server 2>&1 | grep -E "Tool|handler"
```

---

## 安全建议

### 1. 生产环境配置

```yaml
# docker-compose.yaml
mcp-server:
  environment:
    - MONGO_URI=mongodb://mcp_user:${MCP_PASSWORD}@mongodb:27017
    - DATABASE=waf
  # 不要暴露端口
  # ports: [] 
```

### 2. MongoDB 用户权限

```javascript
// 创建专用 MCP 用户
use waf
db.createUser({
  user: "mcp_user",
  pwd: "secure_password",
  roles: [
    { role: "read", db: "waf" }  // 只读权限
  ]
})
```

### 3. 网络隔离

```yaml
# docker-compose.yaml
networks:
  waf-network:
    driver: bridge
    internal: false  # 如果不需要外网访问,设为 true
```

---

## 开发调试

### 启用调试日志

修改 [mcp_server.go](../coraza-spoa/internal/ai-analyzer/mcp_server.go):

```go
// 创建 zerolog
logger := zerolog.New(os.Stderr).
    Level(zerolog.DebugLevel).  // 修改为 DebugLevel
    With().
    Timestamp().
    Logger()
```

### 本地开发流程

```bash
# 1. 启动 MongoDB (仅)
docker compose up -d mongodb

# 2. 本地运行 MCP Server (带热重载)
cd coraza-spoa
go run ./cmd/mcp-server -mongo mongodb://root:example@localhost:27017 -db waf

# 3. 在另一个终端测试
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./mcp-server -mongo mongodb://root:example@localhost:27017 -db waf
```

---

## 更新和维护

### 更新 MCP Server

```bash
# 1. 停止服务
docker compose stop mcp-server

# 2. 重新构建
docker compose build mcp-server

# 3. 启动新版本
docker compose up -d mcp-server

# 4. 验证版本
docker logs ai-waf-mcp-server | head -20
```

### 数据备份

```bash
# 备份 MongoDB 数据
docker exec ai-waf-mongodb mongodump -u root -p example -d waf -o /backup
docker cp ai-waf-mongodb:/backup ./backup-$(date +%Y%m%d)
```

---

## 参考资料

- [MCP 官方文档](https://modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Claude Desktop 配置](https://docs.anthropic.com/claude/docs/mcp)
- [项目 README](../README.md)
