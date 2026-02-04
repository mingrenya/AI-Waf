# Gin 中间件集合

本项目包含了一系列生产级别的 Gin 中间件，用于增强 API 的安全性、可观测性和可靠性。

## 中间件列表

### 1. 安全审计中间件 (`audit.go`)

**功能**：记录所有敏感操作到 MongoDB，包括用户信息、操作详情、耗时等。

**使用场景**：
- 记录所有 POST/PUT/PATCH/DELETE 操作
- 用于安全审计和合规要求
- 追踪敏感数据变更

**使用方法**：
```go
// 在路由组中使用
api := router.Group("/api")
api.Use(middleware.SecurityAudit())
{
    api.POST("/users", userController.Create)
    api.PUT("/users/:id", userController.Update)
}
```

**审计日志查询**：
```go
// 查询特定用户的操作日志
filter := bson.M{"user_id": userId}
result, err := middleware.QueryAuditLogs(filter, 0, 100)

// 清理旧日志（保留30天）
err := middleware.CleanupOldAuditLogs(30)
```

### 2. 请求追踪中间件 (`request_id.go`)

**功能**：为每个请求生成唯一ID，支持分布式链路追踪。

**使用场景**：
- 请求全链路追踪
- 分布式系统调用追踪
- 日志关联和问题排查

**使用方法**：
```go
// 全局使用
router.Use(middleware.RequestID())
router.Use(middleware.TraceID())

// 在处理函数中获取
func handler(c *gin.Context) {
    requestID := c.GetString("requestID")
    traceID := c.GetString("traceID")
    spanID := c.GetString("spanID")
}
```

### 3. 错误恢复中间件 (`recovery.go`)

**功能**：捕获 panic 并记录详细的错误信息和堆栈。

**使用场景**：
- 防止应用崩溃
- 记录详细的错误堆栈
- 返回友好的错误响应

**使用方法**：
```go
// 应该在所有中间件之前使用
router.Use(middleware.Recovery())
router.Use(middleware.ErrorHandler())
```

### 4. 参数验证中间件 (`validator.go`)

**功能**：提供多种参数验证功能。

**使用场景**：
- 验证路径参数格式
- 验证查询参数
- 验证分页参数
- 设置安全响应头

**使用方法**：
```go
// 验证 MongoDB ObjectID
router.GET("/api/users/:id", 
    middleware.ValidateMongoID("id"),
    userController.GetByID)

// 验证 UUID
router.GET("/api/sessions/:uuid",
    middleware.ValidateUUID("uuid"),
    sessionController.Get)

// 验证分页参数
router.GET("/api/users",
    middleware.ValidatePagination(),
    userController.List)

// 验证 Content-Type
router.POST("/api/users",
    middleware.ValidateJSONContentType(),
    userController.Create)

// 设置安全响应头
router.Use(middleware.SecurityHeaders())

// 禁用缓存（敏感接口）
router.GET("/api/profile",
    middleware.NoCache(),
    profileController.Get)
```

### 5. 超时控制中间件 (`timeout.go`)

**功能**：为请求设置超时时间，防止长时间占用资源。

**使用场景**：
- 防止慢查询阻塞
- 控制API响应时间
- 优雅处理超时

**使用方法**：
```go
// 设置 5 秒超时
router.Use(middleware.Timeout(5 * time.Second))

// 或针对特定路由
router.GET("/api/slow-operation",
    middleware.Timeout(30 * time.Second),
    controller.SlowOperation)

// 使用自定义超时处理器
router.Use(middleware.TimeoutWithCustomHandler(
    5 * time.Second,
    func(c *gin.Context) {
        c.JSON(http.StatusGatewayTimeout, gin.H{
            "error": "请求超时，请稍后重试",
        })
    },
))
```

### 6. 限流中间件 (`rate_limit.go`)

**功能**：基于 IP 的请求限流。

**使用场景**：
- 防止API滥用
- DDoS防护
- 保护后端资源

**使用方法**：
```go
// 每分钟最多100次请求
router.Use(middleware.RateLimit(100, time.Minute))

// 针对登录接口限流
router.POST("/api/auth/login",
    middleware.RateLimit(5, time.Minute),
    authController.Login)
```

### 7. JWT认证中间件 (`auth.go`)

**功能**：JWT令牌验证和权限检查。

**使用场景**：
- 用户身份验证
- 基于角色的访问控制
- API权限管理

**使用方法**：
```go
// 需要认证的路由
api := router.Group("/api")
api.Use(middleware.JWTAuth())
{
    // 需要特定权限
    api.POST("/sites",
        middleware.HasPermission(model.PermSiteCreate),
        siteController.Create)
    
    api.DELETE("/sites/:id",
        middleware.HasPermission(model.PermSiteDelete),
        siteController.Delete)
}
```

### 8. CORS中间件 (`middleware.go`)

**功能**：处理跨域请求。

