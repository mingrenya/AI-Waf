package service

import (
	"context"
	"strconv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mingrenya/AI-Waf/server/config"
	"github.com/rs/zerolog"
)

// LokiLogService Loki 日志查询代理服务
type LokiLogService interface {
	QueryLogs(ctx context.Context, req LokiQueryRequest) (*LokiQueryResponse, error)
	QueryRange(ctx context.Context, req LokiRangeRequest) (*LokiQueryResponse, error)
	GetLogStats(ctx context.Context, duration string) (*LokiStatsResult, error)
}

// LokiQueryRequest 即时查询请求
type LokiQueryRequest struct {
	Query     string `json:"query" binding:"required"`   // LogQL 查询语句
	Limit     int    `json:"limit"`                       // 返回条数上限，默认 100
	Start     string `json:"start"`                       // 起始时间，支持相对时间如 "1h"
	End       string `json:"end"`                         // 结束时间
	Direction string `json:"direction"`                   // "forward" 或 "backward"
}

// LokiRangeRequest 范围查询请求
type LokiRangeRequest struct {
	Query string `json:"query" binding:"required"`
	Start string `json:"start" binding:"required"`
	End   string `json:"end" binding:"required"`
	Step  string `json:"step"` // 步长，如 "15s"
	Limit int    `json:"limit"`
}

// LokiQueryResponse Loki 查询响应
type LokiQueryResponse struct {
	Status string           `json:"status"`
	Data   LokiResponseData `json:"data"`
}

// LokiResponseData 查询结果数据
type LokiResponseData struct {
	ResultType string        `json:"resultType"`
	Result     []LokiStream  `json:"result"`
}

// LokiStream 日志流
// 注意: Loki 在 matrix 模式返回的 values 是 [数字时间戳, 字符串值]
// 而 streams 模式返回 [[纳秒字符串, 日志行], ...]
// 所以使用 []interface{} 兼容两种格式
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]interface{}   `json:"values"`
}

