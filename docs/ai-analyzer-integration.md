# AI-WAF 前端集成说明

## 概述

AI-WAF 已经完成了前端和后端的 AI 分析功能集成，提供了完整的攻击模式检测和规则生成能力。

## 已实现的功能

### 1. **前端页面** (`/ai-analyzer`)

#### 📊 攻击模式检测 (`/ai-analyzer/patterns`)
- 显示所有检测到的攻击模式
- 按严重级别过滤（Critical/High/Medium/Low）
- 查看模式详情（频率、置信度、特征）
- 删除误报模式

#### 🛡️ 生成规则管理 (`/ai-analyzer/rules`)
- 显示所有AI生成的防护规则
- 规则状态管理：
  - **待审核** (Pending): 等待人工审核
  - **已批准** (Approved): 已通过审核
  - **已拒绝** (Rejected): 审核未通过
  - **已部署** (Deployed): 已应用到WAF
- 规则审核流程：
  - 查看规则内容（SecLang）
  - 批准/拒绝规则
  - 一键部署已批准的规则

#### ⚙️ AI配置 (`/ai-analyzer/config`)
- 启用/禁用 AI 分析
- 配置分析间隔
- 设置检测阈值
- 配置最小样本数
- 规则生成参数

### 2. **后端API** (`/api/v1/ai-analyzer`)

所有API端点已实现：

```
GET    /patterns              # 列出攻击模式
GET    /patterns/:id          # 获取模式详情
DELETE /patterns/:id          # 删除模式

GET    /rules                 # 列出生成的规则
GET    /rules/:id             # 获取规则详情
DELETE /rules/:id             # 删除规则
POST   /rules/review          # 审核规则
GET    /rules/pending         # 获取待审核规则
POST   /rules/:id/deploy      # 部署规则

GET    /config                # 获取配置
PUT    /config                # 更新配置

GET    /conversations         # MCP对话记录
GET    /conversations/:id     # 对话详情
DELETE /conversations/:id     # 删除对话

GET    /stats                 # 统计信息
```

### 3. **MCP Server** (独立服务)

MCP Server 提供 6 个工具供 Claude Desktop 使用：

1. **analyze_attack_patterns** - 分析攻击模式
2. **generate_waf_rules** - 生成WAF规则
3. **get_waf_statistics** - 获取WAF统计
4. **get_attack_pattern** - 获取特定模式
5. **list_generated_rules** - 列出生成的规则
6. **get_recent_attacks** - 获取最近攻击

## 使用流程

### 方式一：Web界面（推荐日常使用）

1. **访问 AI 分析页面**
   ```
   http://your-domain/ai-analyzer
   ```

2. **配置 AI 引擎**
   - 进入"配置"页面
   - 启用AI分析
   - 设置分析间隔（建议30-60分钟）
   - 保存配置

3. **查看检测结果**
   - 进入"攻击模式检测"页面
   - 查看自动检测到的攻击模式
   - 点击模式查看详细信息

4. **审核生成的规则**
   - 进入"生成规则管理"页面
   - 查看待审核规则列表
   - 审核规则内容
   - 批准或拒绝规则

5. **部署规则**
   - 对已批准的规则点击"部署"
   - 规则将自动应用到WAF

### 方式二：Claude Desktop（AI辅助分析）

1. **配置 Claude Desktop**
   
   编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`:
   
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
         ]
       }
     }
   }
   ```

2. **重启 Claude Desktop**

3. **使用 AI 工具**
   
   在对话中使用：
   ```
   请分析最近24小时的攻击模式
   
   为这个攻击模式生成防护规则：[pattern_id]
   
   查看WAF系统的统计信息
   ```

## 自动化任务

系统会自动执行以下任务：

### 每小时任务
- 检测新的攻击模式
- 为高危模式自动生成规则（Severity: High/Critical）
- 将规则状态设置为"待审核"

### 每日任务（凌晨2点）
- 清理30天前的已拒绝规则
- 统计分析报告

## 数据流程

```
WAF日志 → AI检测引擎 → 攻击模式
                ↓
         规则生成器 → 待审核规则
                ↓
         人工审核 → 已批准规则
                ↓
         一键部署 → 生效规则
```

## 技术架构

```
┌─────────────────────────────────────┐
│         Web Frontend (React)        │
│  - Pattern List                     │
│  - Rule Management                  │
│  - Config Panel                     │
└──────────────┬──────────────────────┘
               │ HTTP API
┌──────────────▼──────────────────────┐
│      Backend Server (Go)            │
│  - REST API                         │
│  - AI Engine Service                │
│  - Cron Jobs                        │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴────────────┐
    │                       │
┌───▼────┐          ┌───────▼─────────┐
│MongoDB │          │ MCP Server      │
│        │          │ (STDIO/Claude)  │
└────────┘          └─────────────────┘
```

## 监控和日志

### 查看MCP Server日志
```bash
docker logs -f ai-waf-mcp-server
```

### 查看后端日志
```bash
docker logs -f ai-waf-mrya
```

### 查看AI分析统计
访问: `http://your-domain/ai-analyzer/config`

## 故障排查

### 1. 前端显示"无数据"
- 检查后端服务是否运行
- 确认MongoDB连接正常
- 查看是否有WAF日志数据

### 2. 规则生成失败
- 检查攻击模式数量是否满足最小样本数
- 查看AI引擎日志
- 确认置信度阈值设置合理

### 3. MCP Server 无响应
- 检查容器状态: `docker ps | grep mcp-server`
- 查看日志: `docker logs ai-waf-mcp-server`
- 重启容器: `docker compose restart mcp-server`

## 下一步优化

1. **添加实时通知** - 新检测到高危模式时推送通知
2. **规则效果评估** - 统计规则拦截效果
3. **模型训练** - 支持自定义模型训练
4. **批量操作** - 批量审核和部署规则
5. **可视化仪表板** - 更丰富的数据可视化

## 参考文档

- [MCP Server 配置](../docs/claude-desktop-setup.md)
- [API 文档](http://your-domain/swagger)
- [系统架构](../README.md)
