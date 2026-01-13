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
	ErrCertNotFound      = errors.New("证书不存在")
	ErrCertNameExists    = errors.New("证书名称已存在")
	ErrInvalidCertFormat = errors.New("无效的证书格式")
)

// CertificateRepository 证书仓库接口
type CertificateRepository interface {
	CreateCertificate(ctx context.Context, certificate *model.CertificateStore) error
	GetCertificates(ctx context.Context, page, size int64) ([]model.CertificateStore, int64, error)
	GetCertificateByID(ctx context.Context, id bson.ObjectID) (*model.CertificateStore, error)
	UpdateCertificate(ctx context.Context, certificate *model.CertificateStore) error
	DeleteCertificate(ctx context.Context, id bson.ObjectID) error
	CheckCertificateNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error)
}

// MongoCertificateRepository MongoDB实现的证书仓库
type MongoCertificateRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewCertificateRepository 创建证书仓库
func NewCertificateRepository(db *mongo.Database) CertificateRepository {
	var cert model.CertificateStore
	collection := db.Collection(cert.GetCollectionName())
	logger := config.GetRepositoryLogger("certificate")

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 证书名称唯一索引
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("创建证书名称索引失败")
	}

	return &MongoCertificateRepository{
		collection: collection,
		logger:     logger,
	}
}

// CreateCertificate 创建证书
func (r *MongoCertificateRepository) CreateCertificate(ctx context.Context, certificate *model.CertificateStore) error {
	// 设置创建和更新时间
	now := time.Now()
	certificate.CreatedAt = now
	certificate.UpdatedAt = now

	// 插入新证书
	result, err := r.collection.InsertOne(ctx, certificate)
	if err != nil {
		r.logger.Error().Err(err).Str("name", certificate.Name).Msg("插入证书时出错")
		return err
	}

	certificate.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// GetCertificates 获取证书列表
func (r *MongoCertificateRepository) GetCertificates(ctx context.Context, page, size int64) ([]model.CertificateStore, int64, error) {
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
		r.logger.Error().Err(err).Msg("查询证书列表时出错")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var certificates []model.CertificateStore
	if err = cursor.All(ctx, &certificates); err != nil {
		r.logger.Error().Err(err).Msg("解析证书列表时出错")
		return nil, 0, err
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("获取证书总数时出错")
		return nil, 0, err
	}

	return certificates, total, nil
}

// GetCertificateByID 根据ID获取证书
func (r *MongoCertificateRepository) GetCertificateByID(ctx context.Context, id bson.ObjectID) (*model.CertificateStore, error) {
	var certificate model.CertificateStore
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&certificate)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCertNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("查询证书时出错")
		return nil, err
	}

	return &certificate, nil
}

// UpdateCertificate 更新证书
func (r *MongoCertificateRepository) UpdateCertificate(ctx context.Context, certificate *model.CertificateStore) error {
	// 更新时间
	certificate.UpdatedAt = time.Now()

	// 更新证书
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: certificate.ID}},
		certificate,
	)

	if err != nil {
		r.logger.Error().Err(err).Str("id", certificate.ID.Hex()).Msg("更新证书时出错")
		return err
	}

	return nil
}

// DeleteCertificate 删除证书
func (r *MongoCertificateRepository) DeleteCertificate(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("删除证书时出错")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrCertNotFound
	}

	return nil
}

// CheckCertificateNameExists 检查证书名称是否已存在
func (r *MongoCertificateRepository) CheckCertificateNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error) {
	filter := bson.D{{Key: "name", Value: name}}

	// 如果是更新操作，需要排除当前证书ID
	if excludeID != bson.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$ne", Value: excludeID}}})
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("name", name).Msg("检查证书名称是否存在时出错")
		return false, err
	}

	return count > 0, nil
}
