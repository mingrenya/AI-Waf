package dto

import "time"

// 时间范围常量
const (
	TimeRange24Hours = "24h"
	TimeRange7Days   = "7d"
	TimeRange30Days  = "30d"
)

// StatsRequest 统计数据请求
// @Description 统计数据请求参数
type StatsRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"` // 时间范围: 24h, 7d, 30d
}

// RealtimeQPSRequest 实时QPS请求
// @Description 实时QPS数据请求参数
type RealtimeQPSRequest struct {
	Limit int `json:"limit" form:"limit" binding:"omitempty,min=1,max=60" default:"30" example:"30"` // 数据点数量限制，默认30个点，最大60个点
}

// TimeSeriesDataRequest 时间序列数据请求
// @Description 时间序列数据请求参数
type TimeSeriesDataRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"`     // 时间范围: 24h, 7d, 30d
	Metric    string `json:"metric" form:"metric" binding:"required,oneof=requests blocks" example:"requests"` // 指标类型: requests(请求数), blocks(拦截数)
}

// TrafficTimeSeriesRequest 流量时间序列数据请求
// @Description 流量时间序列数据请求参数
type TrafficTimeSeriesRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"` // 时间范围: 24h, 7d, 30d
}

// OverviewStats 概览统计数据
// @Description 概览统计数据，包含各项关键指标
type OverviewStats struct {
	// 时间范围
	TimeRange string `json:"timeRange" example:"24h"` // 统计数据的时间范围

	// 流量统计
	TotalRequests   int64 `json:"totalRequests" example:"123456"`     // 总请求数
	InboundTraffic  int64 `json:"inboundTraffic" example:"67890123"`  // 入站流量(字节)
	OutboundTraffic int64 `json:"outboundTraffic" example:"12345678"` // 出站流量(字节)
	MaxQPS          int64 `json:"maxQPS" example:"150"`               // 最大QPS

	// 错误统计
	Error4xx     int64   `json:"error4xx" example:"234"`      // 4xx错误数量
	Error4xxRate float64 `json:"error4xxRate" example:"0.23"` // 4xx错误率(百分比)
	Error5xx     int64   `json:"error5xx" example:"45"`       // 5xx错误数量
	Error5xxRate float64 `json:"error5xxRate" example:"0.05"` // 5xx错误率(百分比)

	// 安全统计
	BlockCount    int64 `json:"blockCount" example:"123"`   // 拦截数量
	AttackIPCount int64 `json:"attackIPCount" example:"45"` // 攻击IP数量
}

// RealtimeQPSData 实时QPS数据
// @Description 实时QPS数据点，包含时间戳和值
type RealtimeQPSData struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:30:45Z"` // 时间戳
	Value     int64     `json:"value" example:"120"`                      // QPS值
}

// RealtimeQPSResponse 实时QPS响应
// @Description 实时QPS数据响应
type RealtimeQPSResponse struct {
	Data []RealtimeQPSData `json:"data"` // QPS数据点列表
}

// TimeSeriesDataPoint 时间序列数据点
// @Description 时间序列图表数据点
type TimeSeriesDataPoint struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"` // 时间戳，表示数据点的时间
	Value     int64     `json:"value" example:"128"`                      // 数值，表示该时间点的指标值
}

// TrafficDataPoint 流量数据点
// @Description 流量时间序列图表数据点
type TrafficDataPoint struct {
	Timestamp       time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"` // 时间戳，表示数据点的时间
	InboundTraffic  int64     `json:"inboundTraffic" example:"1024000"`         // 入站流量(字节)
	OutboundTraffic int64     `json:"outboundTraffic" example:"2048000"`        // 出站流量(字节)
}

// TimeSeriesResponse 时间序列响应
// @Description 时间序列图表数据响应
type TimeSeriesResponse struct {
	Metric    string                `json:"metric" example:"requests"` // 指标名称
	TimeRange string                `json:"timeRange" example:"24h"`   // 时间范围
	Data      []TimeSeriesDataPoint `json:"data"`                      // 数据点列表
}

// TrafficTimeSeriesResponse 流量时间序列响应
// @Description 流量时间序列图表数据响应
type TrafficTimeSeriesResponse struct {
	TimeRange string             `json:"timeRange" example:"24h"` // 时间范围
	Data      []TrafficDataPoint `json:"data"`                    // 流量数据点列表
}

// CombinedTimeSeriesResponse 组合时间序列响应
// @Description 同时包含请求数和拦截数的时间序列数据
type CombinedTimeSeriesResponse struct {
	TimeRange string             `json:"timeRange" example:"24h"` // 时间范围
	Requests  TimeSeriesResponse `json:"requests"`                // 请求数时间序列
	Blocks    TimeSeriesResponse `json:"blocks"`                  // 拦截数时间序列
}

// ========== 综合安全指标 Dashboard API ==========

// SecurityMetricsRequest 综合安全指标请求
// @Description 综合安全指标请求参数
type SecurityMetricsRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"` // 时间范围: 24h, 7d, 30d
}

// RuleTriggerStats 规则触发统计
// @Description 规则触发统计信息
type RuleTriggerStats struct {
	RuleID     int64  `json:"ruleId" example:"10086"`                  // 规则ID
	RuleName   string `json:"ruleName" example:"SQL注入防护规则"`             // 规则名称
	Count      int64  `json:"count" example:"128"`                     // 触发次数
	Percentage float64 `json:"percentage" example:"12.5"`              // 占比(百分比)
}

