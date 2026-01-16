# MCP 工具测试指南

## 📋 测试前准备

1. **确认服务运行:**
```bash
docker compose ps
# 确保 mrya、mongodb、mcp-server 都在运行
```

2. **确认 AnythingLLM 连接:**
- 在 AnythingLLM 界面看到 "AIWaf - 15 tools available"
- 日志中看到 JSON-RPC 消息

---

## 🧪 测试用例

### 1️⃣ 基础查询工具

#### **list_sites - 列出站点**
在 AnythingLLM 中输入:
```
显示所有受保护的站点列表
```

**预期结果:**
- 返回站点列表（JSON 格式）
- 包含域名、状态、配置信息

**调试命令:**
```bash
TOKEN="your-token-here"
curl -H "Authorization: Bearer $TOKEN" http://localhost:2333/api/v1/sites
```

---

#### **get_site_details - 获取站点详情**
```
获取站点 example.com 的详细信息
```

**预期结果:**
- 返回指定站点的完整配置
- 包括防护规则、流量统计

---

### 2️⃣ 日志查询工具

#### **list_attack_logs - 查询攻击日志**
```
帮我查看最近1小时的攻击日志
```

**预期结果:**
- 返回攻击日志列表
- 包含时间、IP、攻击类型、严重程度

**测试变体:**
```
查看最近24小时严重级别为高的攻击日志
查看来自特定IP (1.2.3.4) 的所有攻击
查看SQL注入类型的攻击
```

**调试命令:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:2333/api/v1/waf-logs?timeRange=1h&page=1&pageSize=10"
```

---

#### **get_log_stats - 获取日志统计**
```
获取攻击日志的统计信息
```

**预期结果:**
- 攻击类型分布（饼图数据）
- 来源IP TOP 10
- 时间趋势数据

**调试命令:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:2333/api/v1/waf-logs/stats
```

---

### 3️⃣ 规则管理工具

#### **list_micro_rules - 列出规则**
```
列出所有MicroRule规则
```

**预期结果:**
- 规则列表
- 包含规则名称、条件、动作、状态

**调试命令:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:2333/api/v1/micro-rules
```

---

#### **create_micro_rule - 创建规则**
```
创建一个MicroRule规则，阻止来自IP 1.2.3.4的请求
```

**预期结果:**
- 成功创建规则
- 返回规则ID

**调试命令:**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试规则",
    "condition": "ip == \"1.2.3.4\"",
    "action": "deny",
    "enabled": true
  }' \
  http://localhost:2333/api/v1/micro-rules
```

---

#### **update_micro_rule - 更新规则**
```
更新规则ID为 xxx 的状态为禁用
```

**预期结果:**
- 成功更新规则
- 返回更新后的规则信息

---

#### **delete_micro_rule - 删除规则**
```
删除规则ID为 xxx 的规则
```

**预期结果:**
- 成功删除规则
- 返回确认消息

---

### 4️⃣ IP管理工具

#### **list_blocked_ips - 列出封禁IP**
```
显示所有被封禁的IP地址
```

**预期结果:**
- IP列表
- 包含封禁原因、时间、过期时间

**调试命令:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:2333/api/v1/blocked-ips
```

---

#### **get_blocked_ip_stats - IP封禁统计**
```
获取IP封禁的统计信息
```

**预期结果:**
- 总封禁数量
- 按国家/地区分布
- 封禁原因分布

---

### 5️⃣ AI分析工具

#### **list_attack_patterns - 列出攻击模式**
```
列出AI检测到的攻击模式
```

**预期结果:**
- 攻击模式列表
- 包含模式特征、检测次数、置信度

**调试命令:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:2333/api/v1/ai-analyzer/patterns
```

---

#### **list_generated_rules - 列出生成的规则**
```
显示AI生成的防护规则
```

**预期结果:**
- AI生成的规则列表
- 包含规则内容、状态（待审核/已批准/已拒绝）

**调试命令:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:2333/api/v1/ai-analyzer/generated-rules
```

---

#### **trigger_ai_analysis - 触发AI分析**
```
手动触发一次AI分析任务
```

**预期结果:**
- 返回任务ID
- 异步处理攻击日志并生成规则建议

**调试命令:**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:2333/api/v1/ai-analyzer/analyze
```

---

#### **review_rule - 审核规则**
```
批准AI生成的规则ID为 xxx 的规则
```

**预期结果:**
- 规则状态更新为"已批准"
- 可以进行部署

**测试变体:**
```
拒绝规则ID为 xxx 的规则，原因是误报率过高
```

**调试命令:**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ruleId": "xxx",
    "action": "approve",
    "comment": "规则合理，批准部署"
  }' \
  http://localhost:2333/api/v1/ai-analyzer/review-rule
```

---

#### **deploy_rule - 部署规则**
```
将已审核通过的规则ID为 xxx 部署到生产环境
```

**预期结果:**
- 规则部署到HAProxy
- 返回部署状态

**调试命令:**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ruleId": "xxx"
  }' \
  http://localhost:2333/api/v1/ai-analyzer/deploy-rule
```

---

## 🐛 调试技巧

### 1. 查看 MCP Server 日志
```bash
docker logs -f ai-waf-mcp-server
```

### 2. 查看后端 API 日志
```bash
docker logs -f mrya-waf
```

### 3. 测试 API 直接调用
```bash
# 设置 Token
export TOKEN="eyJhbGci..."

# 测试各个端点
curl -H "Authorization: Bearer $TOKEN" http://localhost:2333/api/v1/sites
curl -H "Authorization: Bearer $TOKEN" http://localhost:2333/api/v1/waf-logs?page=1&pageSize=10
curl -H "Authorization: Bearer $TOKEN" http://localhost:2333/api/v1/micro-rules
```

### 4. 检查 MCP 工具调用
在 AnythingLLM 的日志中查看:
- 工具名称
- 传入参数
- 返回结果

---

## ✅ 测试检查清单

- [ ] list_sites - 站点列表
- [ ] get_site_details - 站点详情
- [ ] list_attack_logs - 攻击日志
- [ ] get_log_stats - 日志统计
- [ ] list_micro_rules - 规则列表
- [ ] create_micro_rule - 创建规则
- [ ] update_micro_rule - 更新规则
- [ ] delete_micro_rule - 删除规则
- [ ] list_blocked_ips - 封禁IP
- [ ] get_blocked_ip_stats - IP统计
- [ ] list_attack_patterns - 攻击模式
- [ ] list_generated_rules - 生成规则
- [ ] trigger_ai_analysis - 触发分析
- [ ] review_rule - 审核规则
- [ ] deploy_rule - 部署规则

---

## 📝 测试报告模板

### 工具名称: [tool_name]
- **测试时间:** 2026-01-15
- **测试输入:** "在 AnythingLLM 中输入的文本"
- **是否调用:** ✅/❌
- **返回结果:** 
  - 成功/失败
  - 数据格式是否正确
  - 响应时间
- **发现问题:**
  - 问题描述
  - 错误日志
- **改进建议:**

---

## 🚀 下一步

测试完成后：
1. 记录所有通过/失败的工具
2. 分析失败原因（API问题 vs MCP工具问题）
3. 优化错误处理
4. 添加更详细的工具描述和参数验证
