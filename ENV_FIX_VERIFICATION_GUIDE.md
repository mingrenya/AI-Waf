# 环境变量修复验证指南

## 问题回顾

**根本原因**: `docker-compose.yaml` 缺少 `env_file` 配置，导致 `.env` 文件中的 `MCP_API_TOKEN` 未被加载到 mcp-server 容器，所有API请求返回401。

**已修复内容**:
- ✅ `docker-compose.yaml` 添加 `env_file: - .env`
- ✅ 修改后端URL为 `http://mrya:2333` (Docker内部网络)
- ✅ 移除环境变量默认值 `${MCP_API_TOKEN:-}` → `${MCP_API_TOKEN}`

---

## 快速验证步骤

### 1. 检查配置文件

```bash
# 确认.env包含有效token
grep MCP_API_TOKEN .env

# 预期输出:
# MCP_API_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 2. 重启服务

```bash
# 停止所有服务
docker compose down

# 重新构建MCP Server镜像
docker compose build mcp-server

# 启动服务
docker compose up -d

# 或强制重新创建容器
docker compose up -d --force-recreate mcp-server
```

### 3. 验证环境变量注入

```bash
# 检查mcp-server容器的环境变量
docker compose exec mcp-server env | grep WAF

# 预期输出:
# WAF_BACKEND_URL=http://mrya:2333
# WAF_API_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY2MWM1NDI3YmFlMmZmM2YzN2Q3NzEi...

# 如果WAF_API_TOKEN显示完整JWT (三段式，含两个点)，表示配置成功！
```

### 4. 查看日志

```bash
# 实时查看MCP Server日志
docker compose logs -f mcp-server

# 查看后端日志
docker compose logs -f mrya

# 成功标志:
# - MCP Server启动日志显示 "Token: eyJhbGc..." (前20字符)
# - 不再有 "警告: 未设置 WAF_API_TOKEN 环境变量"
# - 调用工具时显示 "[API响应] ... 状态码: 200"
```

### 5. 测试工具调用(可选)

如果有测试脚本:
```bash
./test-mcp.sh
```

或手动测试HTTP版本:
```bash
# 启动HTTP版本(如果需要)
docker compose exec mcp-server ./server-http &

# 测试调用
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

---

## 验证检查清单

- [ ] **步骤1**: `grep MCP_API_TOKEN .env` 显示完整JWT token
- [ ] **步骤2**: `docker compose ps` 显示 mcp-server 容器为 `Up` 状态
- [ ] **步骤3**: `docker compose exec mcp-server env | grep WAF_API_TOKEN` 显示完整token
- [ ] **步骤4**: MCP Server日志不再输出token警告
- [ ] **步骤5**: 工具调用日志显示 `状态码: 200` 而非 `状态码: 401`
- [ ] **步骤6**: 后端日志显示成功的 `/api/v1/mcp/tool-calls/record` 请求

---

## 常见问题排查

### Q1: 环境变量仍然为空

**症状**:
```bash
$ docker compose exec mcp-server env | grep WAF_API_TOKEN
WAF_API_TOKEN=
```

**解决方法**:
```bash
# 1. 确认.env文件存在且包含token
cat .env | grep MCP_API_TOKEN

# 2. 确认docker-compose.yaml已添加env_file
grep "env_file" docker-compose.yaml

# 3. 完全删除容器和镜像重建
docker compose down -v
docker compose build --no-cache mcp-server
docker compose up -d
```

### Q2: 仍然返回401错误

**排查步骤**:

1. **检查token是否过期**:
   ```bash
   # 使用jwt.io或命令行解码token
   echo "eyJhbGc..." | cut -d'.' -f2 | base64 -d | jq .exp
   
   # 与当前时间戳对比
   date +%s
   ```

2. **检查后端URL连通性**:
   ```bash
   # 进入mcp-server容器测试
   docker compose exec mcp-server sh
   
   # 测试网络连通
   wget -O- http://mrya:2333/api/v1/health
   # 或
   curl http://mrya:2333/api/v1/health
   ```

3. **检查后端认证日志**:
   ```bash
   # 查看后端认证错误
   docker compose logs mrya | grep -i "unauthorized\|401"
   ```

### Q3: 如何获取新的token

如果token过期:

1. 访问后端管理界面: `http://localhost:2333`
2. 使用管理员账号登录
3. 创建服务账号或获取管理员token
4. 更新 `.env` 文件中的 `MCP_API_TOKEN`
5. 重启容器: `docker compose restart mcp-server`

### Q4: 容器无法访问 mrya 服务

**症状**: `docker compose logs mcp-server` 显示 `dial tcp: lookup mrya: no such host`

**解决方法**:
```bash
# 确认所有服务在同一网络
docker compose ps
docker network ls
docker network inspect ai-waf_waf-network

# 确认mrya容器正在运行
docker compose ps mrya

# 如果mrya未启动，先启动mrya
docker compose up -d mrya

# 再启动mcp-server
docker compose up -d mcp-server
```

---

## 本地开发测试(跳过Docker)

如果想快速验证修复而不重启Docker:

```bash
# 1. 设置环境变量
export WAF_BACKEND_URL="http://localhost:2333"
export WAF_API_TOKEN="eyJhbGc..."  # 从.env复制

# 2. 本地运行MCP Server
cd mcp-server
go run main.go middleware.go

# 3. 观察日志
# 应该显示:
# Token: eyJhbGc...
# AI-Waf MCP Server 启动成功
# 后端URL: http://localhost:2333
```

---

## 成功标志

当看到以下输出时，表示修复成功:

**MCP Server启动日志**:
```
================================
AI-Waf MCP Server 启动成功
后端URL: http://mrya:2333
已注册31个MCP工具
等待MCP客户端连接...
================================
```

**工具调用日志**:
```
[REQUEST] Session: xxx | Method: tools/call
[TOOL CALL] Name: list_attack_logs | Args: {...}
[API请求] GET http://mrya:2333/api/v1/logs
[API响应] GET /api/v1/logs - 状态码: 200 - 耗时: 45ms
[RESPONSE] Session: xxx | Method: tools/call | Status: OK
[TRACKING] Recorded: list_attack_logs
```

---

## 联系与支持

如果问题仍未解决:

1. 检查完整审计报告: [MCP_CODE_AUDIT_REPORT.md](./MCP_CODE_AUDIT_REPORT.md)
2. 查看官方文档: https://github.com/modelcontextprotocol/go-sdk
3. 提交Issue附带日志: `docker compose logs mcp-server > mcp-server.log`

---

**最后更新**: 2026年2月1日  
**验证版本**: Docker Compose v2.x, go-sdk v1.2.0
