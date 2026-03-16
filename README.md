# MRYa WAF (AI-Waf)

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.24.1-00ADD8?style=flat&logo=go)](https://go.dev/)
[![HAProxy](https://img.shields.io/badge/HAProxy-3.0-green?style=flat&logo=haproxy)](https://www.haproxy.org/)
[![Coraza](https://img.shields.io/badge/OWASP-Coraza-blue?style=flat)](https://github.com/corazawaf/coraza)
[![MongoDB](https://img.shields.io/badge/MongoDB-6.0-47A248?style=flat&logo=mongodb)](https://www.mongodb.com/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://react.dev/)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=flat)](./LICENSE)

</div>

AI-Waf 是一个基于 HAProxy、OWASP Coraza、Go 和 React 的 Web 应用防火墙平台，覆盖请求接入、防护决策、日志分析、告警联动和 AI 辅助运维能力。

## 功能概览

- 多引擎防护：Coraza WAF + MicroEngine 微规则引擎
- 统一管理：站点、证书、规则、IP 封禁、系统配置
- 监控告警：攻击日志、统计看板、告警规则与通道
- AI 能力：DeepSeek 对话分析与规则建议
- MCP 集成：支持 MCP 工具调用平台能力

## 架构说明

核心请求链路如下：

1. 流量进入 HAProxy
2. HAProxy 通过 SPOE 调用 coraza-spoa
3. Coraza 执行请求/响应安全检测
4. MicroEngine 执行业务化规则判断
5. server 提供管理 API 与数据服务
6. web 提供可视化控制台
7. mcp-server 提供 AI 工具调用入口

系统设计细节可参考 SYSTEM_OVERVIEW.md。

## 目录结构

```text
coraza-spoa/   SPOE Agent 与规则执行核心
server/        后端管理 API 与业务逻辑
web/           前端控制台（React + Vite）
mcp-server/    MCP Server
pkg/           共享基础库
doc/ docs/     设计与功能文档
```

## 快速开始

### 方式一：Docker Compose（推荐）

1. 在项目根目录准备 `.env`（可参考 `.env.template`）
2. 启动服务

```bash
docker compose up -d --build
```

3. 默认访问地址

- 控制台：http://localhost:2333
- 健康检查：http://localhost:2333/health
- Swagger：http://localhost:2333/swagger/index.html
- ReDoc：http://localhost:2333/redoc

### 方式二：本地开发

依赖要求：

- Go 1.24.1+
- Node.js 23.10.0+
- pnpm 10.11.0+
- MongoDB 6.0+

启动示例：

```bash
# 后端
cd server
go run main.go

# 前端
cd web
pnpm install
pnpm dev
```

## 关键环境变量

在根目录 `.env` 中配置：

- `MONGO_ROOT_PASSWORD`：MongoDB root 密码
- `MONGO_INITDB_DATABASE`：数据库名
- `JWT_SECRET`：JWT 密钥，建议 64 字符以上随机串
- `JWT_EXPIRATION_HRS`：JWT 过期时间（小时）
- `IS_PRODUCTION`：是否生产模式（`true`/`false`）
- `CORS_ALLOWED_ORIGINS`：允许跨域来源，多个用英文逗号分隔
- `VITE_API_BASE_URL`：前端 API 地址
- `WAF_USERNAME`：MCP 服务账号
- `WAF_PASSWORD`：MCP 服务账号密码
- `DEEPSEEK_API_KEY`：DeepSeek API Key（可选）

生产环境建议：

- 启用 HTTPS 域名
- 立即更换默认管理员与服务账号密码
- 将日志级别调整为 `warn` 或更高

详细配置说明见 ENV_CONFIG_GUIDE.md。

## 文档导航

- SYSTEM_OVERVIEW.md：系统总览
- ENV_CONFIG_GUIDE.md：环境变量说明
- MCP_SETUP_GUIDE.md：MCP 部署与排查
- QUICK_REFERENCE.md：MCP 工具参数速查
- AI_ASSISTANT_GUIDE.md：AI 助手集成说明
- SECURITY_AUDIT_REPORT.md：安全审计记录

## 质量检查

```bash
# 后端
cd server
go test ./...
go vet ./...

# 前端
cd web
pnpm install
pnpm lint
pnpm build
```

## 贡献

欢迎提交 Issue 与 Pull Request。建议在 PR 中提供以下信息：

- 变更背景
- 方案说明
- 影响范围
- 验证结果

## 许可证

本项目采用 MIT License，详见 LICENSE。

## 致谢

- OWASP Coraza WAF
- Coraza SPOA
- HAProxy
- Gin
