# AI-Waf MCP Server

AI-Waf的Model Context Protocol (MCP) Server实现，让AI应用（如Claude Desktop、Cursor等）能够通过标准MCP协议访问WAF功能。

## 📋 功能特性

### 支持的工具

#### 1. 日志查询
- `list_attack_logs` - 查询WAF攻击日志
- `get_log_stats` - 获取攻击统计信息

#### 2. 规则管理
- `list_micro_rules` - 列出MicroRule规则
- `create_micro_rule` - 创建新规则
- `update_micro_rule` - 更新规则
- `delete_micro_rule` - 删除规则

#### 3. IP封禁管理
- `list_blocked_ips` - 列出封禁IP
- `get_blocked_ip_stats` - 获取封禁统计

#### 4. 站点管理
- `list_sites` - 列出受保护站点
- `get_site_details` - 获取站点详情

#### 5. AI分析器
- `list_attack_patterns` - 列出检测到的攻击模式
- `list_generated_rules` - 列出AI生成的规则
- `trigger_ai_analysis` - 手动触发分析
- `review_rule` - 审核AI生成的规则
- `deploy_rule` - 部署规则到生产环境

## 🚀 快速开始

### 1. 编译

```bash
cd mcp-server
go mod download
go build -o ai-waf-mcp .
```

### 2. 配置环境变量

```bash
export WAF_BACKEND_URL="http://localhost:8080"
export WAF_API_TOKEN="your-jwt-token"
```

### 3. 在Claude Desktop中配置

编辑 `~/.config/Claude/claude_desktop_config.json` (macOS/Linux) 或 `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "ai-waf": {
      "command": "/path/to/AI-Waf/mcp-server/ai-waf-mcp",
      "env": {
        "WAF_BACKEND_URL": "http://localhost:8080",
        "WAF_API_TOKEN": "your-jwt-token-here"
      }
    }
  }
}
```

### 4. 重启Claude Desktop

重启Claude Desktop后，MCP Server会自动连接。

## 💡 使用示例

### 在Claude中使用

```
👤: 帮我查看最近1小时的攻击日志

🤖: [调用 list_attack_logs(hours=1)]
    发现137次攻击尝试:
    - SQL注入: 45次
    - XSS: 23次
    - 路径穿越: 12次
    主要来源IP: 192.168.1.100

👤: 创建一条规则拦截这个IP

🤖: [调用 create_micro_rule()]
    已创建规则:
    规则名称: Block 192.168.1.100
    规则类型: blacklist
    规则ID: 507f1f77bcf86cd799439011

👤: 触发AI分析任务

🤖: [调用 trigger_ai_analysis()]
    AI分析任务已启动，预计需要2-5分钟完成
```

## 🔧 开发

### 项目结构

```
mcp-server/
├── main.go              # MCP Server入口
├── tools/               # 工具实现
│   ├── client.go       # API客户端
│   ├── logs.go         # 日志查询工具
│   ├── rules.go        # 规则管理工具
│   ├── blocked_ips.go  # IP封禁工具
│   ├── sites.go        # 站点管理工具
│   └── ai_analyzer.go  # AI分析器工具
├── go.mod
└── README.md
```

### 添加新工具

1. 在 `tools/` 目录创建新文件
2. 定义输入输出结构体
3. 实现工具函数
4. 在 `main.go` 中注册工具

示例:

```go
// tools/mytool.go
type MyToolInput struct {
    Param string `json:"param" jsonschema:"参数描述"`
}

type MyToolOutput struct {
    Result string `json:"result"`
}

func CreateMyTool(client *APIClient) func(context.Context, *mcp.CallToolRequest, MyToolInput) (*mcp.CallToolResult, MyToolOutput, error) {
    return func(ctx context.Context, req *mcp.CallToolRequest, input MyToolInput) (*mcp.CallToolResult, MyToolOutput, error) {
        // 实现逻辑
        return nil, MyToolOutput{Result: "success"}, nil
    }
}

// main.go
mcp.AddTool(server, &mcp.Tool{
    Name:        "my_tool",
    Description: "工具描述",
}, tools.CreateMyTool(client))
```

## 📚 API映射

| MCP工具 | 后端API | 说明 |
|---------|---------|------|
| list_attack_logs | GET /api/waf-logs/query | 查询攻击日志 |
| list_micro_rules | GET /api/rules/micro-rule | 查询MicroRule |
| create_micro_rule | POST /api/rules/micro-rule | 创建规则 |
| list_blocked_ips | GET /api/flow-control/blocked-ips | 查询封禁IP |
| list_attack_patterns | GET /api/ai-analyzer/patterns | 查询攻击模式 |
| trigger_ai_analysis | POST /api/ai-analyzer/trigger | 触发分析 |

## ⚠️ 注意事项

1. **认证**: 需要有效的JWT token才能访问后端API
2. **权限**: 不同操作需要对应的角色权限
3. **网络**: MCP Server需要能访问后端服务
4. **安全**: 不要在配置文件中明文存储敏感信息

## 🐛 故障排查

### MCP Server无法连接

检查:
1. 后端服务是否运行 (`curl http://localhost:8080/health`)
2. JWT token是否有效
3. 环境变量是否正确设置

### Claude中看不到工具

1. 确认配置文件路径正确
2. 重启Claude Desktop
3. 检查MCP Server日志输出

### API调用失败

1. 检查后端日志
2. 验证JWT token权限
3. 确认请求参数格式正确

## 📄 许可证

与AI-Waf项目相同
