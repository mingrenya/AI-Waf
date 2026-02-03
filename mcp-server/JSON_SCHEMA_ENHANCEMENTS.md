# JSON Schema 增强示例

本文档展示了为 AI-WAF MCP Server 工具添加的增强 JSON Schema 标注。

## 📋 改进内容

### 1. 添加的约束

- **minimum/maximum**: 数值范围限制
- **minLength/maxLength**: 字符串长度限制  
- **pattern**: 正则表达式验证（如IP地址格式）
- **format**: 标准格式（如 date-time, uuid, objectid）
- **enum**: 枚举值限制

### 2. 添加的示例

- **example**: 为每个字段提供实际使用示例
- 帮助 AI 更好地理解参数用法

### 3. 改进的描述

- 更详细的功能说明
- 包含单位和格式说明
- 添加使用场景和注意事项

---

## 🔧 工具示例

### 1. list_attack_logs（查询攻击日志）

#### 输入参数 JSON Schema

```json
{
  "type": "object",
  "properties": {
    "page": {
      "type": "integer",
      "description": "Page number starting from 1",
      "default": 1,
      "minimum": 1,
      "example": 1
    },
    "pageSize": {
      "type": "integer",
      "description": "Number of items per page",
      "default": 20,
      "minimum": 1,
      "maximum": 100,
      "example": 20
    },
    "ruleId": {
      "type": "integer",
      "description": "Filter by specific rule ID",
      "example": 12345
    },
    "srcIp": {
      "type": "string",
      "description": "Filter by source IP address (supports CIDR)",
      "pattern": "^(([0-9]{1,3}\\.){3}[0-9]{1,3}|([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2})$",
      "example": "192.168.1.100"
    },
    "dstIp": {
      "type": "string",
      "description": "Filter by destination IP address",
      "pattern": "^([0-9]{1,3}\\.){3}[0-9]{1,3}$",
      "example": "10.0.0.1"
    },
    "domain": {
      "type": "string",
      "description": "Filter by domain name",
      "example": "example.com"
    },
    "startTime": {
      "type": "string",
      "description": "Query start time in ISO8601 format (YYYY-MM-DDTHH:MM:SSZ)",
      "format": "date-time",
      "example": "2026-02-01T00:00:00Z"
    },
    "endTime": {
      "type": "string",
      "description": "Query end time in ISO8601 format (YYYY-MM-DDTHH:MM:SSZ)",
      "format": "date-time",
      "example": "2026-02-03T23:59:59Z"
    }
  }
}
```

#### 输出结构 JSON Schema

```json
{
  "type": "object",
  "properties": {
    "count": {
      "type": "integer",
      "description": "Total number of attack logs matching the filter criteria",
      "minimum": 0,
      "example": 137
    },
    "logs": {
      "type": "array",
      "description": "Array of attack log entries with detailed information including timestamp, source IP, attack type, and severity",
      "items": {
        "type": "object"
      }
    }
  },
  "required": ["count", "logs"]
}
```

---

### 2. create_micro_rule（创建规则）

#### 输入参数 JSON Schema

```json
{
  "type": "object",
  "required": ["name", "type", "status", "priority", "condition"],
  "properties": {
    "name": {
      "type": "string",
      "description": "Unique rule name for identification",
      "minLength": 1,
      "maxLength": 100,
      "example": "Block Malicious IP"
    },
    "type": {
      "type": "string",
      "description": "Rule type: blacklist (deny) or whitelist (allow)",
      "enum": ["blacklist", "whitelist"],
      "example": "blacklist"
    },
    "status": {
      "type": "string",
      "description": "Rule activation status",
      "enum": ["enabled", "disabled"],
      "default": "enabled",
      "example": "enabled"
    },
    "priority": {
      "type": "integer",
      "description": "Priority level, higher number means higher priority",
      "minimum": 100,
      "maximum": 1000,
      "example": 500
    },
    "condition": {
      "type": "object",
      "description": "Rule condition object containing match_type and patterns for IP, URL, or header matching",
      "example": {
        "match_type": "ip",
        "patterns": ["192.168.1.0/24", "10.0.0.0/8"]
      }
    }
  }
}
```

#### 输出结构 JSON Schema

