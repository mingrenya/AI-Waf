package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/dto"
	"github.com/mingrenya/AI-Waf/server/middleware"
	"github.com/mingrenya/AI-Waf/server/utils/response"
	"github.com/rs/zerolog"
)

// GeoIPController GeoIP 过滤控制器接口
type GeoIPController interface {
	GetConfig(ctx *gin.Context)
	UpdateConfig(ctx *gin.Context)
}

// GeoIPControllerImpl GeoIP 过滤控制器实现
type GeoIPControllerImpl struct {
	logger zerolog.Logger
}

// NewGeoIPController 创建 GeoIP 控制器
func NewGeoIPController() GeoIPController {
	return &GeoIPControllerImpl{
		logger: config.GetControllerLogger("geoip"),
	}
}

// GetConfig 获取当前 GeoIP 配置
func (c *GeoIPControllerImpl) GetConfig(ctx *gin.Context) {
	cfg := middleware.GetGeoIPConfig()
	response.Success(ctx, "获取 GeoIP 配置成功", dto.GeoIPConfig{
		BlockCountries: cfg.BlockCountries,
		AllowMode:      cfg.AllowMode,
	})
}

// UpdateConfig 更新 GeoIP 配置
func (c *GeoIPControllerImpl) UpdateConfig(ctx *gin.Context) {
	var req dto.GeoIPUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	middleware.UpdateGeoIPConfig(req.BlockCountries, req.AllowMode)

	c.logger.Info().Strs("countries", req.BlockCountries).Bool("allow_mode", req.AllowMode).Msg("GeoIP config updated")
	response.Success(ctx, "GeoIP 配置已更新", nil)
}

func joinCountryList(countries []string) string {
	if len(countries) == 0 {
		return ""
	}
	result := countries[0]
	for _, c := range countries[1:] {
		result += "," + c
	}
	return result
}
