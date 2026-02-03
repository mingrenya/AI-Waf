# 错误处理改进文档
## Enhanced Error Handling Implementation

**日期**: 2025-02-03  
**目标**: 实现更友好的错误消息，提供具体的解决建议和下一步操作指导

---

## 一、改进概述

根据 MCP Builder Skill 的最佳实践，我们为 AI-WAF MCP Server 实现了增强的错误处理机制，使错误消息更加友好、可操作，并提供具体的解决建议。

### 核心原则

1. **可操作性** - 错误消息应指导 AI 代理和用户找到解决方案
2. **上下文感知** - 根据不同错误类型提供特定建议
3. **友好性** - 避免技术术语堆砌，提供清晰的中文说明
4. **一致性** - 所有工具使用统一的错误处理模式

---

## 二、实现的错误类型

### 1. MCPError 结构

创建了新的 `errors.go` 文件，定义了标准的 MCP 错误类型：

```go
type MCPError struct {
    Type        ErrorType `json:"type"`        // 错误类型
    Message     string    `json:"message"`     // 错误消息
    Suggestion  string    `json:"suggestion"`  // 解决建议
    HTTPStatus  int       `json:"http_status"` // HTTP 状态码
    OriginalErr error     `json:"-"`          // 原始错误
}
```

### 2. 错误类型常量

定义了 8 种错误类型：

| 错误类型 | 常量 | 用途 | 建议示例 |
|---------|------|------|---------|
| **认证错误** | `ErrorTypeAuth` | Token 失效、未授权 | 请检查 Authorization token 是否正确配置。可以使用 `scripts/get-mcp-token.sh` 获取新 token。 |
| **未找到** | `ErrorTypeNotFound` | 资源不存在 | 请确认资源 ID 是否正确，或使用列表工具查看所有可用资源。 |
| **验证错误** | `ErrorTypeValidation` | 参数不合法 | 请检查请求参数是否符合要求。可以查看工具的 JSON Schema 了解详细约束。 |
| **网络错误** | `ErrorTypeNetwork` | 连接失败、DNS 错误 | 请检查网络连接状态和 WAF 后端服务是否正常运行。 |
| **权限错误** | `ErrorTypePermission` | 无操作权限 | 请检查当前用户是否有足够的权限。可能需要管理员权限。 |
| **速率限制** | `ErrorTypeRateLimit` | API 调用过频 | 请求过于频繁，请稍后再试。 |
| **服务器错误** | `ErrorTypeServer` | 后端服务异常 | 这是服务器端问题。请检查 WAF 后端服务日志，或联系管理员。 |
| **超时错误** | `ErrorTypeTimeout` | 请求超时 | 操作超时。请检查网络连接，或对于大量数据操作，考虑分批处理。 |

---

## 三、核心错误处理函数

### 1. 创建特定类型错误

```go
// 认证错误
NewAuthError(message string, err error) error

// 资源未找到
NewNotFoundError(resource string, identifier string, err error) error

// 验证错误（增强版，带建议）
NewValidationErrorWithSuggestion(field, message, suggestion string) error

// 网络错误
NewNetworkError(operation string, err error) error

// 权限错误
NewPermissionError(action string, err error) error

// 速率限制错误
NewRateLimitError(retryAfter string, err error) error

// 服务器错误
NewServerError(operation string, err error) error

// 超时错误
NewTimeoutError(operation string, err error) error
```

### 2. 格式化和包装函数

```go
// 根据 HTTP 状态码自动格式化错误
FormatAPIError(operation string, statusCode int, responseBody []byte, originalErr error) error

// 格式化 JSON 解析错误
FormatParseError(dataType string, err error) error

// 包装错误并添加上下文
WrapError(err error, context string) error
```

---

## 四、已更新的文件

### 1. 新增文件

#### `mcp-server/tools/errors.go` (247 行)
- 定义了所有错误类型和创建函数
- 实现了智能错误格式化逻辑
- 包含超时检测、错误分类等辅助函数

### 2. 更新的文件

#### `mcp-server/tools/client.go`
**改进内容**:
- ✅ GET 请求错误处理（使用 `NewNetworkError` 和 `FormatAPIError`）
- ✅ POST 请求错误处理
- ✅ PATCH 请求错误处理
- ✅ PUT 请求错误处理
- ✅ DELETE 请求错误处理

**变化示例**:
```go
// 之前
if err != nil {
    return nil, fmt.Errorf("API错误 %d: %s", resp.StatusCode, string(body))
}

// 现在
if err != nil {
    return nil, FormatAPIError("GET "+path, resp.StatusCode, body, nil)
}
```

#### `mcp-server/tools/helpers.go`
**改进内容**:
- ✅ `GetPaginatedList` 使用 `WrapError` 和 `FormatParseError`
- ✅ `ParseAPIResponse` 使用 `FormatParseError`

#### `mcp-server/tools/logs.go`
**改进内容**:
- ✅ `ListAttackLogsInput.Validate()` 使用 `NewValidationErrorWithSuggestion`
- ✅ API 调用错误使用 `WrapError`
- ✅ JSON 解析错误使用 `FormatParseError`

**示例**:
```go
// 验证错误 - 之前
if input.PageSize > 100 {
    return NewValidationError("pageSize", "不能超过100")
}

// 验证错误 - 现在
if input.PageSize > 100 {
    return NewValidationErrorWithSuggestion(
        "pageSize",
        "不能超过100",
        "请将 pageSize 设置为 1-100 之间的值。对于大量数据，建议使用分页多次请求。",
    )
}
```