// SeverityStats 严重等级统计
// @Description 攻击严重等级分布统计
type SeverityStats struct {
	Level      int64   `json:"level" example:"2"`       // 严重级别(0-5)
	LevelName  string  `json:"levelName" example:"中等"`  // 级别名称
	Count      int64   `json:"count" example:"256"`     // 数量
	Percentage float64 `json:"percentage" example:"25.6"` // 占比(百分比)
}

// AttackTypeStats 攻击类型统计
// @Description 攻击类型分布统计
type AttackTypeStats struct {
	Category   string  `json:"category" example:"SQL注入"`       // 攻击类别
	Count      int64   `json:"count" example:"512"`            // 数量
	Percentage float64 `json:"percentage" example:"51.2"`      // 占比(百分比)
}

// GeoLocationStats 地理位置统计
// @Description 攻击来源地理位置统计
type GeoLocationStats struct {
	Country    string  `json:"country" example:"United States"` // 国家
	CountryCode string  `json:"countryCode" example:"US"`       // 国家代码
	City       string  `json:"city" example:"New York"`        // 城市
	Count      int64   `json:"count" example:"128"`            // 攻击次数
	Percentage float64 `json:"percentage" example:"12.8"`      // 占比(百分比)
}

// RuleEngineStats 规则引擎统计
// @Description 规则引擎性能和效率统计
type RuleEngineStats struct {
	TotalRules       int64   `json:"totalRules" example:"128"`       // 总规则数
	EnabledRules     int64   `json:"enabledRules" example:"120"`     // 已启用规则数
	DisabledRules    int64   `json:"disabledRules" example:"8"`      // 已禁用规则数
	WhitelistRules   int64   `json:"whitelistRules" example:"20"`    // 白名单规则数
	BlacklistRules   int64   `json:"blacklistRules" example:"100"`   // 黑名单规则数
	AvgMatchTime     float64 `json:"avgMatchTime" example:"0.5"`     // 平均匹配时间(毫秒)
	RuleEfficiency   float64 `json:"ruleEfficiency" example:"95.5"`  // 规则效率(百分比)
}

// BlockedIPStats 封禁IP统计
// @Description IP封禁统计信息
type BlockedIPStats struct {
	TotalBlocked          int64 `json:"totalBlocked" example:"256"`          // 总封禁IP数
	ActiveBlocked         int64 `json:"activeBlocked" example:"128"`         // 当前活跃封禁数
	ExpiredBlocked        int64 `json:"expiredBlocked" example:"128"`        // 已过期封禁数
	HighFrequencyVisit    int64 `json:"highFrequencyVisit" example:"50"`     // 高频访问封禁数
	HighFrequencyAttack   int64 `json:"highFrequencyAttack" example:"70"`    // 高频攻击封禁数
	HighFrequencyError    int64 `json:"highFrequencyError" example:"8"`      // 高频错误封禁数
}

// ThreatLevelDistribution 威胁等级分布
// @Description 当前威胁等级分布统计
type ThreatLevelDistribution struct {
	Critical int64 `json:"critical" example:"5"`   // 严重威胁
	High     int64 `json:"high" example:"15"`      // 高威胁
	Medium   int64 `json:"medium" example:"30"`    // 中等威胁
	Low      int64 `json:"low" example:"50"`       // 低威胁
}

// ResponseTimeStats 响应时间统计
// @Description WAF响应时间统计
type ResponseTimeStats struct {
	AvgResponseTime float64 `json:"avgResponseTime" example:"15.5"` // 平均响应时间(毫秒)
	MaxResponseTime float64 `json:"maxResponseTime" example:"250"`  // 最大响应时间(毫秒)
	MinResponseTime float64 `json:"minResponseTime" example:"5"`    // 最小响应时间(毫秒)
	P50ResponseTime float64 `json:"p50ResponseTime" example:"10"`   // P50响应时间(毫秒)
	P95ResponseTime float64 `json:"p95ResponseTime" example:"50"`   // P95响应时间(毫秒)
	P99ResponseTime float64 `json:"p99ResponseTime" example:"100"`  // P99响应时间(毫秒)
}

// SecurityMetricsResponse 综合安全指标响应
// @Description 综合安全指标仪表板数据响应
type SecurityMetricsResponse struct {
	TimeRange               string                   `json:"timeRange" example:"24h"`               // 统计时间范围
	Overview                OverviewStats            `json:"overview"`                              // 概览统计
	RuleEngine              RuleEngineStats          `json:"ruleEngine"`                            // 规则引擎统计
	TopTriggeredRules       []RuleTriggerStats       `json:"topTriggeredRules"`                     // Top触发规则(前10)
	SeverityDistribution    []SeverityStats          `json:"severityDistribution"`                  // 严重等级分布
	AttackTypeDistribution  []AttackTypeStats        `json:"attackTypeDistribution"`                // 攻击类型分布
	TopAttackSources        []GeoLocationStats       `json:"topAttackSources"`                      // Top攻击来源(前10)
	BlockedIPMetrics        BlockedIPStats           `json:"blockedIPMetrics"`                      // 封禁IP指标
	ThreatLevel             ThreatLevelDistribution  `json:"threatLevel"`                           // 威胁等级分布
	ResponseTime            ResponseTimeStats        `json:"responseTime"`                          // 响应时间统计
	RequestTrend            TimeSeriesResponse       `json:"requestTrend"`                          // 请求趋势
	BlockTrend              TimeSeriesResponse       `json:"blockTrend"`                            // 拦截趋势
	TrafficTrend            TrafficTimeSeriesResponse `json:"trafficTrend"`                          // 流量趋势
}
