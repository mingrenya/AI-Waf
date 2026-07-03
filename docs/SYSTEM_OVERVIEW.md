# 系统说明（AI-WAF / MRYa WAF）

## 1. 系统概述
AI-WAF 是基于 HAProxy + OWASP Coraza WAF 的现代化 WAF 管理系统，提供流量防护、规则管理、告警、可视化监控与 AI 辅助分析能力。系统由 Go 后端、SPOA 代理、MCP 服务与前端管理台组成，支持 API 驱动与可视化操作。

## 2. 核心能力
- 多引擎安全防护：Coraza WAF + 规则引擎 + 地理位置过滤 + 访问控制
- 规则管理与生效：规则配置、启用禁用、策略管理
- 告警与通知：告警规则、通道配置、历史记录
- 监控与分析：安全指标、流量趋势、攻击来源分析
- AI 辅助：AI 规则建议与分析对话

## 3. 目录结构说明
- coraza-spoa：SPOA 代理与 Coraza 集成逻辑
- server：后端服务（API、规则、告警、配置、数据处理）
- mcp-server：MCP 服务与工具集合，支撑 AI 交互与扩展能力
- web：前端控制台与可视化界面
- doc / docs：历史文档与设计资料
- scripts：运维与管理脚本

## 4. 核心组件说明
### 4.1 coraza-spoa
作为 HAProxy 的 SPOE 代理，实现 Coraza WAF 请求/响应检查与处理，将检测结果回传给 HAProxy。

### 4.2 server
后端 API 服务，提供规则管理、站点管理、证书管理、告警、日志与指标数据接口。

### 4.3 mcp-server
提供 MCP 协议服务与工具扩展，支持 AI 分析、规则建议与自动化能力。

### 4.4 web
前端控制台，负责规则、告警、监控与 AI 交互界面展示。

## 5. 数据流与请求链路
1. 业务流量进入 HAProxy
2. HAProxy 通过 SPOE 将请求/响应交给 coraza-spoa
3. coraza-spoa 调用 Coraza WAF 引擎进行检测
4. 结果回传 HAProxy，执行放行/阻断/记录
5. server 负责管理配置、规则、告警与日志
6. web 通过 API 展示与管理
7. mcp-server 提供 AI 交互与分析能力

## 6. 运行与部署（简述）
- 后端：Go 服务与 MongoDB
- 前端：Vite + React
- 代理：HAProxy + coraza-spoa
- 容器化：Docker / Docker Compose

## 7. 关键配置点
- 环境变量：见 ENV_CONFIG_GUIDE.md
- 规则与策略：通过前端或 API 配置
- HAProxy 与 SPOE：需要正确绑定与通信配置

## 8. 使用建议
- 优先通过前端进行规则与策略管理
- 生产环境建议启用审计与告警
- 保持 HAProxy、Coraza 与规则库版本一致

---

如需进一步的模块细节或 API 说明，请参考 README.md 与相关服务目录中的说明文档。