#### `mcp-server/tools/rules.go`
**改进内容**:
- ✅ `ListMicroRulesInput.Validate()` 使用 `NewValidationErrorWithSuggestion`
- ⏳ 其他操作函数的错误处理更新（待完成）

---

## 五、验证错误建议示例

### 分页参数验证

| 字段 | 错误消息 | 建议 |
|------|---------|------|
| `pageSize` | 不能超过100 | 请将 pageSize 设置为 1-100 之间的值。对于大量数据，建议使用分页多次请求。 |
| `page` | 必须大于0 | 页码从 1 开始。如需获取第一页数据，请将 page 设置为 1。 |
| `size` | 不能超过100 | 每页最多返回 100 条规则。对于更多数据，请使用分页获取。 |

### 规则参数验证

| 字段 | 错误消息 | 建议 |
|------|---------|------|
| `name` | 规则名称不能为空 | 请为规则提供一个描述性的名称，便于后续管理和识别。 |
| `type` | 必须为 blacklist 或 whitelist | 规则类型只能是 "blacklist"（黑名单）或 "whitelist"（白名单）。 |
| `priority` | 必须在 100-1000 之间 | 优先级范围为 100-1000，数值越小优先级越高。建议常规规则使用 500。 |

---

## 六、错误处理最佳实践

### 1. 网络和 HTTP 错误

```go
// API 请求
data, err := client.GetWithContext(ctx, path)
if err != nil {
    return WrapError(err, "查询日志")  // 自动判断是否为网络/超时错误
}
```

### 2. JSON 解析错误

```go
if err := json.Unmarshal(data, &result); err != nil {
    return FormatParseError("响应", err)  // 提供版本兼容性建议
}
```

### 3. 验证错误

```go
if input.Priority < 100 || input.Priority > 1000 {
    return NewValidationErrorWithSuggestion(
        "priority",
        "必须在 100-1000 之间",
        "优先级范围为 100-1000，数值越小优先级越高。建议常规规则使用 500。",
    )
}
```

### 4. API 错误（自动分类）

```go
if resp.StatusCode >= 400 {
    return FormatAPIError("创建规则", resp.StatusCode, body, nil)
    // 根据状态码自动返回：认证错误、权限错误、未找到、服务器错误等
}
```

---

## 七、未来改进计划

### 待更新文件

以下文件仍需要应用新的错误处理机制：

- [ ] `tools/blocked_ips.go` - 封禁 IP 相关工具
- [ ] `tools/batch_operations.go` - 批量操作工具
- [ ] `tools/ai_analyzer.go` - AI 分析工具
- [ ] `tools/monitoring.go` - 监控工具
- [ ] `tools/security_reports.go` - 安全报告工具
- [ ] `tools/compliance.go` - 合规检查工具
- [ ] 其他工具文件（共 14 个 .go 文件）

### 改进步骤

1. **批量替换** - 使用正则表达式替换所有 `fmt.Errorf`
   - 查询/获取/创建/更新/删除操作 → `WrapError`
   - 解析错误 → `FormatParseError`
   
2. **验证增强** - 为所有 `Validate()` 方法添加 `Suggestion`
   - 分页参数
   - IP 地址格式
   - 规则条件语法
   
3. **测试** - 创建错误处理测试用例
   - 单元测试验证错误类型
   - 集成测试验证建议内容
   - 确保错误消息对 AI 代理友好

---

## 八、测试建议

### 1. 认证错误测试

```bash
# 使用无效 token
export AI_WAF_TOKEN="invalid_token"
# 调用任意 MCP 工具，应返回认证错误及获取新 token 的建议
```

### 2. 验证错误测试

```bash
# 测试 pageSize 超限
# Input: {"page": 1, "pageSize": 150}
# Expected: 错误消息包含 "1-100 之间" 和 "分页多次请求" 建议
```

### 3. 网络错误测试

```bash
# 停止 WAF 后端服务
# 调用任意查询工具
# Expected: 网络错误及检查服务状态的建议
```

---

## 九、参考资料

- **MCP Builder Skill**: `~/.copilot/skills/mcp-builder/SKILL.md`
- **Go MCP 最佳实践**: `mcp-server/GO_MCP_BEST_PRACTICES.md`
- **代码审计报告**: `mcp-server/CODE_AUDIT_REPORT.md`

---

## 十、总结

### 已完成 ✅

1. ✅ 创建了完整的错误处理框架（`errors.go`）
2. ✅ 定义了 8 种标准错误类型，每种都有特定建议
3. ✅ 更新了 HTTP 客户端（`client.go`）的所有请求方法
4. ✅ 更新了辅助函数（`helpers.go`）的错误处理
5. ✅ 部分更新了日志工具（`logs.go`）和规则工具（`rules.go`）
6. ✅ 编译通过，无错误

### 进行中 ⏳

- ⏳ 批量更新其余 10+ 个工具文件
- ⏳ 为所有验证函数添加详细建议

### 待开始 ⬜

- ⬜ 创建错误处理单元测试
- ⬜ 编写错误消息风格指南
- ⬜ 集成到 CI/CD 流程中

---

**建议下一步**: 
1. 完成剩余工具文件的错误处理更新
2. 创建自动化测试验证错误消息质量
3. 进入下一个任务：创建 10 个复杂的评估问题

