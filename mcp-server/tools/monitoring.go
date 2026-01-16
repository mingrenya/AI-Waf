// tools/monitoring.go
// 实时监控工具
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRealtimeQPSInput 获取实时QPS的输入参数
type GetRealtimeQPSInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"返回的数据点数量,默认30,最大60"`
}

// GetRealtimeQPSOutput 实时QPS输出
type GetRealtimeQPSOutput struct {
	DataPoints []QPSDataPoint `json:"dataPoints" jsonschema:"QPS数据点列表"`
	Current    float64        `json:"current" jsonschema:"当前QPS"`
	Avg        float64        `json:"avg" jsonschema:"平均QPS"`
	Peak       float64        `json:"peak" jsonschema:"峰值QPS"`
}

// QPSDataPoint QPS数据点
type QPSDataPoint struct {
	Timestamp int64   `json:"timestamp" jsonschema:"时间戳"`
	Value     float64 `json:"value" jsonschema:"QPS值"`
}

// CreateGetRealtimeQPS 创建获取实时QPS的工具函数
func CreateGetRealtimeQPS(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetRealtimeQPSInput) (*mcp.CallToolResult, GetRealtimeQPSOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetRealtimeQPSInput) (*mcp.CallToolResult, GetRealtimeQPSOutput, error) {
		logger := NewToolLogger("get_realtime_qps")
		logger.LogInput(input)
		
		if input.Limit == 0 {
			input.Limit = 30
		}
		if input.Limit > 60 {
			input.Limit = 60
		}
		
		// 使用实际的API路径 /api/v1/stats/realtime-qps
		path := fmt.Sprintf("/api/v1/stats/realtime-qps?limit=%d", input.Limit)
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, GetRealtimeQPSOutput{}, fmt.Errorf("获取实时QPS失败: %w", err)
		}

		var result struct {
			Data GetRealtimeQPSOutput `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, GetRealtimeQPSOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess(fmt.Sprintf("获取%d个QPS数据点", len(result.Data.DataPoints)))
		return nil, result.Data, nil
	}
}

// GetTimeSeriesDataInput 获取时间序列数据的输入参数
type GetTimeSeriesDataInput struct {
	MetricType string `json:"metricType" jsonschema:"指标类型：requests,errors,responseTime"`
	TimeRange  string `json:"timeRange" jsonschema:"时间范围：1h,6h,24h,7d,30d"`
	Interval   string `json:"interval,omitempty" jsonschema:"数据间隔：1m,5m,1h,1d"`
}

// GetTimeSeriesDataOutput 时间序列数据输出
type GetTimeSeriesDataOutput struct {
	MetricType string             `json:"metricType" jsonschema:"指标类型"`
	DataPoints []TimeSeriesPoint  `json:"dataPoints" jsonschema:"时间序列数据点"`
	Summary    map[string]float64 `json:"summary" jsonschema:"统计摘要"`
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp" jsonschema:"时间戳"`
	Value     float64 `json:"value" jsonschema:"指标值"`
}

