package dto

// GeoIPCountryInfo 国家信息
type GeoIPCountryInfo struct {
	ISOCode string `json:"iso_code"`
	NameZh  string `json:"name_zh"`
	NameEn  string `json:"name_en"`
	Blocked bool   `json:"blocked"`
}

// GeoIPConfig 当前 GeoIP 配置
type GeoIPConfig struct {
	BlockCountries []string `json:"block_countries"`
	AllowMode      bool     `json:"allow_mode"`
}

// GeoIPUpdateRequest 更新 GeoIP 配置
type GeoIPUpdateRequest struct {
	BlockCountries []string `json:"block_countries" binding:"required"`
	AllowMode      bool     `json:"allow_mode"`
}
