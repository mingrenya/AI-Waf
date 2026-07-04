# 2026-07-04 会话进度

## 已完成的任务
- 修复前端 `Outlet` 重复导入导致的运行时语法错误。
- 定位并修复 `/api/v1/nuclei/templates` 返回 500 的问题。
- 将 Nuclei 模板目录改为可配置并支持多候选路径探测。
- 模板目录不存在时返回空列表 + 告警，不再直接抛 500。
- `go test ./...` 后端通过验证。

## 硬编码路径 / 环境依赖排查
- **capture**：`/app/data/captures` → `CAPTURE_DATA_PATH` 环境变量 + 多候选探测
- **backup**：`/app/data/backups` → `BACKUP_DATA_PATH` 环境变量 + 多候选探测
- **K8s 路径**：`config/init.go` → `GEOIP_BASE_PATH` / `HAPROXY_BASE_PATH` 环境变量覆盖
- **FTW**：`server/public/ftw-tests` → `FTW_TESTS_PATH` 环境变量
- **引擎数据库名**：`agent_server.go` 13 处硬编码 `"waf"` → `DB_NAME` 环境变量

## 前后端配置统一
- `.env.template`：重写为结构化完整模板，补齐所有新变量
- 前端 `env.ts`：默认值回退 + 开发环境缺变量告警
- `docker-compose.yaml`：补齐所有环境变量透传，新增 volumes
## 构建验证
- 7 个模块 16 个新测试

## 前端 Bug 修复
- **攻击 IP 跳转 500**：`'/logs/attack'` → `'/logs/event'`
- **攻击日志详情 → 快速处置 IP** ✅ 已实现

## 待实现
- 详见 `docs/DEVELOPMENT_PLAN.md`（17 项，按 P0/P1/P2/P3 分优先级排列）
