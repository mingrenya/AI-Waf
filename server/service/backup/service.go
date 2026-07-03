package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Service 备份服务
type Service struct {
	db      *mongo.Database
	repo    repository.BackupRepository
	backupDir string
	logger  zerolog.Logger
}

// NewService 创建备份服务
func NewService(db *mongo.Database, repo repository.BackupRepository, backupDir string) *Service {
	return &Service{
		db:        db,
		repo:      repo,
		backupDir: backupDir,
		logger:    config.GetServiceLogger("backup"),
	}
}

// CreateBackup 创建备份
func (s *Service) CreateBackup(ctx context.Context, description string, collections []string) (*repository.BackupRecord, error) {
	id := uuid.New().String()
	filename := fmt.Sprintf("backup_%s.tar.gz", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(s.backupDir, filename)

	if len(collections) == 0 {
		collections = s.listCollections(ctx)
	}

	record := &repository.BackupRecord{
		ID:          id,
		FilePath:    filepath,
		Description: description,
		Status:      "running",
		CreatedAt:   time.Now(),
	}
	if err := s.repo.CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("创建备份记录失败: %w", err)
	}

	// 执行备份
	err := s.doBackup(ctx, filepath, collections)
	if err != nil {
		record.Status = "failed"
		record.ErrorMsg = err.Error()
		s.repo.UpdateRecord(ctx, id, bson.M{"status": "failed", "error_msg": err.Error()})
		return record, err
	}

	// 获取文件大小
	if info, err := os.Stat(filepath); err == nil {
		record.FileSize = info.Size()
	}
	record.Status = "completed"
	s.repo.UpdateRecord(ctx, id, bson.M{"status": "completed", "file_size": record.FileSize})

	return record, nil
}

// RestoreBackup 恢复备份
func (s *Service) RestoreBackup(ctx context.Context, backupID string) (*repository.BackupRecord, error) {
	record, err := s.repo.GetRecord(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("备份记录不存在: %w", err)
	}
	if record.Status != "completed" {
		return nil, fmt.Errorf("备份状态为 %s，无法恢复", record.Status)
	}

	restored, err := s.doRestore(ctx, record.FilePath)
	if err != nil {
		return nil, fmt.Errorf("恢复失败: %w", err)
	}

	s.logger.Info().Str("backup_id", backupID).Int("collections", len(restored)).Msg("Backup restored")
	return record, nil
}

// ListBackups 列出备份
func (s *Service) ListBackups(ctx context.Context, skip, limit int64) ([]repository.BackupRecord, int64, error) {
	return s.repo.ListRecords(ctx, skip, limit)
}

// DeleteBackup 删除备份
func (s *Service) DeleteBackup(ctx context.Context, backupID string) error {
	record, err := s.repo.GetRecord(ctx, backupID)
	if err != nil {
		return err
	}
	// 删除文件
	os.Remove(record.FilePath)
	return s.repo.DeleteRecord(ctx, backupID)
}

// 内部方法

func (s *Service) listCollections(ctx context.Context) []string {
	names, err := s.db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil
	}
	return names
}

func (s *Service) doBackup(ctx context.Context, filepath string, collections []string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建备份文件失败: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, colName := range collections {
		cursor, err := s.db.Collection(colName).Find(ctx, bson.M{})
		if err != nil {
			s.logger.Warn().Err(err).Str("collection", colName).Msg("跳过集合")
			continue
		}

		var docs []bson.M
		if err := cursor.All(ctx, &docs); err != nil {
			cursor.Close(ctx)
			continue
		}
		cursor.Close(ctx)

		data, err := json.MarshalIndent(docs, "", "  ")
		if err != nil {
			continue
		}

		hdr := &tar.Header{
			Name: colName + ".json",
			Mode: 0644,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write(data); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) doRestore(ctx context.Context, filepath string) ([]string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	var restored []string

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}

		var docs []bson.M
		if err := json.Unmarshal(data, &docs); err != nil {
			continue
		}

		colName := hdr.Name[:len(hdr.Name)-5] // 去掉 .json
		// 清空集合后插入
		s.db.Collection(colName).DeleteMany(ctx, bson.M{})

		if len(docs) > 0 {
			ifaces := make([]interface{}, len(docs))
			for i, d := range docs {
				ifaces[i] = d
			}
			// 分批插入（每批 100）
			for i := 0; i < len(ifaces); i += 100 {
				end := i + 100
				if end > len(ifaces) {
					end = len(ifaces)
				}
				s.db.Collection(colName).InsertMany(ctx, ifaces[i:end])
			}
		}

		restored = append(restored, colName)
	}

	return restored, nil
}
