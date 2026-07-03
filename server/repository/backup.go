package repository

import (
	"context"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BackupRecord 备份记录
type BackupRecord struct {
	ID          string    `bson:"_id" json:"id"`
	FilePath    string    `bson:"file_path" json:"file_path"`
	FileSize    int64     `bson:"file_size" json:"file_size"`
	Description string    `bson:"description" json:"description"`
	Status      string    `bson:"status" json:"status"`
	ErrorMsg    string    `bson:"error_msg,omitempty" json:"error_msg,omitempty"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
}

// BackupRepository 备份数据访问接口
type BackupRepository interface {
	CreateRecord(ctx context.Context, record *BackupRecord) error
	UpdateRecord(ctx context.Context, id string, updates bson.M) error
	GetRecord(ctx context.Context, id string) (*BackupRecord, error)
	ListRecords(ctx context.Context, skip, limit int64) ([]BackupRecord, int64, error)
	DeleteRecord(ctx context.Context, id string) error
}

type MongoBackupRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

func NewBackupRepository(db *mongo.Database) BackupRepository {
	return &MongoBackupRepository{
		collection: db.Collection("backup_records"),
		logger:     config.GetRepositoryLogger("backup"),
	}
}

func (r *MongoBackupRepository) CreateRecord(ctx context.Context, record *BackupRecord) error {
	_, err := r.collection.InsertOne(ctx, record)
	return err
}

func (r *MongoBackupRepository) UpdateRecord(ctx context.Context, id string, updates bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
	return err
}

func (r *MongoBackupRepository) GetRecord(ctx context.Context, id string) (*BackupRecord, error) {
	var record BackupRecord
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *MongoBackupRepository) ListRecords(ctx context.Context, skip, limit int64) ([]BackupRecord, int64, error) {
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var records []BackupRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *MongoBackupRepository) DeleteRecord(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
