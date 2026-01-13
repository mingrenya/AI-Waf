package repository

import (
	"context"
	"time"

	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	UpdateLastLogin(ctx context.Context, id bson.ObjectID) error
	FindAll(ctx context.Context) ([]*model.User, error)
	InitAdminUser() error
}

// MongoUserRepository MongoDB实现的用户仓库
type MongoUserRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *mongo.Database) UserRepository {
	var user model.User
	collection := db.Collection(user.GetCollectionName())
	logger := config.GetRepositoryLogger("user")

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 用户名唯一索引
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建用户名索引失败")
	}

	repo := &MongoUserRepository{
		collection: collection,
		logger:     logger,
	}

	// 初始化管理员用户
	if err := repo.InitAdminUser(); err != nil {
		logger.Error().Err(err).Msg("初始化管理员用户失败")
	}

	return repo
}

// FindByID 根据ID查找用户
func (r *MongoUserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("查询用户失败")
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *MongoUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.D{{Key: "username", Value: username}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("username", username).Msg("查询用户失败")
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *MongoUserRepository) Create(ctx context.Context, user *model.User) error {
	// 设置时间
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// 不设置ID，让MongoDB自动生成ObjectID
	// MongoDB会在插入时自动创建_id字段

	// 插入数据
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		r.logger.Error().Err(err).Str("username", user.Username).Msg("创建用户失败")
		return err
	}

	// 将生成的ObjectID赋值给用户ID
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		user.ID = oid
	}

	return nil
}

// Update 更新用户
func (r *MongoUserRepository) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: user.ID}}, user)
	if err != nil {
		r.logger.Error().Err(err).Str("id", user.ID.Hex()).Msg("更新用户失败")
		return err
	}
	return nil
}

// UpdateLastLogin 更新最后登录时间
func (r *MongoUserRepository) UpdateLastLogin(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "lastLogin", Value: time.Now()},
			{Key: "updatedAt", Value: time.Now()},
		}}},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("更新登录时间失败")
		return err
	}
	return nil
}

// FindAll 查找所有用户
func (r *MongoUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("查询所有用户失败")
		return nil, err
	}

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error().Err(err).Msg("解析用户列表失败")
		return nil, err
	}

	return users, nil
}

// InitAdminUser 初始化管理员用户
func (r *MongoUserRepository) InitAdminUser() error {
	// 检查是否已有管理员用户
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.D{{Key: "role", Value: model.RoleAdmin}})
	if err != nil {
		r.logger.Error().Err(err).Msg("查询管理员用户失败")
		return err
	}

	// 如果没有管理员用户，创建一个
	if count == 0 {
		// 创建默认管理员用户
		adminUser := &model.User{
			Username:  model.RoleAdmin,
			Password:  "admin123", // 初始密码
			Role:      model.RoleAdmin,
			NeedReset: true, // 需要重置密码
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 哈希密码
		if err := adminUser.HashPassword(); err != nil {
			r.logger.Error().Err(err).Msg("哈希管理员密码失败")
			return err
		}

		// 保存用户
		if err := r.Create(ctx, adminUser); err != nil {
			r.logger.Error().Err(err).Msg("创建管理员用户失败")
			return err
		}

		// 创建审计员用户
		auditorUser := &model.User{
			Username:  model.RoleAuditor,
			Password:  "auditor123", // 初始密码
			Role:      model.RoleAuditor,
			NeedReset: true, // 需要重置密码
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 哈希密码
		if err := auditorUser.HashPassword(); err != nil {
			r.logger.Error().Err(err).Msg("哈希审计员密码失败")
			return err
		}

		// 保存用户
		if err := r.Create(ctx, auditorUser); err != nil {
			r.logger.Error().Err(err).Msg("创建审计员用户失败")
			return err
		}

		// 创建配置管理员用户
		configUser := &model.User{
			Username:  model.RoleConfigurator,
			Password:  "config123", // 初始密码
			Role:      model.RoleConfigurator,
			NeedReset: true, // 需要重置密码
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 哈希密码
		if err := configUser.HashPassword(); err != nil {
			r.logger.Error().Err(err).Msg("哈希配置管理员密码失败")
			return err
		}

		// 保存用户
		if err := r.Create(ctx, configUser); err != nil {
			r.logger.Error().Err(err).Msg("创建配置管理员用户失败")
			return err
		}

		r.logger.Info().Msg("已创建默认用户")
	}

	return nil
}
