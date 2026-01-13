// validators/string_validators.go
package validator

import (
	"net"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// 初始化字符串相关验证器
func init() {
	Register("domain", DomainOrIPValidator)
}

// DomainOrIPValidator 验证字符串是否为有效的域名或IP地址
var DomainOrIPValidator validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 1. 检查是否为有效的IP地址
	if ip := net.ParseIP(value); ip != nil {
		return true
	}

	// 2. 检查是否为有效的域名
	// 域名正则表达式
	// 规则:
	// 1. 由字母、数字、连字符组成，连字符不能在开头或结尾
	// 2. 每个标签(点之间的部分)长度不超过63个字符
	// 3. 至少有一个点，最后一个部分至少2个字符(TLD)
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

	return domainRegex.MatchString(value)
}
