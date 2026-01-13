// validators/validators.go
package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 全局验证器映射
var customValidators = map[string]validator.Func{}

// 注册一个自定义验证器
func Register(tag string, fn validator.Func) {
	customValidators[tag] = fn
}

// 初始化所有自定义验证器
func InitValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册所有验证器
		for tag, fn := range customValidators {
			_ = v.RegisterValidation(tag, fn)
		}
	}
}
