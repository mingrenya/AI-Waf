package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
	captureSvc "github.com/mingrenya/AI-Waf/server/service/capture"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/rs/zerolog"
)

// CaptureController 流量捕获控制器接口
type CaptureController interface {
	StartCapture(ctx *gin.Context)
	StopCapture(ctx *gin.Context)
	GetSession(ctx *gin.Context)
	ListSessions(ctx *gin.Context)
	DownloadPCAP(ctx *gin.Context)
	GetForensicsStats(ctx *gin.Context)
}

// CaptureControllerImpl 流量捕获控制器实现
type CaptureControllerImpl struct {
	svc       *captureSvc.CaptureService
	forensics *captureSvc.ForensicsCapture
	logger    zerolog.Logger
}

// NewCaptureController 创建捕获控制器
func NewCaptureController(svc *captureSvc.CaptureService) CaptureController {
	return &CaptureControllerImpl{
		svc:    svc,
		logger: config.GetControllerLogger("capture"),
	}
}

func NewCaptureControllerWithForensics(svc *captureSvc.CaptureService, forensics *captureSvc.ForensicsCapture) CaptureController {
	return &CaptureControllerImpl{
		svc:       svc,
		forensics: forensics,
		logger:    config.GetControllerLogger("capture"),
	}
}

// StartCapture 启动流量捕获
func (c *CaptureControllerImpl) StartCapture(ctx *gin.Context) {
	var req dto.StartCaptureRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.ErrBadRequest(err), true)
		return
	}
	session, err := c.svc.StartCapture(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}
	response.Success(ctx, "流量捕获已启动", session)
}

// StopCapture 停止流量捕获
func (c *CaptureControllerImpl) StopCapture(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.svc.StopCapture(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}
	session, _ := c.svc.GetSession(ctx.Request.Context(), id)
	response.Success(ctx, "流量捕获已停止", session)
}

// GetSession 获取捕获会话详情
func (c *CaptureControllerImpl) GetSession(ctx *gin.Context) {
	id := ctx.Param("id")
	session, err := c.svc.GetSession(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), false)
		return
	}
	response.Success(ctx, "获取捕获会话成功", session)
}

// ListSessions 列出捕获会话
func (c *CaptureControllerImpl) ListSessions(ctx *gin.Context) {
	sessions, total, err := c.svc.ListSessions(ctx.Request.Context(), 0, 50)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), false)
		return
	}
	response.Success(ctx, "获取捕获会话列表成功", dto.CaptureListResponse{
		Sessions: sessions,
		Total:    total,
	})
}

// DownloadPCAP 下载 PCAP 文件
func (c *CaptureControllerImpl) DownloadPCAP(ctx *gin.Context) {
	id := ctx.Param("id")
	filePath, err := c.svc.GetCaptureFile(id)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), false)
		return
	}
	fileName := fmt.Sprintf("capture_%s.pcap", id)
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	ctx.Header("Content-Type", "application/vnd.tcpdump.pcap")
	ctx.File(filePath)
}

func (c *CaptureControllerImpl) GetForensicsStats(ctx *gin.Context) {
	if c.forensics == nil {
		response.Success(ctx, "forensics capture not initialized", map[string]interface{}{"enabled": false})
		return
	}
	response.Success(ctx, "forensics capture stats", c.forensics.Stats())
}
