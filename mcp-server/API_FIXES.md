# MCP Server 修正说明

## 已完成的修改

### 1. 日志系统增强 ✅
- 创建了 `tools/logger.go` 统一日志记录器
- 所有 HTTP 请求/响应都记录详细日志
- 工具调用记录：参数、耗时、结果

### 2. API 路径修正 ✅

根据 `server/controller/*.go` 中的实际路由，已修正：

#### 日志相关 (logs.go)
- ✅ 攻击日志查询: `/api/v1/waf/logs`
- ⚠️ 日志统计: `/api/waf-logs/stats` → 需要确认实际API

#### 站点相关 (sites.go)
- ✅ 站点列表: `/api/v1/site`
- ✅ 站点详情: `/api/v1/site/{id}`

#### 规则相关 (rules.go)
- ✅ 规则列表: `/api/v1/micro-rules`
- ⚠️ 创建规则: 需要确认
- ⚠️ 更新规则: 需要确认
- ⚠️ 删除规则: 需要确认

#### IP封禁相关 (blocked_ips.go)
- ✅ 封禁IP列表: `/api/v1/blocked-ips`
- ⚠️ IP统计: `/api/v1/blocked-ips/stats` → 需要确认

#### AI分析相关 (ai_analyzer.go)
- ⚠️ 所有端点需要根据 `server/controller/ai_analyzer.go` 修正

### 3. 待修正的工具

需要继续修正以下文件中的API路径：

1. **rules.go** - 创建/更新/删除规则
2. **blocked_ips.go** - IP统计接口
3. **ai_analyzer.go** - 所有AI分析相关接口
4. **logs.go** - 日志统计接口

### 4. 实际 API 路由对照表

根据 `server/controller/*.go` 中的 `@Router` 注解：

```
WAF日志:
- GET /api/v1/waf/logs/events - 日志事件查询
- GET /api/v1/waf/logs - 日志列表

站点:
- POST /api/v1/site - 创建站点
- GET /api/v1/site - 站点列表
- GET /api/v1/site/{id} - 站点详情
- PUT /api/v1/site/{id} - 更新站点
- DELETE /api/v1/site/{id} - 删除站点

MicroRule规则:
- POST /api/v1/micro-rules - 创建规则
- GET /api/v1/micro-rules - 规则列表
- GET /api/v1/micro-rules/{id} - 规则详情
- PUT /api/v1/micro-rules/{id} - 更新规则
- DELETE /api/v1/micro-rules/{id} - 删除规则

封禁IP:
- GET /api/v1/blocked-ips - 封禁IP列表
- GET /api/v1/blocked-ips/stats - IP统计
- DELETE /api/v1/blocked-ips/cleanup - 清理过期记录

AI分析:
- GET /api/v1/ai-analyzer/patterns - 攻击模式列表
- GET /api/v1/ai-analyzer/patterns/{id} - 模式详情
- DELETE /api/v1/ai-analyzer/patterns/{id} - 删除模式
- GET /api/v1/ai-analyzer/rules - 生成的规则列表
- GET /api/v1/ai-analyzer/rules/{id} - 规则详情
- DELETE /api/v1/ai-analyzer/rules/{id} - 删除规则
- POST /api/v1/ai-analyzer/rules/review - 审核规则
- GET /api/v1/ai-analyzer/rules/pending - 待审核规则
- POST /api/v1/ai-analyzer/rules/{id}/deploy - 部署规则
- GET /api/v1/ai-analyzer/config - 获取配置
- PUT /api/v1/ai-analyzer/config - 更新配置
- POST /api/v1/ai-analyzer/trigger - 触发分析
- GET /api/v1/ai-analyzer/stats - 统计信息
```

### 5. 日志输出格式

修改后的日志格式示例：
```
[工具调用] list_attack_logs 开始执行
[工具参数] list_attack_logs - 输入: {"hours":24,"limit":50}
[API请求] GET http://localhost:2333/api/v1/waf/logs?page=1&pageSize=50
[API响应] GET /api/v1/waf/logs - 状态码: 200 - 耗时: 45ms - 响应大小: 2048 bytes
[工具成功] list_attack_logs - 返回 10 条日志 - 耗时: 50ms
```

### 6. 下一步工作

1. **完成所有工具的API路径修正**
2. **添加日志记录到所有工具函数**
3. **重新编译和测试**
4. **验证每个工具是否正常工作**

## 编译和测试

```bash
# 重新编译
cd /Users/duheling/Downloads/AI-Waf/mcp-server
go build -o ai-waf-mcp .

# 或使用 Docker
cd /Users/duheling/Downloads/AI-Waf
docker compose build mcp-server
docker compose up -d --force-recreate mcp-server

# 查看日志
docker logs -f ai-waf-mcp-server
```

## 测试建议

在 AnythingLLM 中测试：
1. "显示所有站点列表" → 测试 list_sites
2. "查看最近的攻击日志" → 测试 list_attack_logs  
3. "列出所有MicroRule规则" → 测试 list_micro_rules
4. "显示封禁的IP列表" → 测试 list_blocked_ips
