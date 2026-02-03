# 快速修复指南 - JWT_SECRET 和路径配置

## 问题总结
1. ✅ **已修复**: JWT_SECRET 长度不足 32 字符
2. ✅ **已修复**: geo-ip 和 HAProxy 路径配置
3. ✅ **已添加**: 配置强制更新功能

## 修复内容

### 1. JWT_SECRET 配置
**文件**: `server/.env`

已更新为安全的 99 字符密钥：
```
JWT_SECRET=aiwaf-secure-jwt-secret-key-2026-production-ready-minimum-32-chars-xK9mP2wQ7vL5nR8tY3cH6fJ4sA1dG0eB
```

### 2. 路径配置自动检测
**文件**: `server/config/init.go`

修改后的逻辑：
- **本地开发**: 自动使用项目根目录 + `/geo-ip`
- **K8s 环境**: 使用容器路径 `/home/mrya/mrya-waf/geo-ip`

```go
// 在 K8s 环境中使用容器路径，本地开发使用项目路径
var geoIPBase, haproxyBase string
if Global.IsK8s {
    geoIPBase = "/home/mrya/mrya-waf/geo-ip"
    haproxyBase = "/home/mrya/mrya-waf"
} else {
    geoIPBase = filepath.Join(projectRoot, "geo-ip")
    haproxyBase = projectRoot
}
```

### 3. 配置强制更新功能
**文件**: `server/config/init.go`

添加了环境变量 `FORCE_RESET_CONFIG` 来强制更新数据库中的配置：

```bash
FORCE_RESET_CONFIG=true go run main.go
```

## 快速启动

### 方法 1: 使用启动脚本（推荐）
```bash
cd server
bash start.sh
```

启动脚本会自动：
- ✅ 检查 JWT_SECRET 长度
- ✅ 如果不足 32 字符，自动生成新密钥
- ✅ 检查 MongoDB 连接
- ✅ 启动服务器

### 方法 2: 强制更新配置并启动
如果数据库中有旧配置，使用此方法：

```bash
cd server
FORCE_RESET_CONFIG=true go run main.go
```

### 方法 3: 直接启动
```bash
cd server
go run main.go
```

## 验证

### 1. 检查 JWT_SECRET
```bash
cd server
grep JWT_SECRET .env
# 应该看到至少 32 个字符的密钥
```

### 2. 检查服务器日志
正常启动应该看到：
```
✨ Application configure loaded successfully
启动Engine服务...
开始启动HAProxy服务...
```

### 3. 测试 API
```bash
curl http://localhost:8080/health
# 应该返回 200 OK
```

## 常见问题

### Q: 仍然看到 "JWT_SECRET must be at least 32 characters"
**A**: 
1. 检查 `.env` 文件是否存在：`ls -la server/.env`
2. 验证文件内容：`cat server/.env | grep JWT_SECRET`
3. 确保没有其他 `.env` 文件覆盖：`find . -name ".env"`

### Q: 仍然看到 "/home/mrya/ruiqi-waf/geo-ip" 路径错误
**A**: 数据库中有旧配置，运行：
```bash
cd server
FORCE_RESET_CONFIG=true go run main.go
```

### Q: MongoDB 连接失败
**A**: 启动 MongoDB：
```bash
# 使用 Docker
cd AI-Waf
docker-compose up -d mongodb

# 或本地 MongoDB
brew services start mongodb-community
```

### Q: 找不到 geo-ip 文件
**A**: 确保文件存在：
```bash
ls -la geo-ip/
# 应该看到:
# GeoLite2-ASN.mmdb
# GeoLite2-City.mmdb
```

## 生产环境部署

### 生成强密钥
```bash
# 方法 1: OpenSSL
openssl rand -base64 48

# 方法 2: Python
python3 -c "import secrets; print(secrets.token_urlsafe(64))"

# 方法 3: Go
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"; "os"); func main() { b := make([]byte, 64); rand.Read(b); fmt.Println(base64.URLEncoding.EncodeToString(b)) }'
```

### 设置环境变量
```bash
export JWT_SECRET="你的安全密钥"
export DB_URI="mongodb://user:pass@host:27017/waf"
export IS_PRODUCTION=true
```

### Docker 部署
确保在 `docker-compose.yaml` 中设置：
```yaml
services:
  ai-waf-server:
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DB_URI=${DB_URI}
      - IS_K8S=false
      - IS_PRODUCTION=true
```

## 文件清单

修改的文件：
- ✅ `server/.env` - 更新 JWT_SECRET
- ✅ `server/.env.template` - 更新模板
- ✅ `server/config/init.go` - 动态路径检测 + 强制更新功能
- ✅ `server/start.sh` - 启动脚本（新建）
- ✅ `server/reset-config.sh` - 配置重置脚本（新建）

## 下一步

服务器成功启动后：
1. 访问前端：`http://localhost:5173`
2. 测试新功能：
   - OWASP 规则模板
   - 规则有效性评分
   - 一键保护配置

详细测试指南：[web/TEST_GUIDE.md](../web/TEST_GUIDE.md)
