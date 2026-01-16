package controller

import (
	"net/http"
	//"time"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/mingrenya/AI-Waf/server/utils/response"
)

type MCPController struct {
	mcpService *service.MCPService
}

func NewMCPController(mcpService *service.MCPService) *MCPController {
	return &MCPController{
		mcpService: mcpService,
	}
}

// GetMCPStatus 获取MCP服务器连接状态
// @Summary 获取MCP服务器连接状态
// @Description 返回MCP服务器的实时连接状态信息
// @Tags MCP
// @Accept json
// @Produce json
// @Success 200 {object} model.Response{data=dto.MCPStatusResponse}
// @Router /api/v1/mcp/status [get]
func (c *MCPController) GetMCPStatus(ctx *gin.Context) {
	status, err := c.mcpService.GetMCPStatus(ctx)
	if err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "获取MCP状态失败", err), true)
		return
	}
	response.Success(ctx, "获取MCP状态成功", status)
}

// GetMCPTools 获取MCP可用工具列表
// @Summary 获取MCP可用工具列表
// @Description 返回所有可用的MCP工具名称
// @Tags MCP
// @Accept json
// @Produce json
// @Success 200 {object} model.Response{data=dto.MCPToolsResponse}
// @Router /api/v1/mcp/tools [get]
func (c *MCPController) GetMCPTools(ctx *gin.Context) {
	tools, err := c.mcpService.GetMCPTools(ctx)
	if err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "获取MCP工具列表失败", err), true)
		return
	}
	response.Success(ctx, "获取MCP工具列表成功", dto.MCPToolsResponse{Tools: tools})
}

// GetMCPToolCallHistory 获取MCP工具调用历史
// @Summary 获取MCP工具调用历史
// @Description 返回MCP工具的调用历史记录
// @Tags MCP
// @Accept json
// @Produce json
// @Param limit query int false "每页数量" default(50)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} model.Response{data=dto.MCPToolCallHistoryResponse}
// @Router /api/v1/mcp/tool-calls [get]
func (c *MCPController) GetMCPToolCallHistory(ctx *gin.Context) {
	var req dto.MCPToolCallHistoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "参数错误", err), true)
		return
	}

	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 50
	}

	calls, total, err := c.mcpService.GetToolCallHistory(ctx, req.Limit, req.Offset)
	if err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "获取工具调用历史失败", err), true)
		return
	}

	response.Success(ctx, "获取工具调用历史成功", dto.MCPToolCallHistoryResponse{
		Data:  calls,
		Total: total,
	})
}

// RecordToolCall 记录工具调用（供MCP Server调用）
// @Summary 记录MCP工具调用
// @Description MCP Server每次执行工具后调用此接口记录调用信息
// @Tags MCP
// @Accept json
// @Produce json
// @Param request body dto.RecordToolCallRequest true "工具调用记录"
// @Success 200 {object} model.Response
// @Router /api/v1/mcp/tool-calls/record [post]
func (c *MCPController) RecordToolCall(ctx *gin.Context) {
	var req dto.RecordToolCallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "参数错误", err), true)
		return
	}

	err := c.mcpService.RecordToolCall(ctx, req.ToolName, req.Duration, req.Success, req.Error)
	if err != nil {
		response.Error(ctx, model.NewAPIError(http.StatusInternalServerError, "记录工具调用失败", err), true)
		return
	}

	response.Success(ctx, "记录成功", nil)
}
