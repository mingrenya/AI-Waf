# AI助手（DeepSeek）集成指南

## 概述

AI助手功能集成了DeepSeek API，提供智能对话功能，帮助用户分析WAF日志、生成防护规则、评估安全威胁。

## 配置步骤

### 1. 获取DeepSeek API Key

1. 访问 [DeepSeek官网](https://platform.deepseek.com/)
2. 注册并登录账号
3. 在API Keys页面创建新的API Key
4. 复制API Key备用

### 2. 配置环境变量

编辑 `.env` 文件，添加DeepSeek配置：

```bash
# DeepSeek AI助手配置
DEEPSEEK_API_KEY=sk-your-api-key-here
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-chat
```

配置说明：
- `DEEPSEEK_API_KEY`: 必填，从DeepSeek平台获取
- `DEEPSEEK_BASE_URL`: 可选，默认为官方API地址
- `DEEPSEEK_MODEL`: 可选，推荐使用 `deepseek-chat`

**注意**: `.env` 文件已包含所有必要的配置项，只需填写实际值即可。

### 3. 重启服务

```bash
# 重启所有服务使配置生效
docker compose down
docker compose up -d
```

### 4. 验证配置

1. 打开浏览器访问 WAF Web UI
2. 点击右上角的 "AI 助手" 按钮
3. 在对话框中输入问题，例如："你好"
4. 如果配置正确，AI助手会正常响应

## 功能说明

### 支持的对话类型

1. **安全分析**
   - 分析最近的攻击模式
   - 识别高频攻击类型
   - 评估安全威胁等级

2. **规则生成**
   - 为特定攻击生成防护规则
   - 优化现有规则
   - 规则效果评估

3. **日志查询**
   - 查询特定时间段的日志
   - 分析攻击来源
   - 统计攻击趋势

4. **系统状态**
   - 查看当前系统状态
   - 检查配置问题
   - 获取运行建议

### API端点

- **普通聊天**: `POST /api/v1/ai/chat`
- **流式聊天**: `POST /api/v1/ai/chat/stream`

请求格式：

```json
{
  "message": "用户的问题",
  "messages": [
    {
      "role": "user",
      "content": "历史消息1"
    },
    {
      "role": "assistant",
      "content": "历史回复1"
    }
  ],
  "stream": false
}
```

响应格式：

```json
{
  "data": {
    "message": "AI的回复",
    "toolCalls": ["tool1", "tool2"],
    "timestamp": "2026-02-03T12:00:00Z"
  }
}
```

## 模型说明

### 可用模型

- **deepseek-chat** (推荐): DeepSeek-V3.2 非思考模式，快速响应
- **deepseek-reasoner**: DeepSeek-V3.2 思考模式，深度推理

### 价格

参考 [DeepSeek定价页面](https://api-docs.deepseek.com/zh-cn/quick_start/pricing)

- Input: ¥1/百万 tokens (缓存命中 ¥0.1/百万)
- Output: ¥2/百万 tokens

## 故障排查

### 问题1：AI助手无响应

**原因**: 未配置API Key或配置错误

**解决方案**:
1. 检查环境变量是否设置正确
2. 验证API Key是否有效
3. 查看服务日志: `docker compose logs server`

### 问题2：显示"未配置DEEPSEEK_API_KEY"

**原因**: 环境变量未传递到容器

**解决方案**:
1. 确认docker-compose.yaml中配置了环境变量
2. 重启服务: `docker compose restart server`
3. 检查容器环境变量: `docker compose exec server env | grep DEEPSEEK`

### 问题3：请求超时

**原因**: 网络连接问题或API服务不可达

**解决方案**:
1. 检查网络连接
2. 尝试设置代理（如需要）
3. 调整超时时间（服务默认60秒）

### 问题4：流式响应不工作

**原因**: 代理服务器可能缓冲了响应

**解决方案**:
1. 使用非流式模式（默认）
2. 配置代理禁用缓冲

## 安全建议

1. **保护API Key**: 不要将API Key提交到版本控制系统（.env已在.gitignore中）
2. **使用强密码**: 修改.env中的默认密码
3. **限制访问**: 只允许授权用户访问AI助手功能
4. **监控使用**: 定期检查API使用情况和费用

## 高级配置

### 自定义系统提示词

修改 `server/service/ai_chat.go` 中的系统提示：

```go
messages = append(messages, deepseekChatMessage{
    Role:    "system",
    Content: "你的自定义系统提示词",
})
```

### 调整超时时间

修改 `server/service/ai_chat.go` 中的超时设置：

```go
httpClient: &http.Client{
    Timeout: 120 * time.Second, // 改为120秒
},
```

### 使用其他兼容API

DeepSeek API兼容OpenAI格式，可以通过修改`DEEPSEEK_BASE_URL`使用其他兼容服务：

```bash
DEEPSEEK_BASE_URL=https://your-compatible-api.com
```

## 相关文档

- [DeepSeek API文档](https://api-docs.deepseek.com/zh-cn/)
- [DeepSeek模型说明](https://api-docs.deepseek.com/zh-cn/quick_start/pricing)
- [OpenAI API兼容性](https://api-docs.deepseek.com/zh-cn/)