**使用场景**：
- 前后端分离
- 跨域API调用

**使用方法**：
```go
// 通过环境变量配置允许的域名
// CORS_ALLOWED_ORIGINS=http://localhost:3000,https://example.com
router.Use(middleware.Cors())
```

### 9. 日志中间件 (`middleware.go`)

**功能**：记录HTTP请求日志。

**使用场景**：
- 访问日志记录
- 性能监控
- 问题排查

**使用方法**：
```go
router.Use(middleware.Logger())
```

## 完整使用示例

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/mingrenya/AI-Waf/server/middleware"
)

func setupRouter() *gin.Engine {
    router := gin.New()

    // 1. 错误恢复（最先）
    router.Use(middleware.Recovery())
    
    // 2. 请求追踪
    router.Use(middleware.RequestID())
    router.Use(middleware.TraceID())
    
    // 3. 日志记录
    router.Use(middleware.Logger())
    
    // 4. 安全响应头
    router.Use(middleware.SecurityHeaders())
    
    // 5. CORS
    router.Use(middleware.Cors())
    
    // 6. 全局限流
    router.Use(middleware.RateLimit(1000, time.Minute))
    
    // 7. 全局超时
    router.Use(middleware.Timeout(30 * time.Second))

    // 公开路由
    public := router.Group("/api/public")
    {
        public.POST("/login", loginHandler)
    }

    // 需要认证的路由
    api := router.Group("/api")
    api.Use(middleware.JWTAuth())
    api.Use(middleware.SecurityAudit()) // 记录所有操作
    {
        // 用户管理（需要权限）
        users := api.Group("/users")
        {
            users.GET("", 
                middleware.ValidatePagination(),
                listUsers)
            
            users.GET("/:id",
                middleware.ValidateMongoID("id"),
                getUser)
            
            users.POST("",
                middleware.ValidateJSONContentType(),
                middleware.HasPermission("user:create"),
                createUser)
            
            users.PUT("/:id",
                middleware.ValidateMongoID("id"),
                middleware.ValidateJSONContentType(),
                middleware.HasPermission("user:update"),
                updateUser)
            
            users.DELETE("/:id",
                middleware.ValidateMongoID("id"),
                middleware.HasPermission("user:delete"),
                deleteUser)
        }
        
        // 敏感操作（禁用缓存）
        api.GET("/profile",
            middleware.NoCache(),
            getProfile)
    }

    return router
}
```

## 中间件执行顺序

中间件的执行顺序很重要，建议按以下顺序使用：

1. **Recovery** - 必须最先，捕获所有panic
2. **RequestID/TraceID** - 生成追踪ID
3. **Logger** - 记录请求日志
4. **SecurityHeaders** - 设置安全响应头
5. **Cors** - 处理跨域
6. **RateLimit** - 限流控制
7. **Timeout** - 超时控制
8. **JWTAuth** - 身份验证
9. **SecurityAudit** - 安全审计
10. **HasPermission** - 权限检查
11. **业务验证中间件** - ValidateXXX系列

## 性能考虑

1. **SecurityAudit** 使用异步写入，不阻塞请求
2. **RateLimit** 使用内存存储，定期清理
3. **Logger** 在生产环境只记录错误和慢请求
4. **Recovery** 捕获详细堆栈但不影响性能

## 配置建议

### 开发环境
```go
router.Use(middleware.Recovery())
router.Use(middleware.RequestID())
router.Use(middleware.Logger()) // 记录所有请求
router.Use(middleware.Cors())
```

### 生产环境
```go
router.Use(middleware.Recovery())
router.Use(middleware.RequestID())
router.Use(middleware.Logger()) // 只记录错误和慢请求
router.Use(middleware.SecurityHeaders())
router.Use(middleware.Cors())
router.Use(middleware.RateLimit(1000, time.Minute))
router.Use(middleware.Timeout(30 * time.Second))
router.Use(middleware.SecurityAudit()) // 记录敏感操作
```

## 监控和维护

### 审计日志维护
```go
// 定时任务：每天清理30天前的审计日志
func setupAuditLogCleanup() {
    ticker := time.NewTicker(24 * time.Hour)
    go func() {
        for range ticker.C {
            if err := middleware.CleanupOldAuditLogs(30); err != nil {
                log.Error().Err(err).Msg("Failed to cleanup audit logs")
            }
        }
    }()
}
```

### 查询审计日志
```go
// 查询特定时间范围的操作
filter := bson.M{
    "timestamp": bson.M{
        "$gte": startTime,
        "$lte": endTime,
    },
    "action": "DELETE",
}
result, err := middleware.QueryAuditLogs(filter, 0, 100)
```

## 扩展建议

可以根据业务需求添加更多中间件：

1. **API版本控制中间件**
2. **IP白名单/黑名单中间件**
3. **请求签名验证中间件**
4. **敏感数据脱敏中间件**
5. **API使用统计中间件**
6. **熔断器中间件**
