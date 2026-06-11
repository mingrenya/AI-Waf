package controller

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/rs/zerolog"
)

// FTWController WAF 回归测试控制器
type FTWController interface {
	RunTests(ctx *gin.Context)
	GetReports(ctx *gin.Context)
	GetTestFiles(ctx *gin.Context)
}

type ftwControllerImpl struct {
	service service.FTWTestService
	logger  zerolog.Logger
}

// NewFTWController 创建 FTW 控制器
func NewFTWController(svc service.FTWTestService) FTWController {
	return &ftwControllerImpl{
		service: svc,
		logger:  config.GetControllerLogger("ftw"),
	}
}

// RunTests 执行 WAF 回归测试
//
//	@Summary		执行 WAF 回归测试
//	@Description	运行 Go-FTW 测试套件，验证 WAF 拦截能力，生成覆盖率报告
//	@Tags			FTW 测试
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.FTWRunRequest	true	"测试目标"
//	@Success		200		{object}	model.SuccessResponse{data=service.FTWReport}
//	@Router			/api/v1/ftw/run [post]
func (c *ftwControllerImpl) RunTests(ctx *gin.Context) {
	var req struct {
		TargetURL string `json:"targetUrl"`
		TestDir   string `json:"testDir"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Allow empty body, use defaults
		req = struct {
			TargetURL string `json:"targetUrl"`
			TestDir   string `json:"testDir"`
		}{}
	}

	if req.TargetURL == "" {
		response.BadRequest(ctx, errors.New("targetUrl 不能为空"), true)
		return
	}

	c.logger.Info().Str("target", req.TargetURL).Msg("开始执行 FTW 测试")

	report, err := c.service.RunTests(ctx.Request.Context(), req.TargetURL, req.TestDir)
	if err != nil {
		c.logger.Error().Err(err).Msg("FTW 测试失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().
		Int("total", report.TotalTests).
		Int("passed", report.Passed).
		Float64("blockRate", report.BlockRate).
		Msg(fmt.Sprintf("FTW 测试完成：拦截率 %.1f%%", report.BlockRate))

	response.Success(ctx, "FTW 测试完成", report)
}

// GetReports 获取历史测试报告
//
//	@Summary		获取 FTW 测试历史报告
//	@Tags			FTW 测试
//	@Produce		json
//	@Param			limit	query	int	false	"返回条数"
//	@Success		200	{object}	model.SuccessResponse{data=[]service.FTWReport}
//	@Router			/api/v1/ftw/reports [get]
func (c *ftwControllerImpl) GetReports(ctx *gin.Context) {
	limit := 20
	if l, ok := ctx.GetQuery("limit"); ok {
		fmt.Sscanf(l, "%d", &limit)
	}
	reports, err := c.service.GetReports(ctx.Request.Context(), limit)
	if err != nil {
		c.logger.Error().Err(err).Msg("获取 FTW 报告失败")
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "获取成功", reports)
}

// GetTestFiles 获取测试文件列表
//
//	@Summary		获取测试文件列表
//	@Tags			FTW 测试
//	@Produce		json
//	@Success		200	{object}	model.SuccessResponse{data=[]string}
//	@Router			/api/v1/ftw/files [get]
func (c *ftwControllerImpl) GetTestFiles(ctx *gin.Context) {
	files := c.service.GetTestFiles()
	response.Success(ctx, "获取成功", files)
}
