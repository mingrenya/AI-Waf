// server/repository/config.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/HUAHUAI23/RuiQi/pkg/model"
	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/constant"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrConfigNotFound = errors.New("配置不存在")
)

// ConfigRepository 配置仓库接口
type ConfigRepository interface {
	GetConfig(ctx context.Context) (*model.Config, error)
	UpdateConfig(ctx context.Context, config *model.Config) error
}

// MongoConfigRepository MongoDB实现的配置仓库
type MongoConfigRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewConfigRepository 创建配置仓库
func NewConfigRepository(db *mongo.Database) ConfigRepository {
	var cfg model.Config
	collection := db.Collection(cfg.GetCollectionName())
	logger := config.GetRepositoryLogger("config")

	return &MongoConfigRepository{
		collection: collection,
		logger:     logger,
	}
}

// GetConfig 获取配置
func (r *MongoConfigRepository) GetConfig(ctx context.Context) (*model.Config, error) {
	var cfg model.Config

	// 使用常量获取配置名称，如果不存在则使用默认值"AppConfig"
	configName := constant.GetString("APP_CONFIG_NAME", "AppConfig")

	// 使用名称查询配置
	err := r.collection.FindOne(
		ctx,
		bson.D{{Key: "name", Value: configName}},
	).Decode(&cfg)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrConfigNotFound
		}
		r.logger.Error().Err(err).Msg("查询配置时出错")
		return nil, err
	}

	return &cfg, nil
}

// UpdateConfig 更新配置
func (r *MongoConfigRepository) UpdateConfig(ctx context.Context, config *model.Config) error {
	// 更新时间
	config.UpdatedAt = time.Now()

	// 使用名称查找并更新配置
	filter := bson.D{{Key: "name", Value: config.Name}}
	update := bson.D{{Key: "$set", Value: config}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Msg("更新配置时出错")
		return err
	}

	if result.MatchedCount == 0 {
		return ErrConfigNotFound
	}

	return nil
}