// LokiLogEntry 格式化后的日志条目
type LokiLogEntry struct {
	Timestamp string            `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
	Message   string            `json:"message"`
	Level     string            `json:"level,omitempty"`
	Component string            `json:"component,omitempty"`
}

// LokiLogQueryResponse 前端的格式化响应
type LokiLogQueryResponse struct {
	Results     []LokiLogEntry `json:"results"`
	TotalHits   int            `json:"totalHits"`
	Query       string         `json:"query"`
	ResultType  string         `json:"resultType"`
}

type lokiLogServiceImpl struct {
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

// NewLokiLogService 创建 Loki 日志查询服务
func NewLokiLogService() LokiLogService {
	logger := config.GetServiceLogger("loki-log")

	return &lokiLogServiceImpl{
		baseURL: "http://loki:3100", // Docker Compose 内部 DNS
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// SetBaseURL 允许覆盖 Loki 地址（用于测试或非容器环境）
func (s *lokiLogServiceImpl) SetBaseURL(url string) {
	s.baseURL = url
}

// QueryLogs 执行即时 LogQL 查询
func (s *lokiLogServiceImpl) QueryLogs(ctx context.Context, req LokiQueryRequest) (*LokiQueryResponse, error) {
	params := url.Values{}
	params.Set("query", req.Query)

	if req.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", req.Limit))
	} else {
		params.Set("limit", "100")
	}

	// Loki 的 /query 仅支持 metric queries，log queries 必须走 /query_range
	// 为兼容两者，默认使用 /query_range，传入最近 1 小时范围
	start := req.Start
	end := req.End
	if start == "" {
		start = time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	} else {
		start = resolveTime(start)
	}
	if end == "" {
		end = time.Now().UTC().Format(time.RFC3339)
	} else {
		end = resolveTime(end)
	}
	params.Set("start", resolveTime(start))
	params.Set("end", resolveTime(end))

	if req.Direction != "" {
		params.Set("direction", req.Direction)
	} else {
		params.Set("direction", "backward")
	}

	// 步长设为合理的默认值
	params.Set("step", "15s")

	return s.doQuery(ctx, "/loki/api/v1/query_range", params)
}

// QueryRange 执行范围 LogQL 查询
func (s *lokiLogServiceImpl) QueryRange(ctx context.Context, req LokiRangeRequest) (*LokiQueryResponse, error) {
	params := url.Values{}
	params.Set("query", req.Query)
	params.Set("start", req.Start)
	params.Set("end", req.End)

	if req.Step != "" {
		params.Set("step", req.Step)
	} else {
		params.Set("step", "15s")
	}

	if req.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", req.Limit))
	} else {
		params.Set("limit", "100")
	}

	return s.doQuery(ctx, "/loki/api/v1/query_range", params)
}

// doQuery 执行实际的 HTTP 请求
func (s *lokiLogServiceImpl) doQuery(ctx context.Context, path string, params url.Values) (*LokiQueryResponse, error) {
	reqURL := s.baseURL + path + "?" + params.Encode()

	s.logger.Debug().Str("url", reqURL).Msg("Querying Loki")

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Loki 请求失败: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Loki 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Loki 返回错误 (status=%d): %s", resp.StatusCode, string(body))
	}

	var result LokiQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析 Loki 响应失败: %w", err)
	}

	return &result, nil
}

// ToLogEntries 将 Loki 原始响应转换为前端友好的结构化日志
func ToLogEntries(resp *LokiQueryResponse) *LokiLogQueryResponse {
	if resp == nil {
		return &LokiLogQueryResponse{
			Results:    []LokiLogEntry{},
			TotalHits:  0,
			ResultType: "streams",
		}
	}

	entries := make([]LokiLogEntry, 0)

	for _, stream := range resp.Data.Result {
		labels := stream.Stream
		for _, value := range stream.Values {
			if len(value) < 2 {
				continue
			}

			// Loki 返回纳秒时间戳（兼容数字和字符串格式）
			entry := LokiLogEntry{
				Timestamp: fmt.Sprintf("%v", value[0]),
				Labels:    labels,
				Message:   fmt.Sprintf("%v", value[1]),
			}

			// 尝试提取结构化字段
			entry.Level = extractField(fmt.Sprintf("%v", value[1]), "level")
			if cn, ok := labels["container_name"]; ok {
				entry.Component = cn
			}

			entries = append(entries, entry)
		}
	}

	return &LokiLogQueryResponse{
		Results:     entries,
		TotalHits:   len(entries),
		Query:       "",
		ResultType:  resp.Data.ResultType,
	}
}

// extractField 从 JSON 字符串中提取指定字段
func extractField(jsonStr string, field string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ""
	}
	if val, ok := data[field]; ok {
		switch v := val.(type) {
		case string:
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}


// resolveTime 将相对/绝对时间字符串转换为 Unix 秒字符串（Loki query_range 要求）
func resolveTime(t string) string {
	switch t {
	case "now", "":
		return fmt.Sprintf("%d", time.Now().Unix())
	case "1h":
		return fmt.Sprintf("%d", time.Now().Add(-1*time.Hour).Unix())
	case "6h":
		return fmt.Sprintf("%d", time.Now().Add(-6*time.Hour).Unix())
	case "24h":
		return fmt.Sprintf("%d", time.Now().Add(-24*time.Hour).Unix())
	default:
		// Try to parse as duration string (e.g., "30m", "2h", "7d")
		if d, err := time.ParseDuration(t); err == nil {
			return fmt.Sprintf("%d", time.Now().Add(-d).Unix())
		}
		// Assume already a unix timestamp
		return t
	}
}

// BuildWAFQuery 构建常见的 WAF 日志查询
func BuildWAFQuery(params ...string) string {
	// 默认查询 mrya-waf 容器的结构化 JSON 日志
	baseSelector := `{container_name="mrya-waf"}`

	conditions := []string{}
	for _, p := range params {
		conditions = append(conditions, p)
	}

	if len(conditions) == 0 {
		return baseSelector
	}

	return baseSelector + ` | json | ` + strings.Join(conditions, " | ")
}

// LokiStatsResult 统计聚合结果
type LokiStatsResult struct {
	LogVolume []TimeSeriesPoint `json:"logVolume"`
	ByLevel   map[string]int    `json:"byLevel"`
	ByComponent map[string]int  `json:"byComponent"`
	TotalHits int               `json:"totalHits"`
	RangeStart string           `json:"rangeStart"`
	RangeEnd   string           `json:"rangeEnd"`
}

// TimeSeriesPoint 时间序列点
type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp"`
	Count     float64 `json:"count"`
}

