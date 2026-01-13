package repository

import (
	"context"
	"errors"
	"time"

	"github.com/HUAHUAI23/RuiQi/server/config"

	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrSiteNotFound       = errors.New("站点不存在")
	ErrDomainPortExists   = errors.New("域名和端口组合已存在")
	ErrDomainPortConflict = errors.New("域名和端口组合已被其他站点使用")
)

// SiteRepository 站点仓库
type SiteRepository interface {
	CreateSite(ctx context.Context, site *model.Site) error
	GetSites(ctx context.Context, page, size int64) ([]model.Site, int64, error)
	GetSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error)
	UpdateSite(ctx context.Context, site *model.Site) error
	DeleteSite(ctx context.Context, id bson.ObjectID) error
	CheckDomainPortExists(ctx context.Context, site *model.Site) error
	CheckDomainPortConflict(ctx context.Context, site *model.Site) error
}

// SiteRepository 站点仓库
type MongoSiteRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewSiteRepository 创建站点仓库
func NewSiteRepository(db *mongo.Database) SiteRepository {
	var site model.Site
	collection := db.Collection(site.GetCollectionName())
	logger := config.GetRepositoryLogger("site")

	return &MongoSiteRepository{
		collection: collection,
		logger:     logger,
	}
}

// CreateSite 创建站点
func (r *MongoSiteRepository) CreateSite(ctx context.Context, site *model.Site) error {
	// 设置创建和更新时间
	now := time.Now()
	site.CreatedAt = now
	site.UpdatedAt = now

	// 插入新站点
	result, err := r.collection.InsertOne(ctx, site)
	if err != nil {
		r.logger.Error().Err(err).Msg("插入站点时出错")
		return err
	}

	// 设置ID
	if id, ok := result.InsertedID.(bson.ObjectID); ok {
		site.ID = id
	}

	return nil
}

// GetSites 获取站点列表
func (r *MongoSiteRepository) GetSites(ctx context.Context, page, size int64) ([]model.Site, int64, error) {
	// 计算分页
	skip := (page - 1) * size

	// 设置查询选项
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "createdAt", Value: -1}}) // 按创建时间降序排序

	// 执行查询
	cursor, err := r.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("查询站点列表时出错")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var sites []model.Site
	if err = cursor.All(ctx, &sites); err != nil {
		r.logger.Error().Err(err).Msg("解析站点列表时出错")
		return nil, 0, err
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("获取站点总数时出错")
		return nil, 0, err
	}

	return sites, total, nil
}

// GetSiteByID 根据ID获取站点
func (r *MongoSiteRepository) GetSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error) {
	var site model.Site
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&site)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrSiteNotFound // 返回找不到错误
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("查询站点时出错")
		return nil, err
	}

	return &site, nil
}

// UpdateSite 更新站点
func (r *MongoSiteRepository) UpdateSite(ctx context.Context, site *model.Site) error {
	// 更新站点
	site.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: site.ID}},
		site,
	)

	if err != nil {
		r.logger.Error().Err(err).Str("id", site.ID.Hex()).Msg("更新站点时出错")
		return err
	}

	return nil
}

// DeleteSite 删除站点
func (r *MongoSiteRepository) DeleteSite(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除站点时出错")
		return err
	}

	return nil
}

func (r *MongoSiteRepository) CheckDomainPortExists(ctx context.Context, site *model.Site) error {
	// 检查域名和端口组合是否已存在
	filter := bson.D{
		{Key: "domain", Value: site.Domain},
		{Key: "listenPort", Value: site.ListenPort},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("检查站点域名和端口是否存在时出错")
		return err
	}
	if count > 0 {
		r.logger.Error().Msg("站点域名和端口组合已存在")
		return ErrDomainPortExists
	}
	return nil
}

func (r *MongoSiteRepository) CheckDomainPortConflict(ctx context.Context, site *model.Site) error {
	// 检查域名和端口组合是否与其他站点冲突
	filter := bson.D{
		{Key: "_id", Value: bson.D{{Key: "$ne", Value: site.ID}}},
		{Key: "domain", Value: site.Domain},
		{Key: "listenPort", Value: site.ListenPort},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("检查站点域名和端口冲突时出错")
		return err
	}
	if count > 0 {
		r.logger.Error().Msg("更新站点失败，站点域名和端口组合已存在")
		return ErrDomainPortConflict
	}
	return nil
}

// GetAllSites 获取所有站点，不分页
func GetAllSites(ctx context.Context, collection *mongo.Collection) ([]model.Site, error) {
	// 设置查询选项，按创建时间降序排序
	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	// 执行查询
	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		config.Logger.Error().Err(err).Msg("查询所有站点时出错")
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var sites []model.Site
	if err = cursor.All(ctx, &sites); err != nil {
		config.Logger.Error().Err(err).Msg("解析所有站点数据时出错")
		return nil, err
	}

	return sites, nil
}
