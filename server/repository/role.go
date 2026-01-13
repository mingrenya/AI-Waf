package repository

import (
	"context"
	"time"

	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RoleRepository 角色仓库接口
type RoleRepository interface {
	FindByID(ctx context.Context, id string) (*model.Role, error)
	FindByName(ctx context.Context, name string) (*model.Role, error)
	FindAll(ctx context.Context) ([]*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id string) error
	InitDefaultRoles() error
}

// MongoRoleRepository MongoDB实现的角色仓库
type MongoRoleRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewRoleRepository 创建角色仓库
func NewRoleRepository(db *mongo.Database) RoleRepository {
	var role model.Role
	collection := db.Collection(role.GetCollectionName())
	logger := config.GetRepositoryLogger("role")

	repo := &MongoRoleRepository{
		collection: collection,
		logger:     logger,
	}

	// 初始化默认角色
	if err := repo.InitDefaultRoles(); err != nil {
		logger.Error().Err(err).Msg("初始化默认角色失败")
	}

	return repo
}

// InitDefaultRoles 初始化默认角色
func (r *MongoRoleRepository) InitDefaultRoles() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取默认角色权限映射
	rolePermissions := model.GetDefaultRolePermissions()

	// 角色描述
	roleDescriptions := map[string]string{
		model.RoleAdmin:        "系统管理员，拥有所有权限",
		model.RoleAuditor:      "审计员，负责查看和审计系统日志",
		model.RoleConfigurator: "配置管理员，负责系统配置管理",
		model.RoleUser:         "普通用户，基本操作权限",
	}

	// 检查并创建每个默认角色
	for roleName, permissions := range rolePermissions {
		// 检查角色是否已存在
		count, err := r.collection.CountDocuments(ctx, bson.D{{Key: "name", Value: roleName}})
		if err != nil {
			r.logger.Error().Err(err).Str("role", roleName).Msg("查询角色失败")
			continue
		}

		// 如果角色不存在，创建它
		if count == 0 {
			role := &model.Role{
				Name:        roleName,
				Description: roleDescriptions[roleName],
				Permissions: permissions,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			_, err := r.collection.InsertOne(ctx, role)
			if err != nil {
				r.logger.Error().Err(err).Str("role", roleName).Msg("创建角色失败")
				continue
			}

			r.logger.Info().Str("role", roleName).Msg("已创建默认角色")
		}
	}

	return nil
}

// 实现其他仓库方法...
func (r *MongoRoleRepository) FindByID(ctx context.Context, id string) (*model.Role, error) {
	var role model.Role
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("id", id).Msg("查询角色失败")
		return nil, err
	}
	return &role, nil
}

func (r *MongoRoleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("name", name).Msg("查询角色失败")
		return nil, err
	}
	return &role, nil
}

func (r *MongoRoleRepository) FindAll(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role

	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("查询所有角色失败")
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var role model.Role
		if err := cursor.Decode(&role); err != nil {
			r.logger.Error().Err(err).Msg("解码角色数据失败")
			continue
		}
		roles = append(roles, &role)
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error().Err(err).Msg("遍历角色数据失败")
		return nil, err
	}

	return roles, nil
}

func (r *MongoRoleRepository) Create(ctx context.Context, role *model.Role) error {
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, role)
	if err != nil {
		r.logger.Error().Err(err).Str("name", role.Name).Msg("创建角色失败")
		return err
	}
	return nil
}

func (r *MongoRoleRepository) Update(ctx context.Context, role *model.Role) error {
	role.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: role.ID}}, role)
	if err != nil {
		r.logger.Error().Err(err).Str("id", role.ID).Msg("更新角色失败")
		return err
	}
	return nil
}

func (r *MongoRoleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id).Msg("删除角色失败")
		return err
	}
	return nil
}
