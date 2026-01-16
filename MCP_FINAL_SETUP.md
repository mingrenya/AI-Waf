# AI-Waf MCP Server 配置完成

## ✅ 当前状态
- ✅ 后端服务运行正常 (http://localhost:2333)
- ✅ API Token已配置
- ✅ MCP Server已重启

## 📋 最后一步：配置Claude Desktop

### 1. 找到配置文件位置

**macOS/Linux:**
```bash
~/.config/Claude/claude_desktop_config.json
```

**Windows:**
```
%APPDATA%\Claude\claude_desktop_config.json
```

### 2. 复制配置内容

将以下内容复制到配置文件中（如果文件不存在，创建它）：

```json
{
  "mcpServers": {
    "ai-waf": {
      "command": "docker",
      "args": [
        "exec",
        "-i",
        "ai-waf-mcp-server",
        "/app/ai-waf-mcp"
      ]
    }
  }
}
```

或者直接运行命令（macOS/Linux）：

```bash
mkdir -p ~/.config/Claude
cp /Users/duheling/Downloads/AI-Waf/claude_desktop_config_example.json ~/.config/Claude/claude_desktop_config.json
```

### 3. 重启Claude Desktop

- 完全退出Claude Desktop（Cmd+Q / 右键退出）
- 重新打开Claude Desktop

### 4. 测试MCP工具

在Claude中输入以下任一命令测试：

```
帮我查看最近1小时的攻击日志
```

```
列出所有MicroRule规则
```

```
显示WAF站点列表
```

## 🛠️ 可用的15个工具

### 日志查询
- `list_attack_logs` - 查询攻击日志（支持时间范围、类型、严重程度过滤）
- `get_log_stats` - 获取攻击统计信息

### 规则管理  
- `list_micro_rules` - 列出所有MicroRule规则
- `create_micro_rule` - 创建新规则
- `update_micro_rule` - 更新规则
- `delete_micro_rule` - 删除规则

### IP管理
- `list_blocked_ips` - 列出被封禁的IP
- `get_blocked_ip_stats` - 获取IP封禁统计

### 站点管理
- `list_sites` - 列出所有受保护站点
- `get_site_details` - 获取站点详细信息

### AI分析器
- `list_attack_patterns` - 列出AI检测到的攻击模式
- `list_generated_rules` - 列出AI生成的防护规则
- `trigger_ai_analysis` - 手动触发AI分析任务
- `review_rule` - 审核AI生成的规则
- `deploy_rule` - 部署规则到生产环境

## 💡 使用示例

**示例1：查看最近攻击**
```
User: 帮我查看最近2小时的SQL注入攻击

Claude会调用: list_attack_logs(hours=2, type="sql_injection")
返回: 攻击日志列表，包括来源IP、攻击时间、攻击详情等
```

**示例2：创建封禁规则**
```
User: 创建一条规则，封禁IP 192.168.1.100

Claude会调用: create_micro_rule(...)
返回: 规则创建成功，包含规则ID和详细信息
```

**示例3：AI分析**
```
User: 触发AI分析，检测最近的攻击模式

Claude会调用: trigger_ai_analysis(force=true)
返回: AI分析任务已启动的确认信息
```

## 🔍 故障排查

### 1. Claude看不到MCP工具

**检查清单：**
- [ ] 配置文件路径正确？
- [ ] JSON格式正确（没有多余逗号）？
- [ ] 完全退出并重启Claude Desktop？
- [ ] MCP Server容器运行正常？

```bash
# 检查容器状态
docker compose ps mcp-server

# 查看日志
docker compose logs mcp-server
```

### 2. 工具调用失败

**检查：**
- [ ] Token是否有效？
- [ ] 后端服务是否运行？

```bash
# 测试后端
curl http://localhost:2333/health

# 测试认证
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:2333/api/v1/sites
```

### 3. Token过期

Token有效期约为1个月，过期后需要重新获取：

```bash
# 重新登录获取token
curl -X POST http://localhost:2333/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 更新.env文件中的MCP_API_TOKEN
# 然后重启MCP Server
docker compose restart mcp-server
```

## 🎉 完成！

配置完成后，你就可以在Claude中通过自然语言与WAF系统交互了！

MCP Server会自动将你的请求转换为API调用，并返回结构化的结果。
