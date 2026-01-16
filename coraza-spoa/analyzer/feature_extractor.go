package analyzer

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mingrenya/AI-Waf/pkg/model"
)

// AttackFeature 攻击特征
type AttackFeature struct {
	// 基础信息
	Timestamp   time.Time `json:"timestamp" bson:"timestamp"`
	RequestID   string    `json:"requestId" bson:"requestId"`
	SrcIP       string    `json:"srcIp" bson:"srcIp"`
	DstIP       string    `json:"dstIp" bson:"dstIp"`
	URI         string    `json:"uri" bson:"uri"`
	Domain      string    `json:"domain" bson:"domain"`
	
	// 规则特征
	RuleID      int       `json:"ruleId" bson:"ruleId"`
	Severity    int       `json:"severity" bson:"severity"`
	Payload     string    `json:"payload" bson:"payload"`
	SecMark     string    `json:"secMark" bson:"secMark"`
	
	// 提取的特征
	URLPattern  string    `json:"urlPattern" bson:"urlPattern"`      // URL模式
	PathPattern string    `json:"pathPattern" bson:"pathPattern"`    // 路径模式
	IPPattern   string    `json:"ipPattern" bson:"ipPattern"`        // IP模式(CIDR)
	PayloadType string    `json:"payloadType" bson:"payloadType"`    // 载荷类型
	
	// 统计特征
	RequestCount   int       `json:"requestCount" bson:"requestCount"`     // 请求次数
	TimeWindowSec  int       `json:"timeWindowSec" bson:"timeWindowSec"`   // 时间窗口(秒)
	Frequency      float64   `json:"frequency" bson:"frequency"`           // 频率(次/秒)
	Burstiness     float64   `json:"burstiness" bson:"burstiness"`         // 突发性指标
	
	// 内容特征
	ContainsSQLi       bool `json:"containsSqli" bson:"containsSqli"`
	ContainsXSS        bool `json:"containsXss" bson:"containsXss"`
	ContainsPathTraversal bool `json:"containsPathTraversal" bson:"containsPathTraversal"`
	ContainsCommandInjection bool `json:"containsCommandInjection" bson:"containsCommandInjection"`
	
	// 元特征
	Confidence float64 `json:"confidence" bson:"confidence"` // 置信度
}

// FeatureExtractor 特征提取器
type FeatureExtractor struct {
	// 预编译的正则表达式
	sqlInjectionRegex       *regexp.Regexp
	xssRegex                *regexp.Regexp
	pathTraversalRegex      *regexp.Regexp
	commandInjectionRegex   *regexp.Regexp
}

// NewFeatureExtractor 创建特征提取器
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{
		sqlInjectionRegex:       regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script|javascript|onerror)`),
		xssRegex:                regexp.MustCompile(`(?i)(<script|<iframe|javascript:|onerror=|onload=|<img[^>]+src)`),
		pathTraversalRegex:      regexp.MustCompile(`(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c)`),
		commandInjectionRegex:   regexp.MustCompile(`(?i)(;\s*(ls|cat|wget|curl|bash|sh|cmd|powershell)|\||&&|\$\(|\x60)`),
	}
}

// ExtractFeatures 从WAF日志中提取特征
func (fe *FeatureExtractor) ExtractFeatures(log *model.WAFLog) (*AttackFeature, error) {
	if log == nil {
		return nil, fmt.Errorf("日志为空")
	}
	
	feature := &AttackFeature{
		Timestamp:   log.CreatedAt,
		RequestID:   log.RequestID,
		SrcIP:       log.SrcIP,
		DstIP:       log.DstIP,
		URI:         log.URI,
		Domain:      log.Domain,
		RuleID:      log.RuleID,
		Severity:    log.Severity,
		Payload:     log.Payload,
		SecMark:     log.SecMark,
		RequestCount: 1,
	}
	
	// 提取URL模式
	feature.URLPattern = fe.extractURLPattern(log.URI)
	feature.PathPattern = fe.extractPathPattern(log.URI)
	feature.IPPattern = fe.extractIPPattern(log.SrcIP)
	feature.PayloadType = fe.classifyPayload(log.Payload)
	
	// 检测内容特征
	payload := strings.ToLower(log.Payload + " " + log.URI)
	feature.ContainsSQLi = fe.sqlInjectionRegex.MatchString(payload)
	feature.ContainsXSS = fe.xssRegex.MatchString(payload)
	feature.ContainsPathTraversal = fe.pathTraversalRegex.MatchString(payload)
	feature.ContainsCommandInjection = fe.commandInjectionRegex.MatchString(payload)
	
	// 计算基础置信度
	feature.Confidence = fe.calculateConfidence(feature)
	
	return feature, nil
}

// extractURLPattern 提取URL模式
func (fe *FeatureExtractor) extractURLPattern(uri string) string {
	if uri == "" {
		return ""
	}
	
	// 解析URL
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	
	// 提取路径并泛化参数
	path := parsedURL.Path
	if parsedURL.RawQuery != "" {
		path += "?*" // 泛化查询参数
	}
	
	return path
}

// extractPathPattern 提取路径模式
func (fe *FeatureExtractor) extractPathPattern(uri string) string {
	if uri == "" {
		return ""
	}
	
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	
	path := parsedURL.Path
	
	// 泛化路径中的数字ID
	digitRegex := regexp.MustCompile(`/\d+`)
	path = digitRegex.ReplaceAllString(path, "/{id}")
	
	// 泛化UUID
	uuidRegex := regexp.MustCompile(`/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	path = uuidRegex.ReplaceAllString(path, "/{uuid}")
	
	return path
}

