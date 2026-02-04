# Gin 中间件审计工具

## 简介

这是一个用于审计 Gin 框架路由配置的自动化工具，它可以检查你的路由是否正确使用了必要的安全中间件。

## 功能特性

✅ **自动扫描路由文件** - 自动查找并分析所有路由配置  
✅ **安全规则检查** - 检查是否遵循安全最佳实践  
✅ **详细报告** - 按严重程度分类显示问题  
✅ **实用建议** - 为每个问题提供具体的修复建议

## 审计规则

### 🔴 高严重度（HIGH）

1. **ID参数验证**
   - 带 `:id` 的路由必须验证 ID 格式
   - 建议使用: `ValidateMongoID()` 或 `ValidateUUID()`

2. **DELETE操作权限**
   - 所有 DELETE 操作必须有权限检查
   - 建议使用: `HasPermission("resource:delete")`

3. **JWT认证**
   - `/api` 下的路由必须有 JWT 认证（除了公开接口）
   - 建议使用: `JWTAuth()`

### 🟡 中严重度（MEDIUM）

1. **Content-Type验证**
   - POST/PUT/PATCH 请求应验证 Content-Type
   - 建议使用: `ValidateJSONContentType()`

2. **敏感接口缓存**
   - 敏感接口应禁用缓存
   - 建议使用: `NoCache()`

### 🟢 低严重度（LOW）

1. **分页参数验证**
   - 列表查询接口应验证分页参数
   - 建议使用: `ValidatePagination()`

## 使用方法

### 方式一：使用脚本（推荐）

```bash
cd server/tools/middleware-audit
./run-audit.sh
```

### 方式二：手动运行

```bash
cd server/tools/middleware-audit
go build -o audit-tool main.go
cd ../..
./tools/middleware-audit/audit-tool
```

## 示例输出

```
🔍 Running Gin Middleware Audit...

=== Gin Middleware Audit Report ===

Total Issues: 5
  🔴 High:   2
  🟡 Medium: 2
  🟢 Low:    1

🔴 HIGH SEVERITY ISSUES
========================

1. [Security] DELETE /api/users/:id: Missing permission check
   File: router/user.go:45
   💡 Suggestion: Add middleware.HasPermission("user:delete")

2. [Security] POST /api/sites/:id: Missing ID format validation
   File: router/site.go:28
   💡 Suggestion: Add middleware.ValidateMongoID("id")

🟡 MEDIUM SEVERITY ISSUES
===========================

1. [Validation] POST /api/users: Missing Content-Type validation
   File: router/user.go:23
   💡 Suggestion: Add middleware.ValidateJSONContentType()

2. [Security] GET /api/profile: Sensitive endpoint should disable cache
   File: router/user.go:67
   💡 Suggestion: Add middleware.NoCache()

🟢 LOW SEVERITY ISSUES
========================

1. [Validation] GET /api/users: Missing pagination validation
   File: router/user.go:12
   💡 Suggestion: Add middleware.ValidatePagination()

✅ Audit complete!
```

## 修复示例

### 问题：缺少ID验证

**审计发现：**
```go
router.GET("/api/users/:id", userController.GetByID)
```

**修复后：**
```go
router.GET("/api/users/:id", 
    middleware.ValidateMongoID("id"),
    userController.GetByID)
```

### 问题：缺少Content-Type验证

**审计发现：**
```go
router.POST("/api/users", userController.Create)
```

**修复后：**
```go
router.POST("/api/users",
    middleware.ValidateJSONContentType(),
    userController.Create)
```

### 问题：缺少权限检查

**审计发现：**
```go
router.DELETE("/api/sites/:id", siteController.Delete)
```

**修复后：**
```go
router.DELETE("/api/sites/:id",
    middleware.ValidateMongoID("id"),
    middleware.HasPermission(model.PermSiteDelete),
    siteController.Delete)
```

## 集成到CI/CD

可以将审计工具集成到 CI/CD 流程中：

```yaml
# .github/workflows/security-audit.yml
name: Middleware Security Audit

on: [push, pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
      
      - name: Run Middleware Audit
        run: |
          cd server/tools/middleware-audit
          ./run-audit.sh
```

## 扩展审计规则

你可以在 `main.go` 中添加自定义审计规则：

```go
// 规则7: 自定义规则示例
if strings.Contains(path, "/admin") && !a.hasMiddleware(middlewares, "AdminOnly") {
    a.issues = append(a.issues, AuditIssue{
        File:        file,
        Line:        line,
        Severity:    "HIGH",
        Category:    "Security",
        Description: fmt.Sprintf("%s %s: Admin route without AdminOnly middleware", method, path),
        Suggestion:  "Add middleware.AdminOnly()",
    })
}
```

## 局限性

- 只能检测静态路由定义
- 无法分析动态注册的路由
- 需要遵循标准的 Gin 路由定义模式

## 贡献

欢迎提交问题和改进建议！

## 许可证

MIT License
