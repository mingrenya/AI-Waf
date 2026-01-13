package constant

import (
	"sync"
)

// 常量存储结构，使用sync.Map确保并发安全
var (
	constants sync.Map
)

// Set 设置常量，如果常量已存在则会覆盖
// 参数:
//   - key: 常量的键
//   - value: 常量的值
func Set(key string, value interface{}) {
	constants.Store(key, value)
}

// Get 获取常量值，如果常量不存在则返回nil和false
// 参数:
//   - key: 常量的键
//
// 返回:
//   - value: 常量的值
//   - exists: 常量是否存在
func Get(key string) (value interface{}, exists bool) {
	return constants.Load(key)
}

// GetString 获取字符串类型的常量，如果常量不存在或类型不匹配则返回默认值
// 参数:
//   - key: 常量的键
//   - defaultValue: 默认值
//
// 返回:
//   - 常量的字符串值或默认值
func GetString(key string, defaultValue string) string {
	if value, exists := constants.Load(key); exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// GetInt 获取整数类型的常量，如果常量不存在或类型不匹配则返回默认值
// 参数:
//   - key: 常量的键
//   - defaultValue: 默认值
//
// 返回:
//   - 常量的整数值或默认值
func GetInt(key string, defaultValue int) int {
	if value, exists := constants.Load(key); exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return defaultValue
}

// GetBool 获取布尔类型的常量，如果常量不存在或类型不匹配则返回默认值
// 参数:
//   - key: 常量的键
//   - defaultValue: 默认值
//
// 返回:
//   - 常量的布尔值或默认值
func GetBool(key string, defaultValue bool) bool {
	if value, exists := constants.Load(key); exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

// GetFloat64 获取浮点数类型的常量，如果常量不存在或类型不匹配则返回默认值
// 参数:
//   - key: 常量的键
//   - defaultValue: 默认值
//
// 返回:
//   - 常量的浮点数值或默认值
func GetFloat64(key string, defaultValue float64) float64 {
	if value, exists := constants.Load(key); exists {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
	}
	return defaultValue
}

// Delete 删除常量
// 参数:
//   - key: 要删除的常量的键
func Delete(key string) {
	constants.Delete(key)
}

// InitSystemConstants 初始化系统默认常量
func InitSystemConstants() {
	// 在这里设置系统默认常量
	Set("APP_CONFIG_NAME", "AppConfig")
	Set("Default_ENGINE_NAME", "coraza")
}
