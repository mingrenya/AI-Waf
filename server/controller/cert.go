package controller

import (
	"errors"
	"net/http"

	"github.com/HUAHUAI23/RuiQi/server/config"
	"github.com/HUAHUAI23/RuiQi/server/dto"
	"github.com/HUAHUAI23/RuiQi/server/model"
	"github.com/HUAHUAI23/RuiQi/server/service"
	"github.com/HUAHUAI23/RuiQi/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// CertificateController 证书控制器接口
type CertificateController interface {
	CreateCertificate(ctx *gin.Context)
	GetCertificates(ctx *gin.Context)
	GetCertificateByID(ctx *gin.Context)
	UpdateCertificate(ctx *gin.Context)
	DeleteCertificate(ctx *gin.Context)
}

// CertificateControllerImpl 证书控制器实现
type CertificateControllerImpl struct {
	certService service.CertificateService
	logger      zerolog.Logger
}

// NewCertificateController 创建证书控制器
func NewCertificateController(certService service.CertificateService) CertificateController {
	logger := config.GetControllerLogger("certificate")
	return &CertificateControllerImpl{
		certService: certService,
		logger:      logger,
	}
}

// CreateCertificate 创建证书
//
//	@Summary		创建新证书
//	@Description	创建一个新的SSL/TLS证书
//	@Tags			证书管理
//	@Accept			json
//	@Produce		json
//	@Param			certificate	body	dto.CertificateCreateRequest	true	"证书信息"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.CertificateStore}	"证书创建成功"
//	@Failure		400	{object}	model.ErrResponse									"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError						"禁止访问"
//	@Failure		409	{object}	model.ErrResponseDontShowError						"证书名称已存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"服务器内部错误"
//	@Router			/api/v1/certificates [post]
func (c *CertificateControllerImpl) CreateCertificate(ctx *gin.Context) {
	var req dto.CertificateCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("name", req.Name).Msg("创建证书请求")
	cert, err := c.certService.CreateCertificate(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrCertificateNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "证书名称已存在", err), false)
			return
		} else if errors.Is(err, service.ErrInvalidCertificate) {
			response.BadRequest(ctx, err, true)
			return
		}
		c.logger.Error().Err(err).Msg("创建证书失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	// 转换为DTO响应
	certResponse := model.CertificateStore{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		PublicKey:   cert.PublicKey,
		PrivateKey:  cert.PrivateKey,
		ExpireDate:  cert.ExpireDate,
		IssuerName:  cert.IssuerName,
		FingerPrint: cert.FingerPrint,
		Domains:     cert.Domains,
		CreatedAt:   cert.CreatedAt,
		UpdatedAt:   cert.UpdatedAt,
	}

	c.logger.Info().Str("id", cert.ID.Hex()).Str("name", cert.Name).Msg("证书创建成功")
	response.Success(ctx, "证书创建成功", certResponse)
}

// GetCertificates 获取证书列表
//
//	@Summary		获取证书列表
//	@Description	获取所有SSL/TLS证书列表，支持分页
//	@Tags			证书管理
//	@Produce		json
//	@Param			page	query	int	false	"页码"	default(1)
//	@Param			size	query	int	false	"每页数量"	default(10)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.CertificateListResponse}	"获取证书列表成功"
//	@Failure		401	{object}	model.ErrResponseDontShowError							"未授权访问"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"服务器内部错误"
//	@Router			/api/v1/certificates [get]
func (c *CertificateControllerImpl) GetCertificates(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")

	c.logger.Info().Str("page", page).Str("size", size).Msg("获取证书列表请求")
	certificates, total, err := c.certService.GetCertificates(ctx, page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("获取证书列表失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Int64("total", total).Msg("获取证书列表成功")
	response.Success(ctx, "获取证书列表成功", gin.H{
		"total": total,
		"items": certificates,
	})
}

// GetCertificateByID 获取单个证书
//
//	@Summary		获取单个证书
//	@Description	根据ID获取证书详情
//	@Tags			证书管理
//	@Produce		json
//	@Param			id	path	string	true	"证书ID"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.CertificateStore}	"获取证书详情成功"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"未授权访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError						"证书不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"服务器内部错误"
//	@Router			/api/v1/certificates/{id} [get]
func (c *CertificateControllerImpl) GetCertificateByID(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("获取证书详情请求")

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	cert, err := c.certService.GetCertificateByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrCertificateNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("获取证书详情失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	// 转换为DTO响应
	certResponse := model.CertificateStore{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		PublicKey:   cert.PublicKey,
		PrivateKey:  cert.PrivateKey,
		ExpireDate:  cert.ExpireDate,
		IssuerName:  cert.IssuerName,
		FingerPrint: cert.FingerPrint,
		Domains:     cert.Domains,
		CreatedAt:   cert.CreatedAt,
		UpdatedAt:   cert.UpdatedAt,
	}

	c.logger.Info().Str("id", id).Str("name", cert.Name).Msg("获取证书详情成功")
	response.Success(ctx, "获取证书详情成功", certResponse)
}

// UpdateCertificate 更新证书
//
//	@Summary		更新证书
//	@Description	更新指定证书的信息
//	@Tags			证书管理
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string							true	"证书ID"
//	@Param			certificate	body	dto.CertificateUpdateRequest	true	"证书更新信息"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.CertificateStore}	"证书更新成功"
//	@Failure		400	{object}	model.ErrResponse									"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError						"禁止访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError						"证书不存在"
//	@Failure		409	{object}	model.ErrResponseDontShowError						"证书名称已存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"服务器内部错误"
//	@Router			/api/v1/certificates/{id} [put]
func (c *CertificateControllerImpl) UpdateCertificate(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.CertificateUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Str("id", id).Msg("请求参数绑定失败")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("id", id).Msg("更新证书请求")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	cert, err := c.certService.UpdateCertificate(ctx, objectID, &req)
	if err != nil {
		if errors.Is(err, service.ErrCertificateNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrCertificateNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "证书名称已存在", err), false)
			return
		} else if errors.Is(err, service.ErrInvalidCertificate) {
			response.BadRequest(ctx, err, true)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("更新证书失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	// 转换为DTO响应
	certResponse := model.CertificateStore{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		PublicKey:   cert.PublicKey,
		PrivateKey:  cert.PrivateKey,
		ExpireDate:  cert.ExpireDate,
		IssuerName:  cert.IssuerName,
		FingerPrint: cert.FingerPrint,
		Domains:     cert.Domains,
		CreatedAt:   cert.CreatedAt,
		UpdatedAt:   cert.UpdatedAt,
	}

	c.logger.Info().Str("id", id).Str("name", cert.Name).Msg("证书更新成功")
	response.Success(ctx, "证书更新成功", certResponse)
}

// ... existing code ...

// DeleteCertificate 删除证书
//
//	@Summary		删除证书
//	@Description	删除指定的SSL/TLS证书
//	@Tags			证书管理
//	@Produce		json
//	@Param			id	path	string	true	"证书ID"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"证书删除成功"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"禁止访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"证书不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"服务器内部错误"
//	@Router			/api/v1/certificates/{id} [delete]
func (c *CertificateControllerImpl) DeleteCertificate(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("删除证书请求")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("无效的ID格式")
		response.BadRequest(ctx, err, true)
		return
	}
	err = c.certService.DeleteCertificate(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrCertificateNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("删除证书失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Msg("证书删除成功")
	response.Success(ctx, "证书删除成功", nil)
}
