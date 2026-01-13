package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/HUAHUAI23/RuiQi/server/repository"
	"github.com/HUAHUAI23/RuiQi/server/utils/jwt"
	"github.com/HUAHUAI23/RuiQi/server/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// HasPermission 权限检查中间件
func HasPermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		role, exists := c.Get("userRole")
		if !exists {
			response.Unauthorized(c, fmt.Errorf("请求上下文中没有用户角色信息"))
			c.Abort()
			return
		}

		// 如果是管理员，直接通过
		if role == model.RoleAdmin {
			c.Next()
			return
		}

		// 获取用户权限
		var userPermissions []string

		roleRepo := c.MustGet("roleRepo").(repository.RoleRepository)
		// 获取角色默认权限
		rolePermissions := model.GetDefaultRolePermissions()[role.(string)]

		if rolePermissions == nil {
			ctx := c.Request.Context()
			roleObj, err := roleRepo.FindByName(ctx, role.(string))
			if err != nil {
				response.InternalServerError(c, err, false)
				c.Abort()
				return
			}
			rolePermissions = roleObj.Permissions
		}

		// 获取用户额外权限
		extraPermissions, exists := c.Get("userPermissions")
		if exists {
			userPermissions = append(rolePermissions, extraPermissions.([]string)...)
		} else {
			userPermissions = rolePermissions
		}

		// 检查是否有所需权限
		hasPermission := false
		for _, perm := range userPermissions {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			response.Forbidden(c, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, fmt.Errorf("未提供令牌"))
			c.Abort()
			return
		}

		// 检查格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, fmt.Errorf("无效的令牌格式"))
			c.Abort()
			return
		}

		// 解析令牌
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			if err == jwt.ErrExpiredToken {
				response.Error(c, model.NewAPIError(http.StatusUnauthorized, "令牌已过期", err), false)
			} else {
				response.Unauthorized(c, err)
			}
			c.Abort()
			return
		}

		// 获取用户信息
		userRepo := c.MustGet("userRepo").(repository.UserRepository)
		userID, err := bson.ObjectIDFromHex(claims.UserID)
		if err != nil {
			response.Unauthorized(c, nil)
			c.Abort()
			return
		}
		user, err := userRepo.FindByID(c, userID)
		if err != nil || user == nil {
			response.Unauthorized(c, nil)
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set("userID", user.ID.Hex())
		c.Set("username", user.Username)
		c.Set("userRole", user.Role)
		c.Set("userPermissions", user.Permissions)
		c.Set("needReset", user.NeedReset)

		c.Next()
	}
}

// PasswordResetRequired 密码重置检查中间件
func PasswordResetRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		needReset, exists := c.Get("needReset")
		if !exists {
			response.Unauthorized(c, fmt.Errorf("密码是否重置标志不存在"))
			c.Abort()
			return
		}

		// 如果需要重置密码，只允许访问密码重置接口
		if needReset.(bool) && c.FullPath() != "/api/v1/auth/reset-password" {
			response.Error(c, model.NewAPIError(http.StatusForbidden, "请先重置密码", nil), false)
			c.Abort()
			return
		}

		c.Next()
	}
}
