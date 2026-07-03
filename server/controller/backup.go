package controller

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	backupSvc "github.com/mingrenya/AI-Waf/server/service/backup"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/rs/zerolog"
)

// BackupController 备份控制器接口
type BackupController interface {
	CreateBackup(ctx *gin.Context)
	ListBackups(ctx *gin.Context)
	RestoreBackup(ctx *gin.Context)
	DeleteBackup(ctx *gin.Context)
	DownloadBackup(ctx *gin.Context)
}

// BackupControllerImpl 备份控制器实现
type BackupControllerImpl struct {
	svc    *backupSvc.Service
	logger zerolog.Logger
}

// NewBackupController 创建备份控制器
func NewBackupController(svc *backupSvc.Service) BackupController {
	return &BackupControllerImpl{
		svc:    svc,
		logger: config.GetControllerLogger("backup"),
	}
}

// CreateBackup 创建备份
func (c *BackupControllerImpl) CreateBackup(ctx *gin.Context) {
	var req struct {
		Description string   `json:"description"`
		Collections []string `json:"collections"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	record, err := c.svc.CreateBackup(ctx, req.Description, req.Collections)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "备份创建成功", dto.BackupResponse{
		ID:          record.ID,
		FilePath:    record.FilePath,
		FileSize:    record.FileSize,
		Description: record.Description,
		Status:      record.Status,
		CreatedAt:   record.CreatedAt.Format(time.RFC3339),
	})
}

// ListBackups 列出备份
func (c *BackupControllerImpl) ListBackups(ctx *gin.Context) {
	records, total, err := c.svc.ListBackups(ctx, 0, 50)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	resp := make([]dto.BackupResponse, 0, len(records))
	for _, r := range records {
		resp = append(resp, dto.BackupResponse{
			ID:          r.ID,
			FilePath:    r.FilePath,
			FileSize:    r.FileSize,
			Description: r.Description,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.Format(time.RFC3339),
			ErrorMsg:    r.ErrorMsg,
		})
	}
	response.Success(ctx, "获取备份列表成功", map[string]interface{}{
		"backups": resp,
		"total":   total,
	})
}

// RestoreBackup 恢复备份
func (c *BackupControllerImpl) RestoreBackup(ctx *gin.Context) {
	id := ctx.Param("id")
	record, err := c.svc.RestoreBackup(ctx, id)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "备份恢复成功", dto.RestoreResponse{
		Success:  true,
		BackupID: record.ID,
		Note:     "数据已从备份恢复",
	})
}

// DeleteBackup 删除备份
func (c *BackupControllerImpl) DeleteBackup(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.svc.DeleteBackup(ctx, id); err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "备份已删除", nil)
}

// DownloadBackup 下载备份文件
func (c *BackupControllerImpl) DownloadBackup(ctx *gin.Context) {
	id := ctx.Param("id")
	record, _, err := c.svc.ListBackups(ctx, 0, 1)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	var filePath string
	for _, r := range record {
		if r.ID == id {
			filePath = r.FilePath
			break
		}
	}
	if filePath == "" {
		response.InternalServerError(ctx, nil, false)
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.InternalServerError(ctx, nil, false)
		return
	}

	ctx.File(filePath)
}
