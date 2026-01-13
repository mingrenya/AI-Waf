package dto

import (
	"github.com/HUAHUAI23/RuiQi/server/model"
)

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserPasswordResetRequest 密码重置请求
type UserPasswordResetRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// UserCreateRequest 创建用户请求（仅管理员可用）
type UserCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	Username  string `json:"username,omitempty" binding:"omitempty,min=3,max=20"`
	Password  string `json:"password,omitempty" binding:"omitempty,min=6"`
	Role      string `json:"role,omitempty" binding:"omitempty,oneof=admin auditor configurator user"`
	NeedReset *bool  `json:"needReset,omitempty"`
}

// LoginResponseData 登录响应数据
type LoginResponseData struct {
	Token string     `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."` // JWT token
	User  model.User `json:"user"`                                    // 用户信息
}

type ResetPasswordResponseData = model.TSuccessResponse[model.User]

type GetUserInfoResponseData struct {
	ID        string `json:"id" example:"1234567890"`
	Username  string `json:"username" example:"user123"`
	Role      string `json:"role" example:"admin"`
	NeedReset bool   `json:"needReset" example:"false"`
}
