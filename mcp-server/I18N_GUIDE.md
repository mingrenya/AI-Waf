# MCP Server 国际化 (I18n) 实现指南

## 概述

本项目实现了一个完整的国际化系统，将面向用户的描述文本与代码结构分离：

- **结构体标签**：使用英文，确保最大兼容性（机器可解析）
- **用户描述**：通过 JSON 配置文件提供多语言支持（人类可读）

## 架构设计

### 1. 国际化模块 (`i18n/i18n.go`)

提供单例模式的国际化管理器：

```go
// 获取实例
i18nInstance := i18n.GetInstance()

// 设置语言
i18nInstance.SetLocale(i18n.LocaleZH) // 中文
i18nInstance.SetLocale(i18n.LocaleEN) // English

// 获取翻译
trans := i18n.T("tool_name")
fmt.Println(trans.Description)

// 获取字段翻译
fieldDesc := i18n.TField("tool_name", "field_name")
```

### 2. 翻译文件

#### 英文翻译 (`i18n/en.json`)
```json
{
  "list_attack_logs": {
    "name": "Query Attack Logs",
    "description": "Query WAF attack logs with filters",
    "fields": {
      "page": "Page number (default: 1)",
      "srcIp": "Filter by source IP address"
    }
  }
}
```

#### 中文翻译 (`i18n/zh.json`)
```json
{
  "list_attack_logs": {
    "name": "查询攻击日志",
    "description": "查询WAF攻击日志，支持按时间范围、攻击类型、严重程度过滤",
    "fields": {
      "page": "页码（默认：1）",
      "srcIp": "按来源IP地址过滤"
    }
  }
}
```

### 3. 结构体标签规范

✅ **正确示例** - 使用英文和标准格式：
```go
type GetLogsInput struct {
    Page     int    `json:"page" jsonschema:"description=Page number,default=1"`
    SrcIP    string `json:"srcIp" jsonschema:"description=Source IP filter"`
    RuleType string `json:"type" jsonschema:"enum=blacklist|whitelist,description=Rule type"`
}
```

❌ **错误示例** - 不要在 jsonschema 中使用中文：
```go
type GetLogsInput struct {
    Page  int    `json:"page" jsonschema:"页码,默认1"`  // ❌ 不兼容
    SrcIP string `json:"srcIp" jsonschema:"来源IP过滤"` // ❌ 解析错误
}
```

## 使用方法

### 在 main.go 中注册工具

```go
import "github.com/mingrenya/AI-Waf/mcp-server/i18n"

func main() {
    // 初始化国际化
    i18nInstance := i18n.GetInstance()
    locale := os.Getenv("WAF_LOCALE")
    if locale == "en" {
        i18nInstance.SetLocale(i18n.LocaleEN)
    } else {
        i18nInstance.SetLocale(i18n.LocaleZH)
    }

    // 注册工具时使用翻译
    registerTool := func(name string, handler interface{}) {
        trans := i18n.T(name)
        mcp.AddTool(server, &mcp.Tool{
            Name:        name,
            Description: trans.Description,
        }, handler)
    }

    // 注册工具
    registerTool("list_attack_logs", tools.CreateListAttackLogs(client))
}
```

### 环境变量配置

```bash
# 设置语言（默认：zh）
export WAF_LOCALE=zh   # 中文
export WAF_LOCALE=en   # English

# 自定义翻译文件路径（可选）
export I18N_PATH=/path/to/i18n
```

## 优势

### 1. 关注点分离
- **代码层**：jsonschema 标签保持英文，确保工具兼容性
- **展示层**：通过 JSON 文件管理多语言描述

### 2. 易于维护
- 添加新语言只需创建新的 JSON 文件
- 修改描述无需重新编译代码
- 非技术人员也能参与翻译工作

### 3. 扩展性强
```go
// 支持添加更多语言
const (
    LocaleEN Locale = "en"
    LocaleZH Locale = "zh"
    LocaleJA Locale = "ja"  // 日语
    LocaleKO Locale = "ko"  // 韩语
)
```

### 4. 向后兼容
- 如果翻译文件不存在，自动回退到英文
- 如果某个工具缺少翻译，使用工具名作为描述

## 最佳实践

### 1. jsonschema 标签规范

```go
// ✅ 使用标准关键字
`jsonschema:"description=Field description"`
`jsonschema:"enum=value1|value2,description=Enum field"`
`jsonschema:"required,description=Required field"`
`jsonschema:"minimum=0,maximum=100,description=Number range"`
`jsonschema:"default=value,description=Field with default"`

// ❌ 避免
`jsonschema:"中文描述"`  // 不要使用非 ASCII 字符
`jsonschema:"字段：值1,值2"` // 不要使用中文冒号和逗号
```

### 2. 翻译文件组织

```
mcp-server/
├── i18n/
│   ├── i18n.go      # 国际化核心代码
│   ├── en.json      # 英文翻译
│   ├── zh.json      # 中文翻译
│   └── README.md    # 翻译指南
```

### 3. 字段描述模式

```json
{
  "tool_name": {
    "name": "工具简称",
    "description": "工具的详细描述，说明功能和用途",
    "fields": {
      "field1": "字段1的描述（包含默认值和格式说明）",
      "field2": "字段2的描述"
    }
  }
}
```

## 迁移指南

如果您有现有的带中文 jsonschema 标签的代码，按以下步骤迁移：

1. **更新结构体标签**：将中文描述改为英文
   ```go
   // Before
   Page int `jsonschema:"页码,默认1"`
   
   // After  
   Page int `jsonschema:"description=Page number,default=1"`
   ```

2. **添加翻译条目**：在 `zh.json` 中添加中文描述
   ```json
   {
     "tool_name": {
       "fields": {
         "page": "页码（默认：1）"
       }
     }
   }
   ```

3. **测试验证**：
   ```bash
   # 测试中文
   WAF_LOCALE=zh go run main.go
   
   # 测试英文
   WAF_LOCALE=en go run main.go
   ```

## 示例：完整工具定义

### 代码（tools/logs.go）
```go
type GetLogsInput struct {
    Page      int    `json:"page" jsonschema:"description=Page number,default=1"`
    PageSize  int    `json:"pageSize" jsonschema:"description=Items per page,default=20,maximum=100"`
    SrcIP     string `json:"srcIp" jsonschema:"description=Source IP filter"`
    StartTime string `json:"startTime" jsonschema:"description=Start time in ISO8601 format"`
}
```

### 翻译（i18n/zh.json）
```json
{
  "list_attack_logs": {
    "name": "查询攻击日志",
    "description": "查询WAF攻击日志，支持按时间范围、IP地址、攻击类型等条件过滤",
    "fields": {
      "page": "页码（默认：1）",
      "pageSize": "每页数量（默认：20，最大：100）",
      "srcIp": "按来源IP地址过滤",
      "startTime": "查询起始时间（ISO8601格式）"
    }
  }
}
```

## 总结

通过这套国际化系统：
- ✅ 代码保持英文，确保工具兼容性
- ✅ 通过配置文件提供多语言支持
- ✅ 易于维护和扩展
- ✅ 符合国际化最佳实践

这种设计完全遵循了"结构体标签用于机器解释，使用有限字符集"的原则，同时又能为用户提供友好的本地化体验。
