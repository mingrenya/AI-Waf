package model

import "time"

// 角色常量
const (
	RoleAdmin        = "admin"        // 管理员，拥有所有权限
	RoleAuditor      = "auditor"      // 审计员，可查看日志和审计信息
	RoleConfigurator = "configurator" // 配置管理员，可管理系统配置
	RoleUser         = "user"         // 普通用户，基本操作权限
)

// 权限常量
const (
	// 用户管理权限
	PermUserCreate = "user:create"
	PermUserRead   = "user:read"
	PermUserUpdate = "user:update"
	PermUserDelete = "user:delete"

	// 站点管理权限
	PermSiteCreate = "site:create"
	PermSiteRead   = "site:read"
	PermSiteUpdate = "site:update"
	PermSiteDelete = "site:delete"

	// 证书管理权限
	PermCertCreate = "cert:create"
	PermCertRead   = "cert:read"
	PermCertUpdate = "cert:update"
	PermCertDelete = "cert:delete"

	// 配置管理权限
	PermConfigRead   = "config:read"
	PermConfigUpdate = "config:update"

	// 审计日志权限
	PermAuditRead = "audit:read"

	// 系统管理权限
	PermSystemRestart = "system:restart"
	PermSystemStatus  = "system:status"

	// WAF日志权限
	PermWAFLogRead = "waf:log:read"
)

// Role 角色模型
type Role struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string    `bson:"name" json:"name"`               // 角色名称
	Description string    `bson:"description" json:"description"` // 角色描述
	Permissions []string  `bson:"permissions" json:"permissions"` // 权限列表
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

// 获取默认角色权限映射
func GetDefaultRolePermissions() map[string][]string {
	return map[string][]string{
		RoleAdmin: {
			// 管理员拥有所有权限
			PermUserCreate, PermUserRead, PermUserUpdate, PermUserDelete,
			PermSiteCreate, PermSiteRead, PermSiteUpdate, PermSiteDelete,
			PermConfigRead, PermConfigUpdate,
			PermAuditRead,
			PermSystemRestart, PermSystemStatus,
			PermWAFLogRead,
			PermCertCreate, PermCertRead, PermCertUpdate, PermCertDelete,
		},
		RoleAuditor: {
			// 审计员可以查看用户、站点、配置和审计日志
			PermUserRead,
			PermSiteRead,
			PermConfigRead,
			PermAuditRead,
			PermSystemStatus,
			PermWAFLogRead,
			PermCertRead,
		},
		RoleConfigurator: {
			// 配置管理员可以管理站点和配置
			PermSiteCreate, PermSiteRead, PermSiteUpdate, PermSiteDelete,
			PermConfigRead, PermConfigUpdate,
			PermSystemStatus,
			PermWAFLogRead,
			PermCertRead, PermCertUpdate, PermCertDelete,
		},
		RoleUser: {
			// 普通用户只能查看站点和系统状态
			PermSiteRead,
			PermSystemStatus,
			PermWAFLogRead,
			PermCertRead,
		},
	}
}

// GetCollectionName 返回集合名称
func (r *Role) GetCollectionName() string {
	return "role"
}
