package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/model"
	"github.com/mingrenya/AI-Waf/server/service"
	"github.com/mingrenya/AI-Waf/server/utils/response"
)

type AIChatController interface {
	Chat(ctx *gin.Context)
	ChatStream(ctx *gin.Context)
}

type aiChatControllerImpl struct {
	chatService service.AIChatService
}

// NewAIChatController 创建AI聊天控制器
func NewAIChatController(chatService service.AIChatService) AIChatController {
	return &aiChatControllerImpl{
		chatService: chatService,
	}
}

// Chat 发送聊天消息
// @Summary 发送AI聊天消息
// @Description 与AI助手进行对话
// @Tags AI
// @Accept json
// @Produce json
// @Param request body dto.AIChatRequest true "聊天请求"
// @Success 200 {object} model.SuccessResponse{data=dto.AIChatResponse}
// @Failure 400 {object} model.ErrResponseDontShowError
// @Failure 500 {object} model.ErrResponseDontShowError
// @Security BearerAuth
// @Router /ai/chat [post]
func (c *aiChatControllerImpl) Chat(ctx *gin.Context) {
	var req dto.AIChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.ErrBadRequest(err), true)
		return
	}

	// 如果请求流式响应，转发到ChatStream
	if req.Stream {
		c.ChatStream(ctx)
		return
	}

	result, err := c.chatService.Chat(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}

	response.Success(ctx, "Chat completed successfully", result)
}

// ChatStream 流式聊天响应
// @Summary 流式AI聊天
// @Description 流式返回AI聊天响应
// @Tags AI
// @Accept json
// @Produce text/event-stream
// @Param request body dto.AIChatRequest true "聊天请求"
// @Success 200 {object} dto.AIChatStreamResponse
// @Failure 400 {object} model.ErrResponseDontShowError
// @Failure 500 {object} model.ErrResponseDontShowError
// @Security BearerAuth
// @Router /ai/chat/stream [post]
func (c *aiChatControllerImpl) ChatStream(ctx *gin.Context) {
	var req dto.AIChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, model.ErrBadRequest(err), true)
		return
	}

	// 设置SSE响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")

	// 强制流式响应
	req.Stream = true

	if err := c.chatService.ChatStream(ctx.Request.Context(), &req, ctx.Writer); err != nil {
		response.Error(ctx, model.ErrInternalServerError(err), true)
		return
	}
}
