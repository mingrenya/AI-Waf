# AI-WAF 中间件使用审计报告

## 审计范围
- **审计文件**: [router/router.go](../../router/router.go)
- **审计日期**: 2025年
- **审计工具**: AST-based Middleware Auditor

## 审计结果总结

### ✅ 总体评价：优秀

经过详细代码审查，当前路由配置符合安全最佳实践，使用了正确的中间件架构。

## 中间件使用情况

### 1. 全局中间件（应用于所有路由）

```go
route.Use(middleware.RequestID())          // ✅ 请求追踪
route.Use(middleware.Logger())             // ✅ 日志记录
route.Use(middleware.Cors())               // ✅ CORS 配置
route.Use(gin.CustomRecovery(...))         // ✅ 错误恢复
```

### 2. 认证相关路由

#### 公开接口（登录）
```go
auth.POST("/login", middleware.RateLimit(5, time.Minute), ...)
auth.POST("/login-service", middleware.RateLimit(5, time.Minute), ...)
```
- ✅ **限流保护**: 5次/分钟，防止暴力破解
- ✅ **适当开放**: 登录接口不需要JWT

#### 需要认证的接口
```go
authRequired.Use(middleware.JWTAuth())
authRequired.Use(middleware.PasswordResetRequired())
```
- ✅ **JWT认证**: 所有认证后的接口都需要JWT
- ✅ **密码重置检查**: 强制重置初始密码

### 3. 业务路由组（统一应用中间件）

所有业务路由都通过以下方式应用中间件：
```go
authenticated := api.Group("")
authenticated.Use(middleware.JWTAuth())
authenticated.Use(middleware.PasswordResetRequired())
```

#### 用户管理模块
```go
userRoutes.POST("", middleware.HasPermission(model.PermUserCreate), ...)
userRoutes.GET("", middleware.HasPermission(model.PermUserRead), ...)
userRoutes.PUT("/:id", middleware.HasPermission(model.PermUserUpdate), ...)
userRoutes.DELETE("/:id", middleware.HasPermission(model.PermUserDelete), ...)
```
- ✅ **JWT**: 通过路由组应用
- ✅ **权限检查**: 每个操作都有细粒度权限
- ⚠️ **建议**: 添加 `ValidateMongoID("id")` 到带 `:id` 的路由

#### 站点管理模块
```go
siteRoutes.POST("", middleware.HasPermission(model.PermSiteCreate), ...)
siteRoutes.GET("/:id", middleware.HasPermission(model.PermSiteRead), ...)
siteRoutes.PUT("/:id", middleware.HasPermission(model.PermSiteUpdate), ...)
siteRoutes.DELETE("/:id", middleware.HasPermission(model.PermSiteDelete), ...)
```
- ✅ **JWT**: 通过路由组应用
- ✅ **权限检查**: 每个操作都有权限验证
- ⚠️ **建议**: 添加 ID 格式验证

#### 证书管理模块
```go
certRoutes.POST("", middleware.HasPermission(model.PermCertCreate), ...)
certRoutes.GET("/:id", middleware.HasPermission(model.PermCertRead), ...)
certRoutes.PUT("/:id", middleware.HasPermission(model.PermCertUpdate), ...)
certRoutes.DELETE("/:id", middleware.HasPermission(model.PermCertDelete), ...)
```
- ✅ **JWT + 权限**: 完整保护
- ⚠️ **建议**: 添加 ID 格式验证

#### AI分析器模块
```go
aiAnalyzerRoutes.GET("/patterns", middleware.HasPermission(model.PermWAFLogRead), ...)
aiAnalyzerRoutes.DELETE("/patterns/:id", middleware.HasPermission(model.PermConfigUpdate), ...)
aiAnalyzerRoutes.POST("/rules/review", middleware.HasPermission(model.PermConfigUpdate), ...)
```
- ✅ **JWT + 权限**: 完整保护
- ⚠️ **建议**: 添加 ID 格式验证到带 `:id` 的路由

#### MCP 服务模块
```go
mcpRoutes.GET("/status", mcpController.GetMCPStatus)
mcpRoutes.GET("/tools", mcpController.GetMCPTools)
mcpRoutes.GET("/tool-calls", middleware.HasPermission(model.PermWAFLogRead), ...)
```
- ✅ **JWT**: 通过路由组应用
- ✅ **适当权限**: 状态查询不需要额外权限，历史记录需要权限
- ✅ **合理设计**: 符合实际使用场景

## 优化建议

### 1. 🟡 中优先级：添加 ID 格式验证

为所有带 `:id` 参数的路由添加 ID 格式验证：

```go
// 当前
userRoutes.GET("/:id", middleware.HasPermission(model.PermUserRead), authController.GetUserByID)

// 建议
userRoutes.GET("/:id", 
    middleware.ValidateMongoID("id"),
    middleware.HasPermission(model.PermUserRead), 
    authController.GetUserByID)
```

**影响的路由（约30+）**:
- `/api/v1/users/:id` (GET, PUT, DELETE)
- `/api/v1/site/:id` (GET, PUT, DELETE)
- `/api/v1/certificate/:id` (GET, PUT, DELETE)
- `/api/v1/ip-groups/:id` (GET, PUT, DELETE)
- `/api/v1/micro-rules/:id` (GET, PUT, DELETE)
- `/api/v1/alerts/channels/:id` (GET, PUT, DELETE)
- `/api/v1/alerts/rules/:id` (GET, PUT, DELETE)
- `/api/v1/ai-analyzer/patterns/:id` (GET, DELETE)
- `/api/v1/ai-analyzer/rules/:id` (GET, DELETE, POST deploy)
- `/api/v1/ai-analyzer/conversations/:id` (GET, DELETE)