```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "description": "Unique identifier of the newly created rule",
      "format": "uuid",
      "example": "507f1f77bcf86cd799439011"
    },
    "message": {
      "type": "string",
      "description": "Success message confirming rule creation",
      "example": "Rule created successfully"
    }
  },
  "required": ["id", "message"]
}
```

---

### 3. get_log_stats（获取日志统计）

#### 输出结构 JSON Schema

```json
{
  "type": "object",
  "properties": {
    "totalAttacks": {
      "type": "integer",
      "description": "Total number of attack attempts in the period",
      "minimum": 0,
      "example": 1337
    },
    "attackTypes": {
      "type": "object",
      "description": "Breakdown of attacks by type (sql_injection, xss, path_traversal, etc.)",
      "additionalProperties": {
        "type": "integer"
      },
      "example": {
        "sql_injection": 450,
        "xss": 230,
        "path_traversal": 120
      }
    },
    "topSourceIPs": {
      "type": "array",
      "description": "Top 10 attack source IPs with counts and percentages",
      "items": {
        "type": "object",
        "properties": {
          "ip": {"type": "string"},
          "count": {"type": "integer"},
          "percentage": {"type": "number"}
        }
      }
    },
    "severityDistribution": {
      "type": "object",
      "description": "Distribution of attacks by severity level (low, medium, high, critical)",
      "example": {
        "high": 500,
        "medium": 300,
        "low": 200
      }
    }
  }
}
```

---

## 📊 改进效果

### Before（改进前）

```go
type ListAttackLogsInput struct {
    Page     int    `json:"page,omitempty" jsonschema:"description=Page number"`
    PageSize int    `json:"pageSize,omitempty" jsonschema:"description=Items per page"`
    SrcIP    string `json:"srcIp,omitempty" jsonschema:"description=Source IP filter"`
}
```

**问题**：
- ❌ 缺少默认值
- ❌ 没有范围限制
- ❌ 缺少示例
- ❌ 描述过于简单

### After（改进后）

```go
type ListAttackLogsInput struct {
    Page     int    `json:"page,omitempty" jsonschema:"description=Page number starting from 1,default=1,minimum=1,example=1"`
    PageSize int    `json:"pageSize,omitempty" jsonschema:"description=Number of items per page,default=20,minimum=1,maximum=100,example=20"`
    SrcIP    string `json:"srcIp,omitempty" jsonschema:"description=Filter by source IP address (supports CIDR),pattern=^(([0-9]{1,3}\\.){3}[0-9]{1,3}|([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2})$,example=192.168.1.100"`
}
```

**改进**：
- ✅ 添加了默认值
- ✅ 设置了范围限制
- ✅ 提供了实际示例
- ✅ 详细描述包含格式说明
- ✅ IP 地址添加了正则验证

---

## 🎯 AI 理解提升

通过增强的 JSON Schema，AI 现在可以：

1. **自动验证参数**
   - 知道 page 必须 >= 1
   - 知道 pageSize 不能超过 100
   - 知道 IP 地址需要符合特定格式

2. **生成更准确的调用**
   - 看到示例值后能正确构造请求
   - 了解每个字段的实际用途
   - 知道哪些参数是必需的

3. **提供更好的错误提示**
   - 当参数超出范围时给出明确提示
   - 展示正确的参数格式
   - 建议合理的默认值

4. **理解数据关系**
   - 通过格式约束理解数据类型
   - 通过枚举值理解可选项
   - 通过描述理解使用场景

---

## ✅ 下一步

继续为以下工具类别增强 JSON Schema：

- [x] ✅ 日志查询工具（logs.go）
- [x] ✅ 规则管理工具（rules.go）
- [ ] ⬜ AI 分析工具（ai_analyzer.go）
- [ ] ⬜ 批量操作工具（batch_operations.go）
- [ ] ⬜ 监控工具（monitoring.go）
- [ ] ⬜ 高级规则管理（rules_advanced.go）
- [ ] ⬜ 扩展工具（extended_tools.go）

---

## 📚 参考资源

- [JSON Schema 规范](https://json-schema.org/)
- [JSON Schema 验证器](https://www.jsonschemavalidator.net/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/specification/draft)
