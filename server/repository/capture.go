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

// CaptureSession 持久化的流量捕获会话
type CaptureSession struct {
	ID           string     `bson:"_id" json:"id"`
	Interface    string     `bson:"interface" json:"interface"`
	BPFFilter    string     `bson:"bpf_filter" json:"bpf_filter"`
	Status       string     `bson:"status" json:"status"` // running/completed/stopped/error
	MaxPackets   int        `bson:"max_packets" json:"max_packets"`
	DurationSecs int        `bson:"duration_secs" json:"duration_secs"`
	Description  string     `bson:"description" json:"description"`
	PacketCount  int        `bson:"packet_count" json:"packet_count"`
	FileSize     int64      `bson:"file_size" json:"file_size"`
	FilePath     string     `bson:"file_path" json:"file_path"`
	CreatedAt    time.Time  `bson:"created_at" json:"created_at"`
	StartedAt    *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
	StoppedAt    *time.Time `bson:"stopped_at,omitempty" json:"stopped_at,omitempty"`
	ErrorMsg     string     `bson:"error_msg,omitempty" json:"error_msg,omitempty"`
}

// CaptureRepository 流量捕获数据访问接口
type CaptureRepository interface {
	CreateSession(ctx context.Context, session *CaptureSession) error
	UpdateSession(ctx context.Context, id string, updates bson.M) error
	GetSession(ctx context.Context, id string) (*CaptureSession, error)
	ListSessions(ctx context.Context, skip, limit int64) ([]CaptureSession, int64, error)
	DeleteSession(ctx context.Context, id string) error
}

// MongoCaptureRepository MongoDB 实现的捕获仓库
type MongoCaptureRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewCaptureRepository 创建捕获仓库
func NewCaptureRepository(db *mongo.Database) CaptureRepository {
	return &MongoCaptureRepository{
		collection: db.Collection("capture_sessions"),
		logger:     config.GetRepositoryLogger("capture"),
	}
}

// CreateSession 创建捕获会话记录
func (r *MongoCaptureRepository) CreateSession(ctx context.Context, session *CaptureSession) error {
	_, err := r.collection.InsertOne(ctx, session)
	return err
}

// UpdateSession 更新捕获会话
func (r *MongoCaptureRepository) UpdateSession(ctx context.Context, id string, updates bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
	return err
}

// GetSession 获取单个捕获会话
func (r *MongoCaptureRepository) GetSession(ctx context.Context, id string) (*CaptureSession, error) {
	var session CaptureSession
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessions 列出捕获会话（按创建时间倒序）
func (r *MongoCaptureRepository) ListSessions(ctx context.Context, skip, limit int64) ([]CaptureSession, int64, error) {
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
	var sessions []CaptureSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

// DeleteSession 删除捕获会话记录
func (r *MongoCaptureRepository) DeleteSession(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