// extractIPPattern 提取IP模式(CIDR)
func (fe *FeatureExtractor) extractIPPattern(ip string) string {
	if ip == "" {
		return ""
	}
	
	// 提取前3段作为C段网络
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.%s.0/24", parts[0], parts[1], parts[2])
	}
	
	return ip
}

// classifyPayload 分类载荷类型
func (fe *FeatureExtractor) classifyPayload(payload string) string {
	payload = strings.ToLower(payload)
	
	if fe.sqlInjectionRegex.MatchString(payload) {
		return "sql_injection"
	}
	if fe.xssRegex.MatchString(payload) {
		return "xss"
	}
	if fe.pathTraversalRegex.MatchString(payload) {
		return "path_traversal"
	}
	if fe.commandInjectionRegex.MatchString(payload) {
		return "command_injection"
	}
	
	return "unknown"
}

// calculateConfidence 计算置信度
func (fe *FeatureExtractor) calculateConfidence(feature *AttackFeature) float64 {
	confidence := 0.0
	
	// 基于严重性
	confidence += float64(feature.Severity) * 0.1
	
	// 基于检测到的攻击类型
	if feature.ContainsSQLi || feature.ContainsXSS {
		confidence += 0.3
	}
	if feature.ContainsPathTraversal || feature.ContainsCommandInjection {
		confidence += 0.2
	}
	
	// 基于载荷类型
	if feature.PayloadType != "unknown" {
		confidence += 0.2
	}
	
	// 归一化到0-1
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// AggregateFeatures 聚合相似特征
func (fe *FeatureExtractor) AggregateFeatures(features []*AttackFeature) []*AttackFeature {
	if len(features) == 0 {
		return features
	}
	
	// 按模式分组
	groups := make(map[string][]*AttackFeature)
	for _, f := range features {
		key := fmt.Sprintf("%s|%s|%s", f.URLPattern, f.IPPattern, f.PayloadType)
		groups[key] = append(groups[key], f)
	}
	
	// 聚合每组
	result := make([]*AttackFeature, 0, len(groups))
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		
		// 使用第一个作为模板
		aggregated := group[0]
		aggregated.RequestCount = len(group)
		
		// 计算时间窗口
		minTime, maxTime := group[0].Timestamp, group[0].Timestamp
		for _, f := range group {
			if f.Timestamp.Before(minTime) {
				minTime = f.Timestamp
			}
			if f.Timestamp.After(maxTime) {
				maxTime = f.Timestamp
			}
		}
		
		timeWindow := maxTime.Sub(minTime).Seconds()
		if timeWindow == 0 {
			timeWindow = 1
		}
		aggregated.TimeWindowSec = int(timeWindow)
		aggregated.Frequency = float64(len(group)) / timeWindow
		
		// 计算突发性 (标准差/均值)
		if len(group) > 1 {
			aggregated.Burstiness = fe.calculateBurstiness(group)
		}
		
		result = append(result, aggregated)
	}
	
	return result
}

// calculateBurstiness 计算突发性指标
func (fe *FeatureExtractor) calculateBurstiness(features []*AttackFeature) float64 {
	if len(features) < 2 {
		return 0.0
	}
	
	// 计算时间间隔
	intervals := make([]float64, len(features)-1)
	for i := 1; i < len(features); i++ {
		intervals[i-1] = features[i].Timestamp.Sub(features[i-1].Timestamp).Seconds()
	}
	
	// 计算平均间隔
	sum := 0.0
	for _, interval := range intervals {
		sum += interval
	}
	mean := sum / float64(len(intervals))
	
	if mean == 0 {
		return 0.0
	}
	
	// 计算标准差
	varSum := 0.0
	for _, interval := range intervals {
		diff := interval - mean
		varSum += diff * diff
	}
	stdDev := 0.0
	if len(intervals) > 1 {
		stdDev = varSum / float64(len(intervals)-1)
	}
	
	// 突发性 = 标准差 / 平均值
	return stdDev / mean
}

