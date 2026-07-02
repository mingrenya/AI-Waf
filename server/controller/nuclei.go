package controller

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/repository"
	nucleiSvc "github.com/mingrenya/AI-Waf/server/service/nuclei"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// NucleiController Nuclei 扫描控制器接口
type NucleiController interface {
	StartScan(ctx *gin.Context)
	GetTask(ctx *gin.Context)
	CancelTask(ctx *gin.Context)
	ListTasks(ctx *gin.Context)
	ListTemplates(ctx *gin.Context)
}

// NucleiControllerImpl Nuclei 扫描控制器实现
type NucleiControllerImpl struct {
	scanner nucleiSvc.Scanner
	repo    repository.NucleiRepository
	tmplMgr *nucleiSvc.TemplateManager
	logger  zerolog.Logger
}

// NewNucleiController 创建 Nuclei 控制器
func NewNucleiController(scanner nucleiSvc.Scanner, repo repository.NucleiRepository, tmplMgr *nucleiSvc.TemplateManager) NucleiController {
	return &NucleiControllerImpl{
		scanner: scanner,
		repo:    repo,
		tmplMgr: tmplMgr,
		logger:  config.GetControllerLogger("nuclei"),
	}
}

// StartScan 发起扫描任务
func (c *NucleiControllerImpl) StartScan(ctx *gin.Context) {
	var req dto.ScanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}
	taskID := uuid.New().String()
	handler := nucleiSvc.NewResultHandler()

	task := nucleiSvc.ScanTask{
		ID:        taskID,
		SiteID:    req.SiteID,
		Target:    req.TargetURL,
		Templates: req.Templates,
		Severity:  req.Severity,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// 持久化任务
	dbTask := &repository.NucleiTask{
		ID:        taskID,
		SiteID:    req.SiteID,
		TargetURL: req.TargetURL,
		Templates: req.Templates,
		Severity:  req.Severity,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	_ = c.repo.CreateTask(ctx, dbTask)

	if err := c.scanner.StartScan(ctx, task, handler.HandleResult); err != nil {
		_ = c.repo.UpdateTask(ctx, taskID, bson.M{"status": "failed"})
		response.InternalServerError(ctx, err, true)
		return
	}

	_ = c.repo.UpdateTask(ctx, taskID, bson.M{"status": "running"})
	response.Success(ctx, "扫描任务已启动", dto.ScanTaskResponse{
		ID:        taskID,
		SiteID:    req.SiteID,
		TargetURL: req.TargetURL,
		Status:    "running",
		CreatedAt: time.Now().Format(time.RFC3339),
	})
}

// GetTask 获取扫描任务
func (c *NucleiControllerImpl) GetTask(ctx *gin.Context) {
	id := ctx.Param("id")
	task, err := c.scanner.GetTask(id)
	if err != nil {
		response.InternalServerError(ctx, err, true)
		return
	}
	response.Success(ctx, "获取任务成功", dto.ScanTaskResponse{
		ID:        task.ID,
		SiteID:    task.SiteID,
		TargetURL: task.Target,
		Status:    task.Status,
		Findings:  task.Progress.Findings,
		Total:     task.Progress.Total,
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
	})
}

// CancelTask 取消扫描任务
func (c *NucleiControllerImpl) CancelTask(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.scanner.CancelScan(id); err != nil {
		response.InternalServerError(ctx, err, true)
		return
	}
	_ = c.repo.UpdateTask(ctx, id, bson.M{"status": "cancelled"})
	response.Success(ctx, "扫描已取消", nil)
}

// ListTasks 列出所有扫描任务
func (c *NucleiControllerImpl) ListTasks(ctx *gin.Context) {
	tasks := c.scanner.ListTasks()
	resp := make([]dto.ScanTaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, dto.ScanTaskResponse{
			ID:        t.ID,
			SiteID:    t.SiteID,
			TargetURL: t.Target,
			Status:    t.Status,
			Findings:  t.Progress.Findings,
			Total:     t.Progress.Total,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		})
	}
	response.Success(ctx, "获取任务列表成功", resp)
}

// ListTemplates 列出可用模板
func (c *NucleiControllerImpl) ListTemplates(ctx *gin.Context) {
	templates, err := c.tmplMgr.ListTemplates(ctx.Request.Context())
	if err != nil {
		response.InternalServerError(ctx, err, true)
		return
	}
	response.Success(ctx, "获取模板列表成功", templates)
}
