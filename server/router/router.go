package router

import (
	"github.com/mingrenya/AI-Waf/server/controller"
	"github.com/mingrenya/AI-Waf/server/middleware"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/mingrenya/AI-Waf/server/service"
	alertChecker "github.com/mingrenya/AI-Waf/server/service/cornjob/alert"
	"github.com/mingrenya/AI-Waf/server/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Setup configures all the routes for the application
func Setup(route *gin.Engine, db *mongo.Database) {
	// 基础中间件
	route.Use(middleware.RequestID())
	route.Use(middleware.Logger())
	route.Use(middleware.Cors())
	route.Use(gin.CustomRecovery(middleware.CustomErrorHandler))

	// 创建仓库
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	siteRepo := repository.NewSiteRepository(db)
	wafLogRepo := repository.NewWAFLogRepository(db)
	certRepo := repository.NewCertificateRepository(db)
	configRepo := repository.NewConfigRepository(db)
	ipGroupRepo := repository.NewIPGroupRepository(db)
	ruleRepo := repository.NewMicroRuleRepository(db)
	blockedIPRepo := repository.NewBlockedIPRepository(db)
	alertChannelRepo := repository.NewAlertChannelRepository(db)
	alertRuleRepo := repository.NewAlertRuleRepository(db)
	alertHistoryRepo := repository.NewAlertHistoryRepository(db)
	adaptiveThrottlingRepo := repository.NewAdaptiveThrottlingRepository(db)
	attackPatternRepo := repository.NewAttackPatternRepository(db)
	generatedRuleRepo := repository.NewGeneratedRuleRepository(db)
	aiAnalyzerConfigRepo := repository.NewAIAnalyzerConfigRepository(db)
	mcpConversationRepo := repository.NewMCPConversationRepository(db)
	mcpRepo := repository.NewMCPRepository(db)

	// 创建服务
	authService := service.NewAuthService(userRepo, roleRepo)
	siteService := service.NewSiteService(siteRepo)
	wafLogService := service.NewWAFLogService(wafLogRepo)
	certService := service.NewCertificateService(certRepo)
	runnerService, _ := service.NewRunnerService()
	configService := service.NewConfigService(configRepo)
	ipGroupService := service.NewIPGroupService(ipGroupRepo)
	ruleService := service.NewMicroRuleService(ruleRepo)
	statsService := service.NewStatsService(wafLogRepo)
	blockedIPService := service.NewBlockedIPService(blockedIPRepo)
	alertService := service.NewAlertService(alertChannelRepo, alertRuleRepo, alertHistoryRepo, statsService)
	adaptiveThrottlingService := service.NewAdaptiveThrottlingService(adaptiveThrottlingRepo)
	aiAnalyzerService := service.NewAIAnalyzerService(attackPatternRepo, generatedRuleRepo, aiAnalyzerConfigRepo, mcpConversationRepo)
	mcpService := service.NewMCPService(mcpRepo)
	
	// 启动告警后台任务
	logger := config.GetServiceLogger("router")
	alertCleanup, err := alertChecker.Start(alertService, logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to start alert checker")
	} else {
		logger.Info().Msg("Alert checker started successfully")
		// 注册清理函数到 Gin 的 shutdown hook
		// 注意: 这里我们不能直接 defer，因为 Setup 函数会返回
		// 实际应用中，应该在 main.go 中管理清理函数
		_ = alertCleanup // 保留引用避免未使用变量错误
	}
	
	// 创建控制器
	authController := controller.NewAuthController(authService)
	siteController := controller.NewSiteController(siteService)
	wafLogController := controller.NewWAFLogController(wafLogService)
	certController := controller.NewCertificateController(certService)
	runnerController := controller.NewRunnerController(runnerService)
	configController := controller.NewConfigController(configService)
	ipGroupController := controller.NewIPGroupController(ipGroupService)
	ruleController := controller.NewMicroRuleController(ruleService)
	statsController := controller.NewStatsController(runnerService, statsService)
	blockedIPController := controller.NewBlockedIPController(blockedIPService)
	alertController := controller.NewAlertController(alertService)
	adaptiveThrottlingController := controller.NewAdaptiveThrottlingController(adaptiveThrottlingService)
	aiAnalyzerController := controller.NewAIAnalyzerController(aiAnalyzerService)
	mcpController := controller.NewMCPController(mcpService)
	// 将仓库添加到上下文中，供中间件使用
	route.Use(func(c *gin.Context) {
		c.Set("userRepo", userRepo)
		c.Set("roleRepo", roleRepo)
		c.Next()
	})

	// 健康检查端点
	route.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 路由
	api := route.Group("/api/v1")

	// 认证相关路由 - 不需要权限检查
	auth := api.Group("/auth")
	{
		auth.POST("/login", authController.Login)

		// 需要认证的路由
		authRequired := auth.Group("")
		authRequired.Use(middleware.JWTAuth())
		{
			// 密码重置接口 - 任何已认证用户都可访问
			authRequired.POST("/reset-password", authController.ResetPassword)

			// 需要密码重置检查的路由
			passwordChecked := authRequired.Group("")
			passwordChecked.Use(middleware.PasswordResetRequired())
			{
				// 获取个人信息 - 任何已认证用户都可访问
				passwordChecked.GET("/me", authController.GetUserInfo)
			}
		}
	}

	// 需要认证和密码重置检查的API路由
	authenticated := api.Group("")
	authenticated.Use(middleware.JWTAuth())
	authenticated.Use(middleware.PasswordResetRequired())

	// 用户管理模块
	userRoutes := authenticated.Group("/users")
	{
		// 创建用户 - 需要user:create权限
		userRoutes.POST("", middleware.HasPermission(model.PermUserCreate), authController.CreateUser)
		// 获取用户列表 - 需要user:read权限
		userRoutes.GET("", middleware.HasPermission(model.PermUserRead), authController.GetUsers)
		// 更新用户 - 需要user:update权限
		userRoutes.PUT("/:id", middleware.HasPermission(model.PermUserUpdate), authController.UpdateUser)
		// 删除用户 - 需要user:delete权限
		userRoutes.DELETE("/:id", middleware.HasPermission(model.PermUserDelete), authController.DeleteUser)
	}

	// 站点管理模块
	siteRoutes := authenticated.Group("/site")
	{
		// 创建站点 - 需要site:create权限
		siteRoutes.POST("", middleware.HasPermission(model.PermSiteCreate), siteController.CreateSite)
		// 获取站点列表 - 需要site:read权限
		siteRoutes.GET("", middleware.HasPermission(model.PermSiteRead), siteController.GetSites)
		// 获取单个站点 - 需要site:read权限
		siteRoutes.GET("/:id", middleware.HasPermission(model.PermSiteRead), siteController.GetSiteByID)
		// 更新站点 - 需要site:update权限
		siteRoutes.PUT("/:id", middleware.HasPermission(model.PermSiteUpdate), siteController.UpdateSite)
		// 删除站点 - 需要site:delete权限
		siteRoutes.DELETE("/:id", middleware.HasPermission(model.PermSiteDelete), siteController.DeleteSite)
	}

	// 证书管理路由
	certRoutes := authenticated.Group("/certificate")
	{
		certRoutes.POST("", middleware.HasPermission(model.PermCertCreate), certController.CreateCertificate)
		certRoutes.GET("", middleware.HasPermission(model.PermCertRead), certController.GetCertificates)
		certRoutes.GET("/:id", middleware.HasPermission(model.PermCertRead), certController.GetCertificateByID)
		certRoutes.PUT("/:id", middleware.HasPermission(model.PermCertUpdate), certController.UpdateCertificate)
		certRoutes.DELETE("/:id", middleware.HasPermission(model.PermCertDelete), certController.DeleteCertificate)
	}

	// IP组管理路由
	ipGroupRoutes := authenticated.Group("/ip-groups")
	{
		ipGroupRoutes.POST("", middleware.HasPermission(model.PermConfigUpdate), ipGroupController.CreateIPGroup)
		ipGroupRoutes.GET("", middleware.HasPermission(model.PermConfigRead), ipGroupController.GetIPGroups)
		ipGroupRoutes.GET("/:id", middleware.HasPermission(model.PermConfigRead), ipGroupController.GetIPGroupByID)
		ipGroupRoutes.PUT("/:id", middleware.HasPermission(model.PermConfigUpdate), ipGroupController.UpdateIPGroup)
		ipGroupRoutes.DELETE("/:id", middleware.HasPermission(model.PermConfigUpdate), ipGroupController.DeleteIPGroup)
		// 添加IP到系统默认黑名单
		ipGroupRoutes.POST("/blacklist/add", middleware.HasPermission(model.PermConfigUpdate), ipGroupController.AddIPToBlacklist)
	}

	// rule 管理路由
	ruleRoutes := authenticated.Group("/micro-rules")
	{
		ruleRoutes.POST("", middleware.HasPermission(model.PermConfigUpdate), ruleController.CreateMicroRule)
		ruleRoutes.GET("", middleware.HasPermission(model.PermConfigRead), ruleController.GetMicroRules)
		ruleRoutes.GET("/:id", middleware.HasPermission(model.PermConfigRead), ruleController.GetMicroRuleByID)
		ruleRoutes.PUT("/:id", middleware.HasPermission(model.PermConfigUpdate), ruleController.UpdateMicroRule)
		ruleRoutes.DELETE("/:id", middleware.HasPermission(model.PermConfigUpdate), ruleController.DeleteMicroRule)
	}

	// 日志
	wafLogRoutes := authenticated.Group("/log")
	{
		// 获取攻击事件 - 需要logs:read权限
		wafLogRoutes.GET("/event", middleware.HasPermission(model.PermWAFLogRead), wafLogController.GetAttackEvents)
		// 获取攻击日志 - 需要logs:read权限
		wafLogRoutes.GET("", middleware.HasPermission(model.PermWAFLogRead), wafLogController.GetAttackLogs)
	}

	// 统计信息路由
	statsRoutes := authenticated.Group("/stats")
	{
		// 获取概览统计 - 需要config:read权限
		statsRoutes.GET("/overview", middleware.HasPermission(model.PermWAFLogRead), statsController.GetOverviewStats)
		// 获取实时QPS - 需要config:read权限
		statsRoutes.GET("/realtime-qps", middleware.HasPermission(model.PermWAFLogRead), statsController.GetRealtimeQPS)
		// 获取时间序列数据 - 需要config:read权限
		statsRoutes.GET("/time-series", middleware.HasPermission(model.PermWAFLogRead), statsController.GetTimeSeriesData)
		// 获取组合时间序列数据 - 需要config:read权限
		statsRoutes.GET("/combined-time-series", middleware.HasPermission(model.PermWAFLogRead), statsController.GetCombinedTimeSeriesData)
		// 获取流量时间序列数据 - 需要config:read权限
		statsRoutes.GET("/traffic-time-series", middleware.HasPermission(model.PermWAFLogRead), statsController.GetTrafficTimeSeriesData)
		// 获取综合安全指标 - 需要config:read权限
		statsRoutes.GET("/security-metrics", middleware.HasPermission(model.PermWAFLogRead), statsController.GetSecurityMetrics)
	}

	// 配置管理模块
	runnerRoutes := authenticated.Group("/runner")
	{
		// 获取配置 - 需要config:read权限
		runnerRoutes.GET("/status", middleware.HasPermission(model.PermConfigRead), runnerController.GetStatus)
		// 更新配置 - 需要config:update权限
		runnerRoutes.POST("/control", middleware.HasPermission(model.PermConfigUpdate), runnerController.Control)
	}
	configRoutes := authenticated.Group("/config")
	{
		// 获取配置 - 需要config:read权限
		configRoutes.GET("", middleware.HasPermission(model.PermConfigRead), configController.GetConfig)
		// 更新配置 - 需要config:update权限
		configRoutes.PATCH("", middleware.HasPermission(model.PermConfigUpdate), configController.PatchConfig)
	}

	// 封禁IP管理模块
	blockedIPRoutes := authenticated.Group("/blocked-ips")
	{
		blockedIPRoutes.GET("", middleware.HasPermission(model.PermConfigRead), blockedIPController.GetBlockedIPs)
		blockedIPRoutes.GET("/stats", middleware.HasPermission(model.PermConfigRead), blockedIPController.GetBlockedIPStats)
		blockedIPRoutes.DELETE("/cleanup", middleware.HasPermission(model.PermConfigUpdate), blockedIPController.CleanupExpiredBlockedIPs)
	}

	// 自适应限流模块
	adaptiveThrottlingRoutes := authenticated.Group("/adaptive-throttling")
	{
		// 配置管理
		adaptiveThrottlingRoutes.GET("", middleware.HasPermission(model.PermConfigRead), adaptiveThrottlingController.GetConfig)
		adaptiveThrottlingRoutes.PUT("", middleware.HasPermission(model.PermConfigUpdate), adaptiveThrottlingController.UpdateConfig)
		adaptiveThrottlingRoutes.DELETE("", middleware.HasPermission(model.PermConfigUpdate), adaptiveThrottlingController.DeleteConfig)
		
		// 数据查询
		adaptiveThrottlingRoutes.GET("/patterns", middleware.HasPermission(model.PermConfigRead), adaptiveThrottlingController.GetTrafficPatterns)
		adaptiveThrottlingRoutes.GET("/baselines", middleware.HasPermission(model.PermConfigRead), adaptiveThrottlingController.GetBaselines)
		adaptiveThrottlingRoutes.GET("/logs", middleware.HasPermission(model.PermConfigRead), adaptiveThrottlingController.GetAdjustmentLogs)
		adaptiveThrottlingRoutes.GET("/stats", middleware.HasPermission(model.PermConfigRead), adaptiveThrottlingController.GetStats)
		
		// 操作
		adaptiveThrottlingRoutes.POST("/recalculate-baseline", middleware.HasPermission(model.PermConfigUpdate), adaptiveThrottlingController.RecalculateBaseline)
		adaptiveThrottlingRoutes.POST("/reset-learning", middleware.HasPermission(model.PermConfigUpdate), adaptiveThrottlingController.ResetLearning)
	}

	// 告警管理模块
	alertRoutes := authenticated.Group("/alerts")
	{
		// 告警渠道管理
		alertRoutes.POST("/channels", middleware.HasPermission(model.PermAlertChannelCreate), alertController.CreateChannel)
		alertRoutes.GET("/channels", middleware.HasPermission(model.PermAlertChannelRead), alertController.GetChannels)
		alertRoutes.GET("/channels/:id", middleware.HasPermission(model.PermAlertChannelRead), alertController.GetChannelByID)
		alertRoutes.PUT("/channels/:id", middleware.HasPermission(model.PermAlertChannelUpdate), alertController.UpdateChannel)
		alertRoutes.DELETE("/channels/:id", middleware.HasPermission(model.PermAlertChannelDelete), alertController.DeleteChannel)
		alertRoutes.POST("/channels/:id/test", middleware.HasPermission(model.PermAlertChannelUpdate), alertController.TestChannel)

		// 告警规则管理
		alertRoutes.POST("/rules", middleware.HasPermission(model.PermAlertRuleCreate), alertController.CreateRule)
		alertRoutes.GET("/rules", middleware.HasPermission(model.PermAlertRuleRead), alertController.GetRules)
		alertRoutes.GET("/rules/:id", middleware.HasPermission(model.PermAlertRuleRead), alertController.GetRuleByID)
		alertRoutes.PUT("/rules/:id", middleware.HasPermission(model.PermAlertRuleUpdate), alertController.UpdateRule)
		alertRoutes.DELETE("/rules/:id", middleware.HasPermission(model.PermAlertRuleDelete), alertController.DeleteRule)

		// 告警历史查询
		alertRoutes.GET("/history", middleware.HasPermission(model.PermAlertHistoryRead), alertController.GetAlertHistory)
		alertRoutes.POST("/history/:id/acknowledge", middleware.HasPermission(model.PermAlertHistoryRead), alertController.AcknowledgeAlert)
		
		// 告警统计
		alertRoutes.GET("/statistics", middleware.HasPermission(model.PermAlertHistoryRead), alertController.GetStatistics)
	}

	// AI分析器模块
	aiAnalyzerRoutes := authenticated.Group("/ai-analyzer")
	{
		// 攻击模式管理
		aiAnalyzerRoutes.GET("/patterns", middleware.HasPermission(model.PermWAFLogRead), aiAnalyzerController.ListAttackPatterns)
		aiAnalyzerRoutes.GET("/patterns/:id", middleware.HasPermission(model.PermWAFLogRead), aiAnalyzerController.GetAttackPattern)
		aiAnalyzerRoutes.DELETE("/patterns/:id", middleware.HasPermission(model.PermConfigUpdate), aiAnalyzerController.DeleteAttackPattern)

		// 生成规则管理
		aiAnalyzerRoutes.GET("/rules", middleware.HasPermission(model.PermConfigRead), aiAnalyzerController.ListGeneratedRules)
		aiAnalyzerRoutes.GET("/rules/:id", middleware.HasPermission(model.PermConfigRead), aiAnalyzerController.GetGeneratedRule)
		aiAnalyzerRoutes.DELETE("/rules/:id", middleware.HasPermission(model.PermConfigUpdate), aiAnalyzerController.DeleteGeneratedRule)
		aiAnalyzerRoutes.POST("/rules/review", middleware.HasPermission(model.PermConfigUpdate), aiAnalyzerController.ReviewRule)
		aiAnalyzerRoutes.GET("/rules/pending", middleware.HasPermission(model.PermConfigRead), aiAnalyzerController.GetPendingRules)
		aiAnalyzerRoutes.POST("/rules/:id/deploy", middleware.HasPermission(model.PermConfigUpdate), aiAnalyzerController.DeployRule)

		// AI分析器配置
		aiAnalyzerRoutes.GET("/config", middleware.HasPermission(model.PermConfigRead), aiAnalyzerController.GetAnalyzerConfig)
		aiAnalyzerRoutes.PUT("/config", middleware.HasPermission(model.PermConfigUpdate), aiAnalyzerController.UpdateAnalyzerConfig)

		// MCP对话管理
		aiAnalyzerRoutes.GET("/conversations", middleware.HasPermission(model.PermWAFLogRead), aiAnalyzerController.ListMCPConversations)
		aiAnalyzerRoutes.GET("/conversations/:id", middleware.HasPermission(model.PermWAFLogRead), aiAnalyzerController.GetMCPConversation)
		aiAnalyzerRoutes.DELETE("/conversations/:id", middleware.HasPermission(model.PermConfigUpdate), aiAnalyzerController.DeleteMCPConversation)

		// 统计分析
		aiAnalyzerRoutes.GET("/stats", middleware.HasPermission(model.PermWAFLogRead), aiAnalyzerController.GetAnalyzerStats)
		
		// 手动触发AI分析
		aiAnalyzerRoutes.POST("/trigger", middleware.HasPermission(model.PermConfigUpdate), aiAnalyzerController.TriggerAnalysis)
	}

	// MCP 服务模块
	mcpRoutes := authenticated.Group("/mcp")
	{
		// 获取MCP连接状态 - 任何已认证用户都可访问
		mcpRoutes.GET("/status", mcpController.GetMCPStatus)
		// 获取MCP工具列表 - 任何已认证用户都可访问
		mcpRoutes.GET("/tools", mcpController.GetMCPTools)
		// 获取工具调用历史 - 需要logs:read权限
		mcpRoutes.GET("/tool-calls", middleware.HasPermission(model.PermWAFLogRead), mcpController.GetMCPToolCallHistory)
		// 记录工具调用 - MCP Server调用，需要认证但不需要特殊权限
		mcpRoutes.POST("/tool-calls/record", mcpController.RecordToolCall)
	}

	// 审计日志模块
	auditRoutes := authenticated.Group("/audit")
	{
		// 获取审计日志 - 需要audit:read权限
		auditRoutes.GET("", middleware.HasPermission(model.PermAuditRead), nil)
	}

	// 系统管理模块
	systemRoutes := authenticated.Group("/system")
	{
		// 获取系统状态 - 需要system:status权限
		systemRoutes.GET("/status", middleware.HasPermission(model.PermSystemStatus), nil)
		// 重启系统 - 需要system:restart权限
		systemRoutes.POST("/restart", middleware.HasPermission(model.PermSystemRestart), nil)
	}

	// ===== 前端静态资源托管 =====
	SetStaticFileRouter(route)
}
