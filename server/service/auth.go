package service

import (
	"context"
	"errors"
	"time"

	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/dto"
	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/HUAHUAI23/RuiQi/server/repository"
	"github.com/HUAHUAI23/RuiQi/server/utils/jwt"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// 定义错误
var (
	ErrUserNotFound     = errors.New("用户不存在")
	ErrInvalidPassword  = errors.New("密码错误")
	ErrUserAlreadyExist = errors.New("用户已存在")
	ErrForbidden        = errors.New("没有权限")
)

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, req dto.UserLoginRequest) (string, *model.User, error)
	ResetPassword(ctx context.Context, userID bson.ObjectID, req dto.UserPasswordResetRequest) error
	CreateUser(ctx context.Context, adminID bson.ObjectID, req dto.UserCreateRequest) (*model.User, error)
	GetUsers(ctx context.Context) ([]*model.User, error)
}

// AuthServiceImpl 认证服务实现
type AuthServiceImpl struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	logger   zerolog.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) AuthService {
	return &AuthServiceImpl{
		userRepo: userRepo,
		roleRepo: roleRepo,
		logger:   config.GetServiceLogger("auth"),
	}
}

// Login 用户登录
func (s *AuthServiceImpl) Login(ctx context.Context, req dto.UserLoginRequest) (string, *model.User, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, ErrUserNotFound
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return "", nil, ErrInvalidPassword
	}

	// 更新最后登录时间
	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		s.logger.Warn().Err(err).Str("userId", user.ID.Hex()).Msg("更新登录时间失败")
	}

	// 生成令牌
	token, err := jwt.GenerateToken(*user, 24*time.Hour)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// ResetPassword 重置密码
func (s *AuthServiceImpl) ResetPassword(ctx context.Context, userID bson.ObjectID, req dto.UserPasswordResetRequest) error {
	// 查找用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 验证旧密码
	if !user.CheckPassword(req.OldPassword) {
		return ErrInvalidPassword
	}

	// 更新密码
	user.Password = req.NewPassword
	if err := user.HashPassword(); err != nil {
		return err
	}

	// 标记密码已重置
	user.NeedReset = false
	user.UpdatedAt = time.Now()

	// 保存用户
	return s.userRepo.Update(ctx, user)
}

// CreateUser 创建用户（仅管理员可用）
func (s *AuthServiceImpl) CreateUser(ctx context.Context, adminID bson.ObjectID, req dto.UserCreateRequest) (*model.User, error) {
	// 验证管理员权限
	admin, err := s.userRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if admin == nil || admin.Role != model.RoleAdmin {
		return nil, ErrForbidden
	}

	// 检查用户是否已存在
	existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExist
	}

	// 创建新用户
	user := &model.User{
		Username:  req.Username,
		Password:  req.Password,
		Role:      req.Role,
		NeedReset: true, // 新用户需要重置密码
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 哈希密码
	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	// 保存用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 不返回密码
	user.Password = ""
	return user, nil
}

// GetUsers 获取所有用户（仅管理员可用）
func (s *AuthServiceImpl) GetUsers(ctx context.Context) ([]*model.User, error) {
	// 实现获取所有用户的逻辑
	// 这需要在 UserRepository 中添加 FindAll 方法
	return s.userRepo.FindAll(ctx)
}
