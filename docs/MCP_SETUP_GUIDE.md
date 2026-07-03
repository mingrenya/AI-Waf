# AI-Waf MCP Server 使用指南

## ✅ 当前状态
- ✅ 所有服务已启动
- ✅ 后端运行在 http://localhost:2333
- ✅ MCP Server容器已创建
- ⚠️ 需要配置API Token

## 📋 下一步操作

### 1. 获取API Token

1. 打开浏览器访问：http://localhost:2333
2. 使用默认账号登录（首次需要注册）
3. 进入"系统设置" → "用户管理"
4. 创建一个服务账号（用于MCP Server）
5. 复制生成的JWT Token

### 2. 配置环境变量

编辑 `/Users/duheling/Downloads/AI-Waf/.env` 文件：

```bash
# 将复制的token粘贴到这里
MCP_API_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...your-token-here
```

然后重启MCP Server：
```bash
cd /Users/duheling/Downloads/AI-Waf
docker compose restart mcp-server
```

### 3. 配置Claude Desktop

#### macOS/Linux
编辑文件：`~/.config/Claude/claude_desktop_config.json`

#### Windows
编辑文件：`%APPDATA%\Claude\claude_desktop_config.json`

#### 配置内容

**方式A：使用Docker容器**
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

**方式B：使用本地编译版本**
```bash
# 先编译
cd /Users/duheling/Downloads/AI-Waf/mcp-server
make build
```

然后配置：
```json
{
  "mcpServers": {
    "ai-waf": {
      "command": "/Users/duheling/Downloads/AI-Waf/mcp-server/ai-waf-mcp",
      "env": {
        "WAF_BACKEND_URL": "http://localhost:2333",
        "WAF_API_TOKEN": "your-token-here"
      }
    }
  }
}
```

### 4. 重启Claude Desktop

配置完成后，完全退出Claude Desktop并重新打开。

### 5. 测试MCP工具

在Claude中输入：

```
帮我查看最近1小时的攻击日志
```

或：

```
列出所有MicroRule规则
```

Claude会自动调用MCP Server提供的工具。

## 🛠️ 可用工具列表

### 日志查询
- `list_attack_logs` - 查询攻击日志
- `get_log_stats` - 获取攻击统计

### 规则管理
- `list_micro_rules` - 列出规则
- `create_micro_rule` - 创建规则
- `update_micro_rule` - 更新规则
- `delete_micro_rule` - 删除规则

### IP管理
- `list_blocked_ips` - 列出封禁IP
- `get_blocked_ip_stats` - 封禁统计

### 站点管理
- `list_sites` - 列出站点
- `get_site_details` - 站点详情

### AI分析
- `list_attack_patterns` - 攻击模式
- `list_generated_rules` - AI生成的规则
- `trigger_ai_analysis` - 触发分析
- `review_rule` - 审核规则
- `deploy_rule` - 部署规则

## 🐛 故障排查

### MCP Server无法连接

检查容器状态：
```bash
docker compose ps mcp-server
docker compose logs mcp-server
```

### Claude看不到工具

1. 确认配置文件路径正确
2. 检查JSON格式是否正确
3. 完全退出并重启Claude Desktop
4. 检查WAF后端是否运行正常

### API调用失败

1. 检查API Token是否有效
2. 确认后端服务运行正常：`curl http://localhost:2333/api/v1/health`
3. 查看MCP Server日志了解详细错误

## 📚 使用示例

### 示例1：查看最近攻击
```
User: 帮我查看最近1小时的SQL注入攻击

Claude: [调用 list_attack_logs(hours=1, type="sql_injection")]
发现23次SQL注入尝试：
- 来源IP: 192.168.1.100 (15次)
- 来源IP: 10.0.0.50 (8次)
...
```

### 示例2：创建封禁规则
```
User: 帮我创建一条规则，封禁IP 192.168.1.100

Claude: [调用 create_micro_rule(...)]
已创建规则：
- 规则ID: 507f1f77bcf86cd799439011
- 规则名称: Block 192.168.1.100
- 类型: blacklist
- 状态: 已启用
```

### 示例3：触发AI分析
```
User: 触发AI分析任务

Claude: [调用 trigger_ai_analysis()]
AI分析任务已启动，预计需要2-5分钟完成。
系统将自动检测攻击模式并生成防护规则。
```

## 🔒 安全建议

1. **保护API Token**：不要将token提交到git仓库
2. **定期轮换**：定期更新API Token
3. **权限最小化**：为MCP Server创建专门的服务账号，只授予必要权限
4. **网络隔离**：确保MCP Server只能访问本地后端

## 📝 注意事项

- MCP Server通过HTTP API调用后端，不直接访问数据库
- 所有操作都需要有效的JWT Token认证
- 工具调用受后端权限系统控制
- 日志和统计数据实时查询，可能有延迟
