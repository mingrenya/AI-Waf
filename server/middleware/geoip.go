package middleware

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	mongodb "github.com/mingrenya/AI-Waf/pkg/database/mongo"
	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/mingrenya/AI-Waf/server/model"
)


var geoipMu sync.RWMutex
var geoipCfg = struct {
	blockCountries []string
	allowMode      bool
	initialized    bool
}{}

// GeoIPBlock 按国家/地区封禁请求
// 通过 GEOIP_BLOCK_COUNTRIES 环境变量配置（ISO 代码，逗号分隔，如 "CN,RU,KP"）
// GEOIP_ALLOW_MODE 设为 "true" 时切换为白名单模式
func GeoIPBlock() gin.HandlerFunc {
	geoipMu.RLock()
	blockCountries := geoipCfg.blockCountries
	allowMode := geoipCfg.allowMode
	geoipMu.RUnlock()
	if !geoipCfg.initialized {
		blockCountries = parseCountryList(os.Getenv("GEOIP_BLOCK_COUNTRIES"))
		allowMode = os.Getenv("GEOIP_ALLOW_MODE") == "true"
	}

	if len(blockCountries) == 0 {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		clientIP := extractClientIP(c)
		country := lookupCountryFromContext(c)
		_ = clientIP // 记录 IP 用于日志上下文

		if country == "" {
			c.Next()
			return
		}

		blocked := containsCountry(blockCountries, country)
		match := blocked
		if allowMode {
			match = !blocked
		}

		if match {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(http.StatusForbidden,
				"访问被拒绝：您所在地区不可访问此服务", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractClientIP 提取客户端真实 IP
func extractClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	host, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return host
}

// lookupCountryFromContext 从 Gin context 中查找国家信息
func lookupCountryFromContext(c *gin.Context) string {
	if country, ok := c.Get("geo_country"); ok {
		if s, ok := country.(string); ok {
			return s
		}
	}
	return ""
}

func parseCountryList(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var cleaned []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToUpper(p))
		if len(p) == 2 {
			cleaned = append(cleaned, p)
		}
	}
	return cleaned
}

func containsCountry(list []string, country string) bool {
	country = strings.ToUpper(strings.TrimSpace(country))
	for _, c := range list {
		if c == country {
			return true
		}
	}
	return false
}

// GeoIPConfig 当前 GeoIP 配置（供前端 API 调用）
type GeoIPConfig struct {
	BlockCountries []string `json:"block_countries"`
	AllowMode      bool     `json:"allow_mode"`
}

// GetGeoIPConfig 获取当前 GeoIP 配置
func GetGeoIPConfig() GeoIPConfig {
	geoipMu.RLock()
	defer geoipMu.RUnlock()
	return GeoIPConfig{
		BlockCountries: geoipCfg.blockCountries,
		AllowMode:      geoipCfg.allowMode,
	}
}

// UpdateGeoIPConfig 更新 GeoIP 配置
func UpdateGeoIPConfig(countries []string, allowMode bool) {
	geoipMu.Lock()
	geoipCfg.blockCountries = countries
	geoipCfg.allowMode = allowMode
	geoipCfg.initialized = true
	geoipMu.Unlock()

	// Async persist to MongoDB
	go func() {
		client, err := mongodb.Connect(config.Global.DBConfig.URI)
		if err != nil || client == nil { return }
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		collection := client.Database(config.Global.DBConfig.Database).Collection("config")
		collection.UpdateOne(ctx,
			map[string]interface{}{"name": "AppConfig"},
			map[string]interface{}{"$set": map[string]interface{}{
				"geoip": map[string]interface{}{"blockCountries": countries, "allowMode": allowMode},
			}},
		)
	}()
}
