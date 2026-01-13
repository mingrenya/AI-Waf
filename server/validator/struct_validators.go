package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 结构体类型到验证函数的映射
var structValidators = map[interface{}]validator.StructLevelFunc{}

// 注册结构级验证器
func RegisterStructValidator(structType interface{}, fn validator.StructLevelFunc) {
	// 将验证器保存到映射中
	structValidators[structType] = fn
}

// 初始化所有结构级验证器
func InitStructValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 从映射中注册所有结构验证器
		for structType, fn := range structValidators {
			v.RegisterStructValidation(fn, structType)
		}
	}
}
