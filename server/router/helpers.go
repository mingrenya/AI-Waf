package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// IDRoute 为带ID参数的路由创建标准中间件链
// 包括: ID验证 + 权限检查 + 处理函数
func IDRoute(idParam string, permission string, handler gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.ValidateMongoID(idParam),
		middleware.HasPermission(permission),
		handler,
	}
}

// CreateRoute 为创建接口创建标准中间件链
// 包括: Content-Type验证 + 权限检查 + 处理函数
func CreateRoute(permission string, handler gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.ValidateJSONContentType(),
		middleware.HasPermission(permission),
		handler,
	}
}

// UpdateRoute 为更新接口创建标准中间件链
// 包括: ID验证 + Content-Type验证 + 权限检查 + 处理函数
func UpdateRoute(idParam string, permission string, handler gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.ValidateMongoID(idParam),
		middleware.ValidateJSONContentType(),
		middleware.HasPermission(permission),
		handler,
	}
}

// DeleteRoute 为删除接口创建标准中间件链
// 包括: ID验证 + 审计日志 + 权限检查 + 处理函数
func DeleteRoute(db *mongo.Database, idParam string, permission string, handler gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.ValidateMongoID(idParam),
		middleware.SecurityAudit(db),
		middleware.HasPermission(permission),
		handler,
	}
}

// ListRoute 为列表查询接口创建标准中间件链
// 包括: 分页验证 + 权限检查 + 处理函数
func ListRoute(permission string, handler gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.ValidatePagination(),
		middleware.HasPermission(permission),
		handler,
	}
}

// SensitiveRoute 为敏感操作创建标准中间件链
// 包括: 审计日志 + 权限检查 + 处理函数
func SensitiveRoute(db *mongo.Database, permission string, handler gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.SecurityAudit(db),
		middleware.HasPermission(permission),
		handler,
	}
}
