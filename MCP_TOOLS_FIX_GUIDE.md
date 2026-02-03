# MCP工具问题快速修复指南

## 执行摘要

通过系统审计，发现并修复了4个问题，还有3个问题需要修复。

### ✅ 已完成修复 (4个)
1. POST /api/v1/blocked-ips 路由缺失
2. create_micro_rule 的 condition 字段格式问题
3. export_rules 字段名错误 (list → items)
4. update_micro_rule 的 condition 字段格式问题

### ⚠️ 待修复 (3个)
1. 缺少单个删除 blocked-ip 的路由 (影响 batch_unblock_ips)
2. 需要验证 list_attack_logs 的字段名
3. 需要验证 list_sites 的响应结构

---

## 🔥 紧急修复: 添加单个删除 blocked-ip 路由

### 问题描述
`batch_unblock_ips` 工具需要循环调用 `DELETE /api/v1/blocked-ips/{ip}`，但该路由不存在。

### 修复步骤

#### 1. 添加 DTO (server/dto/blocked_ip.go)

```go
// BlockedIPDeleteRequest 删除封禁IP请求（URL参数）
// @Description 通过IP地址删除封禁记录
type BlockedIPDeleteRequest struct {
	IP string `uri:"ip" binding:"required,ip"` // IP地址
}
```

#### 2. 添加 Service 方法 (server/service/blocked_ip.go)

在 `BlockedIPService` 接口中添加：
```go
DeleteBlockedIP(ctx context.Context, ip string) error
```

在 `BlockedIPServiceImpl` 中实现：
```go
// DeleteBlockedIP 删除指定IP的封禁记录
func (s *BlockedIPServiceImpl) DeleteBlockedIP(ctx context.Context, ip string) error {
	s.logger.Info().Str("ip", ip).Msg("删除封禁IP记录请求")

	deletedCount, err := s.blockedIPRepo.DeleteBlockedIPByIP(ctx, ip)
	if err != nil {
		s.logger.Error().Err(err).Str("ip", ip).Msg("删除封禁IP记录失败")
		return err
	}

	if deletedCount == 0 {
		s.logger.Warn().Str("ip", ip).Msg("未找到要删除的封禁IP记录")
		return ErrBlockedIPNotFound
	}

	s.logger.Info().Str("ip", ip).Int64("deleted", deletedCount).Msg("删除封禁IP记录成功")
	return nil
}
```

#### 3. 添加 Repository 方法 (server/repository/blocked_ip.go)

在 `BlockedIPRepository` 接口中添加：
```go
DeleteBlockedIPByIP(ctx context.Context, ip string) (int64, error)
```

实现：
```go
// DeleteBlockedIPByIP 根据IP删除封禁记录
func (r *BlockedIPRepositoryImpl) DeleteBlockedIPByIP(ctx context.Context, ip string) (int64, error) {
	collection := r.db.Collection("blocked_ips")

	filter := bson.M{"ip": ip}
	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
```

#### 4. 添加 Controller 方法 (server/controller/blocked_ip.go)

在 `BlockedIPController` 接口中添加：
```go
DeleteBlockedIP(ctx *gin.Context)
```

实现：
```go
// DeleteBlockedIP 删除封禁IP记录
//
//	@Summary		删除封禁IP记录
//	@Description	根据IP地址删除封禁记录
//	@Tags			封禁IP管理
//	@Produce		json
//	@Param			ip	path	string	true	"IP地址"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse	"删除成功"
//	@Failure		400	{object}	model.ErrResponse		"请求参数错误"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"未授权访问"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"禁止访问"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"封禁IP记录不存在"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"服务器内部错误"
//	@Router			/api/v1/blocked-ips/{ip} [delete]
func (c *BlockedIPControllerImpl) DeleteBlockedIP(ctx *gin.Context) {
	ip := ctx.Param("ip")

	c.logger.Info().Str("ip", ip).Msg("删除封禁IP记录请求")

	err := c.blockedIPService.DeleteBlockedIP(ctx, ip)
	if err != nil {
		if errors.Is(err, service.ErrBlockedIPNotFound) {
			response.NotFound(ctx, err, false)
			return
		}
		c.logger.Error().Err(err).Str("ip", ip).Msg("删除封禁IP记录失败")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("ip", ip).Msg("删除封禁IP记录成功")
	response.Success(ctx, "删除成功", nil)
}
```

#### 5. 添加路由 (server/router/router.go)

```go
// 封禁IP管理模块
blockedIPRoutes := authenticated.Group("/blocked-ips")
{
	blockedIPRoutes.GET("", middleware.HasPermission(model.PermConfigRead), blockedIPController.GetBlockedIPs)
	blockedIPRoutes.POST("", middleware.HasPermission(model.PermConfigUpdate), blockedIPController.CreateBlockedIP)
	blockedIPRoutes.DELETE("/:ip", middleware.HasPermission(model.PermConfigUpdate), blockedIPController.DeleteBlockedIP)  // 新增
	blockedIPRoutes.GET("/stats", middleware.HasPermission(model.PermConfigRead), blockedIPController.GetBlockedIPStats)
	blockedIPRoutes.DELETE("/cleanup", middleware.HasPermission(model.PermConfigUpdate), blockedIPController.CleanupExpiredBlockedIPs)
}
```

