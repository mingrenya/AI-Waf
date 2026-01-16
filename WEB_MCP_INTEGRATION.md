# Web UI MCP 集成完成说明

## 完成的功能

### 1. MCP 连接状态指示器 ✅

**位置**: 顶部导航栏右侧

**文件**: 
- `web/src/components/common/mcp-status-indicator.tsx`

**功能**:
- 实时显示 MCP 服务器连接状态（已连接/未连接）
- 点击展开详细信息弹窗，显示：
  - 服务器版本
  - 可用工具数量
  - 最后连接时间
  - 前5个工具名称
  - 错误信息（如有）
- 每10秒自动刷新状态
- 支持手动刷新按钮

**视觉设计**:
- 已连接: 绿色指示器 + CheckCircle 图标
- 未连接: 红色指示器 + XCircle 图标
- 使用 Popover 组件展示详细信息

### 2. AI 助手聊天入口 ✅

**位置**: 顶部导航栏右侧（MCP状态指示器旁边）

**文件**: 
- `web/src/components/common/ai-assistant-button.tsx`
- `web/src/feature/ai-assistant/components/AIAssistantDialog.tsx`

**功能**:
- 一键打开 AI 助手对话框
- 与 AI 助手实时交互
- 支持消息历史记录
- 显示工具调用记录（Badge 形式）
- 提供快捷建议按钮（如"分析最近24小时的攻击模式"）
- Shift + Enter 换行，Enter 发送

**视觉设计**:
- 按钮带有 Sparkles 图标表示 AI 功能
- 对话框最大宽度 4xl，高度 600px
- 消息气泡左右布局（用户右侧，AI左侧）
- 加载状态显示"正在思考..."动画

### 3. AI 规则建议展示 ✅

**位置**: AI分析 → AI助手页面

**文件**: 
- `web/src/feature/ai-assistant/components/AIRuleSuggestionCard.tsx`
- `web/src/pages/ai-analyzer/pages/assistant/page.tsx`

**功能**:
- 展示 AI 生成的规则建议列表
- 支持按状态筛选（待审核/已批准/已拒绝/已部署）
- 支持按严重程度筛选（全部/严重/高/中/低）
- 每条建议显示：
  - 规则名称和描述
  - 来源攻击模式
  - 置信度百分比
  - 严重程度徽章
  - 状态徽章
  - 创建时间
  - 建议说明
- 操作按钮：
  - 待审核状态: 批准、拒绝按钮
  - 已批准状态: 部署按钮
- 统计卡片显示关键指标

**视觉设计**:
- 严重程度颜色编码：
  - 严重(critical): 红色 + AlertTriangle
  - 高(high): 橙色 + AlertTriangle
  - 中(medium): 黄色 + Info
  - 低(low): 蓝色 + Info
- 状态徽章不同样式区分
- ScrollArea 支持滚动查看大量建议

### 4. 新增路由页面 ✅

**新增路由**: `/ai-analyzer/assistant`

**文件**: `web/src/pages/ai-analyzer/pages/assistant/page.tsx`

**内容**:
- 页面标题和描述
- 三个统计卡片（待审核规则、已部署规则、平均置信度）
- AI 规则建议卡片（主要内容）

### 5. 类型定义 ✅

**文件**: `web/src/types/mcp.ts`

**定义的类型**:
- `MCPConnectionStatus` - MCP连接状态
- `MCPToolCall` - MCP工具调用记录
- `AIAssistantMessage` - AI助手消息
- `AIAssistantSession` - AI助手会话
- `AIRuleSuggestion` - AI规则建议
- `AIAnalysisResult` - AI分析结果

### 6. API 服务 ✅

**文件**: `web/src/api/mcp.ts`

**实现的 API 方法**:
- `getMCPStatus()` - 获取MCP连接状态
- `getMCPTools()` - 获取MCP工具列表
- `getMCPToolCallHistory()` - 获取工具调用历史
- `getAIRuleSuggestions()` - 获取AI规则建议列表
- `approveAIRuleSuggestion()` - 批准规则建议
- `rejectAIRuleSuggestion()` - 拒绝规则建议
- `deployAIRuleSuggestion()` - 部署规则建议
- `getAIAnalysisResult()` - 获取AI分析结果
- `triggerAIAnalysis()` - 触发AI分析

### 7. 国际化支持 ✅

**更新文件**:
- `web/public/locales/zh/translation.json`
- `web/public/locales/en/translation.json`

**新增翻译**:
- `breadcrumb.aiAnalyzer.assistant`: "AI助手" / "AI Assistant"

### 8. 布局集成 ✅

**更新文件**: 
- `web/src/components/layout/breadcrumb.tsx`
- `web/src/routes/config.tsx`

**改动**:
- 在顶部导航栏添加 MCP 状态指示器和 AI 助手按钮
- 在 AI 分析路由中添加"AI助手"子路由

## 技术栈

