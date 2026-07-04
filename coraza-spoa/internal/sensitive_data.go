package internal

import (
	"regexp"
	"strings"
	"sync"
)

// SensitiveDataType 敏感数据类型
type SensitiveDataType string

const (
	SensitiveIDCard    SensitiveDataType = "id_card"    // 身份证号
	SensitivePhone     SensitiveDataType = "phone"      // 手机号
	SensitiveBankCard  SensitiveDataType = "bank_card"  // 银行卡号
	SensitiveEmail     SensitiveDataType = "email"      // 邮箱地址
	SensitiveIP        SensitiveDataType = "ip"         // IP 地址
)

// SensitiveDataRule 敏感数据检测规则
type SensitiveDataRule struct {
	Type        SensitiveDataType
	Pattern     *regexp.Regexp
	MaskFunc    func(string) string
	Description string
}

// SensitiveDataFilter 敏感数据过滤器
type SensitiveDataFilter struct {
	mu      sync.RWMutex
	rules   []SensitiveDataRule
	enabled bool
}

// 预编译的正则模式
var (
	idCardPattern   = regexp.MustCompile(`\b[1-9]\d{5}(?:19|20)\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01])\d{3}[\dXx]\b`)
	phonePattern    = regexp.MustCompile(`\b1[3-9]\d{9}\b`)
	bankCardPattern = regexp.MustCompile(`\b\d{16,19}\b`)
)

// defaultRules 默认敏感数据检测规则
var defaultRules = []SensitiveDataRule{
	{
		Type:        SensitiveIDCard,
		Pattern:     idCardPattern,
		Description: "身份证号码",
		MaskFunc:    maskIDCard,
	},
	{
		Type:        SensitivePhone,
		Pattern:     phonePattern,
		Description: "手机号码",
		MaskFunc:    maskPhone,
	},
	{
		Type:        SensitiveBankCard,
		Pattern:     bankCardPattern,
		Description: "银行卡号",
		MaskFunc:    maskBankCard,
	},
}

// NewSensitiveDataFilter 创建敏感数据过滤器
func NewSensitiveDataFilter() *SensitiveDataFilter {
	return &SensitiveDataFilter{
		rules:   defaultRules,
		enabled: false, // 默认关闭，需要显式启用
	}
}

// Enable 启用过滤器
func (f *SensitiveDataFilter) Enable() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enabled = true
}

// Disable 禁用过滤器
func (f *SensitiveDataFilter) Disable() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enabled = false
}

// IsEnabled 检查是否启用
func (f *SensitiveDataFilter) IsEnabled() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.enabled
}

// AddRule 添加自定义规则
func (f *SensitiveDataFilter) AddRule(rule SensitiveDataRule) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rules = append(f.rules, rule)
}

// SetRules 替换所有规则
func (f *SensitiveDataFilter) SetRules(rules []SensitiveDataRule) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rules = rules
}

// Filter 过滤敏感数据，返回过滤后的内容和发现的敏感信息
// 返回: (过滤后内容, 发现的敏感信息列表)
func (f *SensitiveDataFilter) Filter(body []byte) ([]byte, []SensitiveMatch) {
	f.mu.RLock()
	enabled := f.enabled
	rules := make([]SensitiveDataRule, len(f.rules))
	copy(rules, f.rules)
	f.mu.RUnlock()

	if !enabled || len(body) == 0 {
		return body, nil
	}

	content := string(body)
	var matches []SensitiveMatch

	for _, rule := range rules {
		locs := rule.Pattern.FindAllStringIndex(content, -1)
		for _, loc := range locs {
			original := content[loc[0]:loc[1]]
			masked := rule.MaskFunc(original)
			matches = append(matches, SensitiveMatch{
				Type:        rule.Type,
				Description: rule.Description,
				Position:    loc[0],
				Original:    "",
				Masked:      masked,
			})
		}
		content = rule.Pattern.ReplaceAllStringFunc(content, rule.MaskFunc)
	}

	return []byte(content), matches
}

// SensitiveMatch 敏感数据匹配结果
type SensitiveMatch struct {
	Type        SensitiveDataType `json:"type"`
	Description string            `json:"description"`
	Position    int               `json:"position"`
	Original    string            `json:"-"` // 不序列化原始值
	Masked      string            `json:"masked"`
}

// 脱敏函数：保留首尾，中间用 * 替换
func maskIDCard(s string) string {
	if len(s) < 8 {
		return strings.Repeat("*", len(s))
	}
	// 保留前 2 位和后 4 位
	keepHead, keepTail := 2, 4
	return s[:keepHead] + strings.Repeat("*", len(s)-keepHead-keepTail) + s[len(s)-keepTail:]
}

func maskPhone(s string) string {
	if len(s) < 7 {
		return strings.Repeat("*", len(s))
	}
	// 保留前 3 位和后 4 位
	keepHead, keepTail := 3, 4
	return s[:keepHead] + strings.Repeat("*", len(s)-keepHead-keepTail) + s[len(s)-keepTail:]
}

func maskBankCard(s string) string {
	if len(s) < 8 {
		return strings.Repeat("*", len(s))
	}
	// 保留前 4 位和后 4 位
	keepHead, keepTail := 4, 4
	return s[:keepHead] + strings.Repeat("*", len(s)-keepHead-keepTail) + s[len(s)-keepTail:]
}

// hasSensitiveData 检查 body 中是否包含敏感数据（不执行脱敏，仅快速检测）
func (f *SensitiveDataFilter) HasSensitiveData(body []byte) bool {
	f.mu.RLock()
	enabled := f.enabled
	rules := make([]SensitiveDataRule, len(f.rules))
	copy(rules, f.rules)
	f.mu.RUnlock()

	if !enabled || len(body) == 0 {
		return false
	}

	content := string(body)
	for _, rule := range rules {
		if rule.Pattern.MatchString(content) {
			return true
		}
	}
	return false
}
