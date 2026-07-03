# 环境配置说明

## 配置文件位置

所有环境变量统一配置在根目录的 `.env` 文件中。

## 配置DeepSeek AI助手

1. **获取API Key**
   - 访问 https://platform.deepseek.com/
   - 注册账号并创建API Key

2. **编辑.env文件**
   ```bash
   # 编辑配置文件
   nano .env
   
   # 找到以下配置并填写实际的API Key
   DEEPSEEK_API_KEY=sk-your-actual-api-key-here
   ```

3. **重启服务**
   ```bash
   docker compose down
   docker compose up -d
   ```

4. **验证配置**
   - 打开浏览器访问 http://localhost:2333
   - 点击右上角"AI 助手"按钮
   - 输入问题测试

## 环境变量说明

### 必填配置
- `JWT_SECRET`: JWT密钥（至少32字符，建议64字符以上随机串）
- `MONGO_ROOT_PASSWORD`: MongoDB root 密码
- `WAF_USERNAME`: MCP 服务账号（建议最小权限）
- `WAF_PASSWORD`: MCP 服务账号密码

### AI助手配置（可选）
- `DEEPSEEK_API_KEY`: DeepSeek API密钥（留空则AI助手不可用）
- `DEEPSEEK_BASE_URL`: API地址（默认：https://api.deepseek.com）
- `DEEPSEEK_MODEL`: 模型选择（推荐：deepseek-chat）

### 数据库配置
- `MONGO_ROOT_PASSWORD`: MongoDB密码（建议修改）
- `MONGO_INITDB_DATABASE`: 数据库名称

### 其他配置
- `IS_PRODUCTION`: 生产模式（生产环境必须为 `true`）
- `JWT_EXPIRATION_HRS`: JWT有效期（生产建议：12）
- `CORS_ALLOWED_ORIGINS`: 允许跨域源（多个域名用`,`分隔）
- `VITE_API_BASE_URL`: 前端 API 地址（生产建议 HTTPS 域名）
- `LOG_LEVEL`: 日志级别（生产建议：warn）
- `LOG_FORMAT`: 日志格式（生产建议：json）

## 生产环境建议配置

```env
IS_PRODUCTION=true
JWT_EXPIRATION_HRS=12
LOG_LEVEL=warn
LOG_FORMAT=json
CORS_ALLOWED_ORIGINS=https://waf.example.com,https://admin.waf.example.com
VITE_API_BASE_URL=https://waf.example.com/api/v1
```

## 注意事项

1. **.env文件已在.gitignore中**，不会被提交到版本控制
2. **修改配置后需要重启服务**才能生效
3. **前后端共用同一个.env文件**，确保配置一致
4. **DeepSeek API Key留空时**，AI助手功能不可用，其他功能正常

## 故障排查

### AI助手不工作
```bash
# 检查环境变量是否生效
docker compose exec mrya env | grep DEEPSEEK

# 查看服务日志
docker compose logs mrya | grep deepseek
```

### 配置未生效
```bash
# 完全重启服务
docker compose down
docker compose up -d

# 确认.env文件格式正确（无多余空格）
cat .env | grep DEEPSEEK

# 检查生产关键配置
cat .env | grep -E "IS_PRODUCTION|JWT_EXPIRATION_HRS|CORS_ALLOWED_ORIGINS|LOG_LEVEL|LOG_FORMAT"
```

## 参考文档

- [AI助手完整指南](./AI_ASSISTANT_GUIDE.md)
- [修复总结](./FIX_SUMMARY.md)
- [DeepSeek官方文档](https://api-docs.deepseek.com/zh-cn/)
