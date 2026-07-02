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

// NucleiTask 持久化的扫描任务
type NucleiTask struct {
	ID          string          `bson:"_id" json:"id"`
	SiteID      string          `bson:"site_id" json:"site_id"`
	TargetURL   string          `bson:"target_url" json:"target_url"`
	Templates   []string        `bson:"templates" json:"templates"`
	Severity    string          `bson:"severity" json:"severity"`
	Status      string          `bson:"status" json:"status"`
	Findings    []NucleiFinding `bson:"findings" json:"findings"`
	CreatedAt   time.Time       `bson:"created_at" json:"created_at"`
	StartedAt   *time.Time      `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt *time.Time      `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// NucleiFinding 扫描发现
type NucleiFinding struct {
	TemplateID       string   `bson:"template_id" json:"template_id"`
	Name             string   `bson:"name" json:"name"`
	Severity         string   `bson:"severity" json:"severity"`
	MatchedAt        string   `bson:"matched_at" json:"matched_at"`
	CurlCommand      string   `bson:"curl_command" json:"curl_command"`
	ExtractedResults []string `bson:"extracted_results" json:"extracted_results"`
}

// NucleiRepository Nuclei 数据访问接口
type NucleiRepository interface {
	CreateTask(ctx context.Context, task *NucleiTask) error
	UpdateTask(ctx context.Context, id string, updates bson.M) error
	GetTask(ctx context.Context, id string) (*NucleiTask, error)
	ListTasks(ctx context.Context, siteID string, skip, limit int64) ([]NucleiTask, int64, error)
}

// MongoNucleiRepository MongoDB 实现的 Nuclei 仓库
type MongoNucleiRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewNucleiRepository 创建 Nuclei 仓库
func NewNucleiRepository(db *mongo.Database) NucleiRepository {
	return &MongoNucleiRepository{
		collection: db.Collection("nuclei_tasks"),
		logger:     config.GetRepositoryLogger("nuclei"),
	}
}

// CreateTask 创建扫描任务
func (r *MongoNucleiRepository) CreateTask(ctx context.Context, task *NucleiTask) error {
	_, err := r.collection.InsertOne(ctx, task)
	return err
}

// UpdateTask 更新扫描任务
func (r *MongoNucleiRepository) UpdateTask(ctx context.Context, id string, updates bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
	return err
}

// GetTask 获取单个扫描任务
func (r *MongoNucleiRepository) GetTask(ctx context.Context, id string) (*NucleiTask, error) {
	var task NucleiTask
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 列出扫描任务（支持按 siteID 过滤和分页）
func (r *MongoNucleiRepository) ListTasks(ctx context.Context, siteID string, skip, limit int64) ([]NucleiTask, int64, error) {
	filter := bson.M{}
	if siteID != "" {
		filter["site_id"] = siteID
	}
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var tasks []NucleiTask
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}