// CreateGetTimeSeriesData 创建获取时间序列数据的工具函数
func CreateGetTimeSeriesData(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetTimeSeriesDataInput) (*mcp.CallToolResult, GetTimeSeriesDataOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetTimeSeriesDataInput) (*mcp.CallToolResult, GetTimeSeriesDataOutput, error) {
		logger := NewToolLogger("get_time_series_data")
		logger.LogInput(input)
		
		if input.TimeRange == "" {
			input.TimeRange = "24h"
		}
		
		// 使用实际的API路径 /api/v1/stats/time-series
		path := fmt.Sprintf("/api/v1/stats/time-series?metricType=%s&timeRange=%s", 
			input.MetricType, input.TimeRange)
		if input.Interval != "" {
			path += "&interval=" + input.Interval
		}
		
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, GetTimeSeriesDataOutput{}, fmt.Errorf("获取时间序列数据失败: %w", err)
		}

		var result struct {
			Data GetTimeSeriesDataOutput `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, GetTimeSeriesDataOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess(fmt.Sprintf("获取%s时间序列数据成功，共%d个数据点", 
			input.MetricType, len(result.Data.DataPoints)))
		return nil, result.Data, nil
	}
}

// GetSecurityMetricsInput 获取安全指标的输入参数
type GetSecurityMetricsInput struct {
	TimeRange string `json:"timeRange,omitempty" jsonschema:"时间范围：24h,7d,30d"`
}

// GetSecurityMetricsOutput 安全指标输出
type GetSecurityMetricsOutput struct {
	TotalAttacks       int                    `json:"totalAttacks" jsonschema:"总攻击次数"`
	BlockedAttacks     int                    `json:"blockedAttacks" jsonschema:"已阻止的攻击"`
	AttackTypes        map[string]int         `json:"attackTypes" jsonschema:"攻击类型分布"`
	TopAttackerIPs     []IPAttackInfo         `json:"topAttackerIPs" jsonschema:"攻击源IP TOP10"`
	ThreatLevel        string                 `json:"threatLevel" jsonschema:"威胁等级：low,medium,high,critical"`
	RecentAttacks      []interface{}          `json:"recentAttacks" jsonschema:"最近的攻击记录"`
}

// IPAttackInfo IP攻击信息
type IPAttackInfo struct {
	IP           string `json:"ip" jsonschema:"IP地址"`
	AttackCount  int    `json:"attackCount" jsonschema:"攻击次数"`
	LastAttackAt string `json:"lastAttackAt" jsonschema:"最后攻击时间"`
}

// CreateGetSecurityMetrics 创建获取安全指标的工具函数
func CreateGetSecurityMetrics(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetSecurityMetricsInput) (*mcp.CallToolResult, GetSecurityMetricsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetSecurityMetricsInput) (*mcp.CallToolResult, GetSecurityMetricsOutput, error) {
		logger := NewToolLogger("get_security_metrics")
		logger.LogInput(input)
		
		if input.TimeRange == "" {
			input.TimeRange = "24h"
		}
		
		// 使用实际的API路径 /api/v1/stats/security
		path := fmt.Sprintf("/api/v1/stats/security?timeRange=%s", input.TimeRange)
		data, err := client.Get(path)
		if err != nil {
			logger.LogError(err)
			return nil, GetSecurityMetricsOutput{}, fmt.Errorf("获取安全指标失败: %w", err)
		}

		var result struct {
			Data GetSecurityMetricsOutput `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, GetSecurityMetricsOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess(fmt.Sprintf("获取安全指标成功：总攻击%d次，已阻止%d次", 
			result.Data.TotalAttacks, result.Data.BlockedAttacks))
		return nil, result.Data, nil
	}
}

// GetSystemHealthInput 获取系统健康状态的输入参数
type GetSystemHealthInput struct{}

// GetSystemHealthOutput 系统健康状态输出
type GetSystemHealthOutput struct {
	Status     string                 `json:"status" jsonschema:"系统状态：healthy,degraded,critical"`
	Services   map[string]ServiceInfo `json:"services" jsonschema:"各服务状态"`
	CPU        float64                `json:"cpu" jsonschema:"CPU使用率(%)"`
	Memory     float64                `json:"memory" jsonschema:"内存使用率(%)"`
	DiskUsage  float64                `json:"diskUsage" jsonschema:"磁盘使用率(%)"`
	Uptime     int64                  `json:"uptime" jsonschema:"运行时间(秒)"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Status  string `json:"status" jsonschema:"服务状态：running,stopped,error"`
	Message string `json:"message,omitempty" jsonschema:"状态消息"`
}

// CreateGetSystemHealth 创建获取系统健康状态的工具函数
func CreateGetSystemHealth(client *APIClient) func(context.Context, *mcp.CallToolRequest, GetSystemHealthInput) (*mcp.CallToolResult, GetSystemHealthOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetSystemHealthInput) (*mcp.CallToolResult, GetSystemHealthOutput, error) {
		logger := NewToolLogger("get_system_health")
		
		// 使用实际的API路径 /api/v1/stats/health
		data, err := client.Get("/api/v1/stats/health")
		if err != nil {
			logger.LogError(err)
			return nil, GetSystemHealthOutput{}, fmt.Errorf("获取系统健康状态失败: %w", err)
		}

		var result struct {
			Data GetSystemHealthOutput `json:"data"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			logger.LogError(err)
			return nil, GetSystemHealthOutput{}, fmt.Errorf("解析响应失败: %w", err)
		}

		logger.LogSuccess(fmt.Sprintf("系统状态：%s，CPU：%.1f%%，内存：%.1f%%", 
			result.Data.Status, result.Data.CPU, result.Data.Memory))
		return nil, result.Data, nil
	}
}