**收益**:
- ✅ 防止无效 ID 查询到达数据库
- ✅ 提早返回错误，节省资源
- ✅ 统一错误响应格式

### 2. 🟢 低优先级：添加 Content-Type 验证

为所有 POST/PUT/PATCH 接口添加 Content-Type 验证：

```go
// 当前
userRoutes.POST("", middleware.HasPermission(model.PermUserCreate), authController.CreateUser)

// 建议
userRoutes.POST("", 
    middleware.ValidateJSONContentType(),
    middleware.HasPermission(model.PermUserCreate), 
    authController.CreateUser)
```

**收益**:
- ✅ 防止错误的请求格式
- ✅ 提升API的健壮性

### 3. 🟢 低优先级：添加分页验证

为列表查询接口添加分页参数验证：

```go
// 当前
userRoutes.GET("", middleware.HasPermission(model.PermUserRead), authController.GetUsers)

// 建议  
userRoutes.GET("", 
    middleware.ValidatePagination(),
    middleware.HasPermission(model.PermUserRead), 
    authController.GetUsers)
```

**影响的路由**:
- `/api/v1/users` (GET)
- `/api/v1/site` (GET)
- `/api/v1/certificate` (GET)
- `/api/v1/ip-groups` (GET)
- `/api/v1/micro-rules` (GET)
- `/api/v1/log/event` (GET)
- `/api/v1/log` (GET)
- `/api/v1/alerts/channels` (GET)
- `/api/v1/alerts/rules` (GET)
- `/api/v1/alerts/history` (GET)

### 4. 🟢 可选：添加审计日志

为敏感操作添加审计日志中间件：

```go
// DELETE 操作
userRoutes.DELETE("/:id", 
    middleware.ValidateMongoID("id"),
    middleware.SecurityAudit(),  // 新增
    middleware.HasPermission(model.PermUserDelete), 
    authController.DeleteUser)

// 配置更新操作
configRoutes.PATCH("", 
    middleware.SecurityAudit(),  // 新增
    middleware.HasPermission(model.PermConfigUpdate), 
    configController.PatchConfig)
```

## 实施计划

### Phase 1: ID 验证（优先）

1. **创建辅助函数**
```go
// router/helpers.go
func IDRoute(id string, permission string, handler gin.HandlerFunc) []gin.HandlerFunc {
    return []gin.HandlerFunc{
        middleware.ValidateMongoID(id),
        middleware.HasPermission(permission),
        handler,
    }
}
```

2. **批量更新路由**
```go
// 简化前
userRoutes.GET("/:id", middleware.HasPermission(model.PermUserRead), authController.GetUserByID)

// 简化后
userRoutes.GET("/:id", IDRoute("id", model.PermUserRead, authController.GetUserByID)...)
```

### Phase 2: Content-Type 验证（次要）

创建包装函数统一处理：
```go
func CreateRoute(permission string, handler gin.HandlerFunc) []gin.HandlerFunc {
    return []gin.HandlerFunc{
        middleware.ValidateJSONContentType(),
        middleware.HasPermission(permission),
        handler,
    }
}
```

### Phase 3: 审计日志（可选）

为关键操作添加审计：
- 用户CRUD
- 站点CRUD  
- 证书管理
- 配置变更
- AI规则部署

## 现有安全措施总结

### ✅ 已正确实施的安全措施

1. **认证与授权**
   - JWT 全局应用于 `/api/v1` 路由组
   - 细粒度权限控制（40+ 不同权限）
   - 强制密码重置机制

2. **限流保护**
   - 登录接口：5次/分钟
   - 防止暴力破解

3. **错误处理**
   - 全局 Recovery 中间件
   - 自定义错误处理器

4. **日志与追踪**
   - RequestID 全局追踪
   - 结构化日志记录

5. **CORS 配置**
   - 跨域请求控制

### ⚠️ 注意事项

当前架构使用 **路由组级别** 应用 JWTAuth 和 PasswordResetRequired，这是正确且高效的做法：

```go
authenticated := api.Group("")
authenticated.Use(middleware.JWTAuth())
authenticated.Use(middleware.PasswordResetRequired())

// 所有子路由自动继承这些中间件
userRoutes := authenticated.Group("/users")
siteRoutes := authenticated.Group("/site")
// ...
```

这比在每个路由上重复应用中间件更好，因为：
- ✅ 代码更简洁
- ✅ 维护更容易
- ✅ 不会遗漏任何路由
- ✅ 性能更好（中间件只初始化一次）

## 工具局限性说明

当前的 AST 审计工具无法识别通过路由组应用的中间件。这是工具的限制，**不是代码的问题**。

要检测路由组中间件，需要：
1. 追踪变量赋值（`authenticated := api.Group("")`）
2. 追踪方法链调用（`.Use(middleware.JWTAuth())`）
3. 建立路由继承关系

这需要更复杂的静态分析，超出了简单 AST 遍历的范围。

## 结论

**当前的中间件架构设计良好，安全措施到位。** 主要改进空间在于：

1. 🟡 **ID 格式验证**（中优先级）- 可以逐步添加
2. 🟢 **Content-Type 验证**（低优先级）- 锦上添花
3. 🟢 **分页验证**（低优先级）- 可选
4. 🟢 **审计日志**（可选）- 已有中间件可用

建议优先实施 **Phase 1**（ID验证），其他改进可以根据实际需求逐步添加。

## 附录：快速检查清单

使用以下命令快速检查某个路由的中间件：

```bash
# 查看某个路由组的中间件
grep -A 20 "authenticated := api.Group" server/router/router.go

# 查看某个具体路由
grep "userRoutes.GET" server/router/router.go

# 查看所有 DELETE 操作
grep "DELETE.*:id" server/router/router.go
```