// GetLogStats 获取日志统计数据
func (s *lokiLogServiceImpl) GetLogStats(ctx context.Context, duration string) (*LokiStatsResult, error) {
	start := resolveTime(duration)
	end := resolveTime("now")
	step := "60s" // 1分钟步长

	stats := &LokiStatsResult{
		ByLevel:     make(map[string]int),
		ByComponent: make(map[string]int),
		RangeStart:  start,
		RangeEnd:    end,
	}

	// 1. 日志总量时序
	query := `sum(count_over_time({container_name="mrya-waf"}[` + step + `]))`
	resp, err := s.doQuery(ctx, "/loki/api/v1/query_range", url.Values{
		"query": {query},
		"start": {start},
		"end":   {end},
		"step":  {step},
	})
	if err != nil {
		s.logger.Warn().Err(err).Msg("Log volume query failed")
	} else {
		for _, stream := range resp.Data.Result {
			for _, v := range stream.Values {
				if len(v) >= 2 {
					ts := fmt.Sprintf("%v", v[0])
					countStr := fmt.Sprintf("%v", v[1])
					tsVal, errTs := strconv.ParseInt(ts, 10, 64)
					countVal, errCount := strconv.ParseFloat(countStr, 64)
					if errTs == nil && errCount == nil {
						stats.LogVolume = append(stats.LogVolume, TimeSeriesPoint{Timestamp: tsVal, Count: countVal})
					}
				}
			}
		}
	}

	// 2. 按级别统计
	for _, level := range []string{"error", "info", "warn", "debug"} {
		q := `sum(count_over_time({container_name="mrya-waf"} |= "` + level + `" [` + duration + `]))`
		resp, err := s.doQuery(ctx, "/loki/api/v1/query", url.Values{
			"query": {q},
			"start": {start},
			"end":   {end},
		})
		if err == nil && len(resp.Data.Result) > 0 {
			for _, stream := range resp.Data.Result {
				if len(stream.Values) > 0 && len(stream.Values[0]) >= 2 {
				countStr := fmt.Sprintf("%v", stream.Values[0][1])
				count, _ := strconv.Atoi(countStr)
					stats.ByLevel[level] += count
					stats.TotalHits += count
				}
			}
		}
	}

	// 3. 按组件统计
	components := []string{"traffic-analyzer", "baseline", "flow-controller", "pattern-detector"}
	for _, comp := range components {
		q := `sum(count_over_time({container_name="mrya-waf"} |= "` + comp + `" [` + duration + `]))`
		resp, err := s.doQuery(ctx, "/loki/api/v1/query", url.Values{
			"query": {q},
			"start": {start},
			"end":   {end},
		})
		if err == nil && len(resp.Data.Result) > 0 {
			for _, stream := range resp.Data.Result {
				if len(stream.Values) > 0 && len(stream.Values[0]) >= 2 {
					countStr := fmt.Sprintf("%v", stream.Values[0][1])
					c, _ := strconv.Atoi(countStr)
					stats.ByComponent[comp] += c
				}
			}
		}
	}

	return stats, nil
}
