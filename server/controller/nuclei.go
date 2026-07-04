package controller

import (
	"context"
	"sync"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
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

// ------- Cron Scanner -------

type cronScannerState struct {
	mu      sync.Mutex
	jobs    map[string]*cronJobMeta
	started bool
}

type cronJobMeta struct {
	target   string
	cronExpr string
	severity string
	lastRun  time.Time
	lastErr  string
}

var cronState = cronScannerState{jobs: make(map[string]*cronJobMeta)}

func startCronLoop(ctrl NucleiController) {
	cronState.mu.Lock()
	if cronState.started {
		cronState.mu.Unlock()
		return
	}
	cronState.started = true
	cronState.mu.Unlock()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			cronState.mu.Lock()
			for id, job := range cronState.jobs {
				if shouldRun(job) {
					runCronJob(id, job, ctrl)
				}
			}
			cronState.mu.Unlock()
		}
	}()
	config.Logger.Info().Msg("Nuclei cron scheduler started")
}

func AddCronScan(id, target, cronExpr, severity string) {
	cronState.mu.Lock()
	defer cronState.mu.Unlock()
	cronState.jobs[id] = &cronJobMeta{target: target, cronExpr: cronExpr, severity: severity}
}

func RemoveCronScan(id string) {
	cronState.mu.Lock()
	defer cronState.mu.Unlock()
	delete(cronState.jobs, id)
}

func shouldRun(job *cronJobMeta) bool {
	now := time.Now()
	switch job.cronExpr {
	case "every_1h": return now.Sub(job.lastRun) >= time.Hour
	case "every_6h": return now.Sub(job.lastRun) >= 6*time.Hour
	case "every_12h": return now.Sub(job.lastRun) >= 12*time.Hour
	case "every_24h": return now.Sub(job.lastRun) >= 24*time.Hour
	case "daily_03:00":
		if now.Hour() != 3 || now.Minute() > 1 { return false }
		return now.Sub(job.lastRun) >= 23*time.Hour
	default: return false
	}
}

func runCronJob(id string, job *cronJobMeta, ctrl NucleiController) {
	job.lastRun = time.Now()
	job.lastErr = ""

	impl, ok := ctrl.(*NucleiControllerImpl)
	if !ok || impl.scanner == nil {
		job.lastErr = "scanner unavailable"
		return
	}

	taskID := uuid.New().String()
	task := nucleiSvc.ScanTask{
		ID: taskID, Target: job.target, Severity: job.severity,
		Status: "cron", CreatedAt: time.Now(),
	}
	handler := nucleiSvc.NewResultHandler()
	if err := impl.scanner.StartScan(context.Background(), task, handler.HandleResult); err != nil {
		job.lastErr = err.Error()
		return
	}
	config.Logger.Info().Str("job_id", id).Str("target", job.target).Str("task_id", taskID).Msg("cron scan executed")
}

// // NewNucleiController 创建 Nuclei 控制器
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
	if c.scanner == nil {
		response.Error(ctx, model.NewAPIError(http.StatusServiceUnavailable, "扫描引擎未初始化，无法启动扫描", nil), true)
		return
	}
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
	if c.scanner == nil {
		response.Error(ctx, model.NewAPIError(http.StatusServiceUnavailable, "扫描引擎未初始化", nil), true)
		return
	}
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
	if c.scanner == nil {
		response.Error(ctx, model.NewAPIError(http.StatusServiceUnavailable, "扫描引擎未初始化", nil), true)
		return
	}
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
	if c.scanner == nil {
		response.Success(ctx, "扫描引擎未初始化", []dto.ScanTaskResponse{})
		return
	}
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
	if c.tmplMgr == nil {
		response.Success(ctx, "模板管理器未初始化", []any{})
		return
	}
	templates, err := c.tmplMgr.ListTemplates(ctx.Request.Context())
	if err != nil {
		response.InternalServerError(ctx, err, true)
		return
	}
	response.Success(ctx, "获取模板列表成功", templates)
}