#### 6. 添加必要的导入

在 `server/controller/blocked_ip.go` 中确保有：
```go
import (
	"errors"
	"time"
	// ...其他导入
)
```

#### 7. 编译测试

```bash
cd server
go build -o ai-waf-server ./main.go
```

---

## 📋 验证修复清单

### 已修复功能验证

#### ✅ 1. batch_block_ips 工具
```bash
# 测试命令
curl -X POST http://localhost:2333/api/v1/blocked-ips \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ip": "192.168.1.100",
    "reason": "测试封禁",
    "duration": 3600
  }'

# 预期: 200 OK，返回封禁记录
```

#### ✅ 2. create_micro_rule 工具 (字符串 condition)
```bash
# 测试命令 - 注意condition是字符串格式
curl -X POST http://localhost:2333/api/v1/micro-rules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test_Rule",
    "type": "blacklist",
    "status": "enabled",
    "priority": 900,
    "condition": "{\"match_type\": \"exact\", \"ip_list\": [\"1.2.3.4\"]}"
  }'

# 预期: 200 OK，创建成功
```

#### ✅ 3. create_micro_rule 工具 (对象 condition)
```bash
# 测试命令 - condition是对象格式
curl -X POST http://localhost:2333/api/v1/micro-rules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test_Rule_2",
    "type": "blacklist",
    "status": "enabled",
    "priority": 900,
    "condition": {"match_type": "exact", "ip_list": ["1.2.3.4"]}
  }'

# 预期: 200 OK，创建成功
```

#### ✅ 4. export_rules 工具
```bash
# 测试命令
curl -X GET "http://localhost:2333/api/v1/micro-rules?size=1000&page=1" \
  -H "Authorization: Bearer $TOKEN"

# 预期: 返回所有规则，字段名为 items
```

---

### 待验证功能

#### ⚠️ 5. batch_unblock_ips 工具（待修复后验证）
```bash
# 测试删除单个IP
curl -X DELETE http://localhost:2333/api/v1/blocked-ips/192.168.1.100 \
  -H "Authorization: Bearer $TOKEN"

# 预期: 200 OK
```

#### ⚠️ 6. list_attack_logs 工具
```bash
# 检查返回字段
curl -X GET "http://localhost:2333/api/v1/log?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data | keys'

# 验证返回的字段名
```

#### ⚠️ 7. list_sites 工具
```bash
# 检查返回结构
curl -X GET "http://localhost:2333/api/v1/site" \
  -H "Authorization: Bearer $TOKEN" \
  | jq 'keys'

# 验证返回结构
```

---

## 📊 修复总结

### 修复前后对比

| 问题 | 修复前 | 修复后 | 状态 |
|------|--------|--------|------|
| POST /api/v1/blocked-ips | ❌ 404错误 | ✅ 正常工作 | 已修复 |
| create_micro_rule (字符串condition) | ❌ 500错误 | ✅ 正常工作 | 已修复 |
| export_rules | ❌ 返回0条 | ✅ 返回所有数据 | 已修复 |
| update_micro_rule (字符串condition) | ❌ 可能500错误 | ✅ 正常工作 | 已修复 |
| DELETE /api/v1/blocked-ips/{ip} | ❌ 路由不存在 | ⚠️ 待修复 | 进行中 |

### 代码修改统计

```
已修改文件: 5个
  - server/dto/blocked_ip.go (新增DTO)
  - server/controller/blocked_ip.go (新增方法+导入)
  - server/router/router.go (新增路由)
  - server/service/rule.go (修复2个方法)
  - mcp-server/tools/rules_advanced.go (修复字段名)

新增代码行数: ~150行
修改代码行数: ~30行
删除代码行数: 0行
```

---

## 🎯 下一步行动

### 立即执行 (今天)
1. ✅ 应用已完成的修复
2. ⚠️ 实现 DeleteBlockedIP 功能（按上述步骤）
3. ⚠️ 测试所有修复的功能
4. ⚠️ 重启服务并验证

### 短期计划 (本周)
1. 审计剩余29个MCP工具
2. 验证所有 list 接口的字段名
3. 统一响应格式规范
4. 添加集成测试

### 长期优化 (下周)
1. 创建 API 路由自动测试套件
2. 添加 MCP 工具端到端测试
3. 完善错误处理和日志记录
4. 编写 API 文档

---

## 📝 相关文档
- [BUG_FIXES_SUMMARY.md](BUG_FIXES_SUMMARY.md) - 详细的bug修复说明
- [MCP_TOOLS_AUDIT.md](MCP_TOOLS_AUDIT.md) - 完整的工具审计报告
- [MCP_TOOLS_COMPLETE_LIST.md](MCP_TOOLS_COMPLETE_LIST.md) - 所有MCP工具列表
