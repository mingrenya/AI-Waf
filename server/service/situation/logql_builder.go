package situation

import (
	"fmt"
	"strings"
)

// LogQLBuilder 动态构造 LogQL 查询
type LogQLBuilder struct {
	selector string
	filters  []string
}

// NewLogQLBuilder 创建新的 LogQL 构造器
func NewLogQLBuilder() *LogQLBuilder {
	return &LogQLBuilder{
		selector: `{container_name="mrya-waf"}`,
		filters:  make([]string, 0),
	}
}

// Filter 添加标签过滤器
func (b *LogQLBuilder) Filter(key, operator, value string) *LogQLBuilder {
	b.filters = append(b.filters, fmt.Sprintf(`%s%s"%s"`, key, operator, value))
	return b
}

// FilterIP 按源 IP 过滤
func (b *LogQLBuilder) FilterIP(ip string) *LogQLBuilder {
	return b.Filter("source_ip", `=`, ip)
}

// FilterAttackType 按攻击类型过滤
func (b *LogQLBuilder) FilterAttackType(attackType string) *LogQLBuilder {
	return b.Filter("attack_type", `=`, attackType)
}

// FilterSeverity 按严重级别过滤（支持 regex）
func (b *LogQLBuilder) FilterSeverity(severity string) *LogQLBuilder {
	return b.Filter("severity", `=~`, severity)
}

// CountOverTime 按时间窗口计数聚合
func (b *LogQLBuilder) CountOverTime(duration string) string {
	sel := b.buildSelector()
	return fmt.Sprintf(`sum by (source_ip) (count_over_time(%s[%s]))`, sel, duration)
}

// RawQuery 返回原始 LogQL log stream 选择器
func (b *LogQLBuilder) RawQuery() string {
	return b.buildSelector()
}

// AttackChainQuery 构建单个 IP 的攻击链查询
func (b *LogQLBuilder) AttackChainQuery(ip string) string {
	return fmt.Sprintf(
		`{container_name="mrya-waf",source_ip="%s"} | json | line_format "{{.attack_type}} {{.severity}} {{.action}} {{.waf_phase}} {{.site_id}}"`,
		ip,
	)
}

func (b *LogQLBuilder) buildSelector() string {
	filters := make([]string, 0, len(b.filters)+1)
	filters = append(filters, `container_name="mrya-waf"`)
	filters = append(filters, b.filters...)
	return "{" + strings.Join(filters, ",") + "}"
}