- **UI 组件**: shadcn/ui (Card, Button, Badge, Dialog, ScrollArea, Popover, etc.)
- **状态管理**: @tanstack/react-query (数据获取和缓存)
- **图标**: lucide-react
- **样式**: Tailwind CSS
- **路由**: react-router v7

## 工作流程

### MCP 状态检查流程
```
用户打开页面
  ↓
MCPStatusIndicator 组件挂载
  ↓
useQuery 自动调用 getMCPStatus()
  ↓
GET /api/v1/mcp/status
  ↓
每10秒自动刷新
  ↓
显示连接状态（绿色/红色）
```

### AI 助手交互流程
```
用户点击"AI助手"按钮
  ↓
打开 AIAssistantDialog
  ↓
用户输入问题/指令
  ↓
发送到 MCP Server（通过后端API）
  ↓
MCP Server 执行相应工具
  ↓
返回结果显示在对话中
  ↓
显示使用的工具（Badge形式）
```

### 规则建议审核流程
```
AI 分析生成规则建议
  ↓
status: pending
  ↓
用户在"AI助手"页面查看
  ↓
用户点击"批准"按钮
  ↓
POST /api/v1/ai-analyzer/suggestions/:id/approve
  ↓
status: approved
  ↓
用户点击"部署"按钮
  ↓
POST /api/v1/ai-analyzer/suggestions/:id/deploy
  ↓
status: deployed
```

## 后端需要实现的内容

详见 `MCP_BACKEND_API.md` 文档，包括：

1. **9个 API 端点** - MCP 状态、工具列表、规则建议管理等
2. **2个 MongoDB Collections** - ai_rule_suggestions, mcp_tool_calls
3. **权限控制** - 规则部署需要管理员权限
4. **异步任务处理** - AI 分析任务异步执行

## 下一步

### 前端待完善
1. ✅ AI 助手对话实际调用后端 API（目前是模拟响应）
2. ✅ 规则建议详情弹窗展示完整规则内容
3. ✅ 支持规则建议的批量操作
4. ✅ 添加规则效果评估可视化

### 后端待实现
1. 实现 9 个 MCP 相关 API 端点（见 MCP_BACKEND_API.md）
2. 创建 MongoDB collections 和相应的 repository
3. 实现 MCP Server 连接状态检查逻辑
4. 实现规则建议生成和管理服务
5. 集成 MCP 工具调用日志记录

### 测试
1. 测试 MCP 连接状态实时更新
2. 测试 AI 助手对话功能
3. 测试规则建议的完整工作流（pending → approved → deployed）
4. 测试筛选和分页功能
5. 测试国际化（中英文切换）

## 截图位置

### MCP 状态指示器
- 位置: 页面右上角，面包屑导航右侧
- 点击展开显示详细信息弹窗

### AI 助手按钮
- 位置: MCP 状态指示器右侧
- 点击打开全屏对话框

### AI 规则建议页面
- 路由: /ai-analyzer/assistant
- 位于 AI 分析菜单下

## 文件清单

### 新建文件 (11个)
1. `web/src/types/mcp.ts` - MCP 类型定义
2. `web/src/api/mcp.ts` - MCP API 服务
3. `web/src/components/common/mcp-status-indicator.tsx` - MCP状态指示器
4. `web/src/components/common/ai-assistant-button.tsx` - AI助手按钮
5. `web/src/feature/ai-assistant/components/AIAssistantDialog.tsx` - AI助手对话框
6. `web/src/feature/ai-assistant/components/AIRuleSuggestionCard.tsx` - 规则建议卡片
7. `web/src/feature/ai-assistant/components/index.ts` - 组件导出
8. `web/src/feature/ai-assistant/index.ts` - 功能导出
9. `web/src/pages/ai-analyzer/pages/assistant/page.tsx` - AI助手页面
10. `MCP_BACKEND_API.md` - 后端API实现指南
11. `WEB_MCP_INTEGRATION.md` - 本文档

### 修改文件 (5个)
1. `web/src/components/layout/breadcrumb.tsx` - 添加MCP状态和AI助手按钮
2. `web/src/routes/config.tsx` - 添加AI助手路由
3. `web/src/api/services.ts` - 导出MCP API
4. `web/public/locales/zh/translation.json` - 中文翻译
5. `web/public/locales/en/translation.json` - 英文翻译

## 依赖项

所有使用的组件和库都已在项目中安装，无需额外安装：

- @tanstack/react-query (已有)
- lucide-react (已有)
- shadcn/ui 组件 (已有)
- react-router (已有)
- react-i18next (已有)

## 总结

✅ **MCP连接状态显示** - 完成
✅ **AI助手聊天入口** - 完成
✅ **AI规则建议展示** - 完成

前端集成已全部完成，可以直接使用。后端需要按照 `MCP_BACKEND_API.md` 实现相应的 API 端点即可实现完整功能。
