# AI-WAF 告警系统前端 Web UI 快速启动指南

## 系统概述

已为 AI-WAF 项目创建完整的告警系统前端 Web UI,包含:
- 📢 **告警通道管理** - 支持 Webhook、Slack、Discord、钉钉、企业微信
- 📋 **告警规则配置** - 灵活的条件配置和通道绑定
- 📊 **告警历史查看** - 完整的历史记录和统计信息

## 文件结构

```
web/
├── src/
│   ├── types/alert.ts                          # 类型定义
│   ├── api/alert.ts                            # API服务
│   ├── pages/alert/                            # 页面组件
│   │   ├── layout.tsx
│   │   └── pages/
│   │       ├── channel/page.tsx                # 通道管理
│   │       ├── rule/page.tsx                   # 规则管理  
│   │       └── history/page.tsx                # 历史查看
│   ├── feature/alert/components/               # Feature组件
│   │   ├── ChannelTable.tsx
│   │   ├── ChannelDialog.tsx
│   │   ├── ChannelForm.tsx
│   │   ├── DeleteChannelDialog.tsx
│   │   ├── TestChannelDialog.tsx
│   │   └── index.ts
│   ├── routes/
│   │   ├── constants.ts                        # 已更新: 添加ALERTS路由
│   │   └── config.tsx                          # 已更新: 添加告警路由配置
│   └── components/layout/
│       └── sidebar.tsx                         # 已更新: 添加告警菜单
├── public/locales/
│   ├── en/translation.json                     # 英文翻译
│   └── zh/translation.json                     # 中文翻译
├── update_translations.py                      # 翻译更新脚本
└── ALERT_UI_IMPLEMENTATION.md                  # 详细实现文档
```

## 快速启动步骤

### 1. 更新翻译文件
```bash
cd web
python3 update_translations.py
```

预期输出:
```
✓ web/public/locales/en/translation.json 更新成功
✓ web/public/locales/zh/translation.json 更新成功
```

### 2. 安装依赖(如需要)
```bash
pnpm install
# 或
npm install
```

### 3. 启动开发服务器
```bash
pnpm dev
# 或  
npm run dev
```

### 4. 访问告警系统
打开浏览器访问:
- http://localhost:5173/alerts/channel - 通道管理
- http://localhost:5173/alerts/rule - 规则管理
- http://localhost:5173/alerts/history - 告警历史

## 功能特性

### ✅ 已完成功能

#### 1. 告警通道管理
- [x] 通道列表展示(无限滚动)
- [x] 创建新通道
- [x] 编辑通道配置
- [x] 删除通道
- [x] 启用/禁用切换
- [x] 测试通道功能
- [x] 支持5种通道类型:
  - Webhook (自定义URL + Headers)
  - Slack (Bot Token + Channel)
  - Discord (Webhook URL + Username)
  - 钉钉 (AccessToken + Secret)
  - 企业微信 (Webhook Key)

#### 2. 路由和导航
- [x] 新增 `/alerts` 路由
- [x] 侧边栏添加"告警"菜单项(Bell图标)
- [x] 子导航:通道/规则/历史
- [x] 面包屑导航
- [x] 国际化支持(中英文)

#### 3. UI/UX
- [x] 响应式设计
- [x] 深色模式支持
- [x] 动画过渡效果
- [x] Toast通知
- [x] 加载状态
- [x] 错误处理

### ⏳ 待实现功能

#### 1. 告警规则管理
需要创建:
- [ ] RuleTable.tsx - 规则列表表格
- [ ] RuleForm.tsx - 规则表单
  - 条件配置器(支持多个条件)
  - 通道多选
  - 严重等级选择
  - 冷却时间设置
  - 模板编辑器
- [ ] DeleteRuleDialog.tsx - 删除确认

#### 2. 告警历史查看
需要创建:
- [ ] HistoryTable.tsx - 历史记录表格
  - 时间范围过滤
  - 状态筛选
  - 严重等级筛选
  - 规则筛选
- [ ] HistoryDetailDialog.tsx - 详情对话框
  - 完整消息内容
  - 触发条件
  - 发送状态
  - 确认操作
- [ ] AlertStatsCards.tsx - 统计卡片
  - 总数统计
  - 状态分布
  - 严重等级分布

## 技术栈

- **React 18** - UI框架
- **TypeScript** - 类型安全
- **TanStack Query** - 数据获取和缓存
- **React Hook Form** - 表单管理
- **Zod** - 表单验证
- **Shadcn UI** - 组件库
- **Framer Motion** - 动画
- **i18next** - 国际化
- **Lucide React** - 图标

## API集成

### 端点说明
```typescript
// 通道管理
GET    /alert/channel          // 获取通道列表
POST   /alert/channel          // 创建通道
GET    /alert/channel/:id      // 获取通道详情
PUT    /alert/channel/:id      // 更新通道
DELETE /alert/channel/:id      // 删除通道
POST   /alert/channel/:id/test // 测试通道

// 规则管理
GET    /alert/rule             // 获取规则列表
POST   /alert/rule             // 创建规则
GET    /alert/rule/:id         // 获取规则详情
PUT    /alert/rule/:id         // 更新规则
DELETE /alert/rule/:id         // 删除规则

// 历史管理
GET    /alert/history          // 获取历史列表
GET    /alert/history/:id      // 获取历史详情
POST   /alert/history/:id/acknowledge  // 确认告警
GET    /alert/history/stats    // 获取统计信息
```

### 使用示例
```typescript
import { alertChannelApi } from '@/api/alert'

// 获取通道列表
const { data } = useInfiniteQuery({
    queryKey: ['alertChannels'],
    queryFn: ({ pageParam }) => alertChannelApi.getChannels(pageParam, 20)
})

// 创建通道
const mutation = useMutation({
    mutationFn: alertChannelApi.createChannel,
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['alertChannels'] })
        toast.success('Channel created successfully')
    }
})
```

## 开发指南

### 添加新的通道类型
1. 在 `web/src/types/alert.ts` 中添加新类型到 `AlertChannelType` 枚举
2. 更新 `AlertChannelConfig` 接口添加新配置字段
3. 在 `ChannelForm.tsx` 的 `renderConfigFields()` 中添加新case
4. 更新翻译文件添加新类型名称

### 添加新的告警规则条件
1. 在 `web/src/types/alert.ts` 中更新 `ConditionOperator` 枚举
2. 创建条件配置器组件
3. 在RuleForm中集成条件配置器

### 自定义样式
所有组件使用 Tailwind CSS 和 Shadcn UI,可通过以下方式自定义:
- 修改 `tailwind.config.ts` 中的主题配置
- 使用 `cn()` 工具函数合并类名
- 覆盖 Shadcn UI 组件样式

## 故障排除

### 问题1: 翻译键未找到
**症状**: 页面显示 `alert.channelName` 而不是实际文本
**解决**: 运行 `python3 update_translations.py` 更新翻译文件

### 问题2: 类型错误
**症状**: TypeScript报告类型不匹配
**解决**: 确保 `web/src/types/alert.ts` 文件存在且完整

### 问题3: API请求失败
**症状**: 无法加载通道列表
**解决**: 
- 检查后端服务是否运行
- 确认API端点路径正确
- 查看浏览器控制台网络标签

### 问题4: 路由404
**症状**: 访问 /alerts 显示404
**解决**:
- 确认 `web/src/routes/constants.ts` 包含 `ALERTS: "/alerts"`
- 确认 `web/src/routes/config.tsx` 包含告警路由配置
- 清除浏览器缓存并重新加载

## 下一步开发建议

### 短期 (1-2周)
1. 完成RuleTable和RuleForm组件
2. 完成HistoryTable和详情对话框
3. 添加统计卡片组件
4. 编写单元测试

### 中期 (2-4周)
5. 添加规则条件可视化配置器
6. 实现告警模板编辑器(支持变量)
7. 添加历史记录图表分析
8. 优化性能(虚拟滚动)

### 长期 (1-2月)
9. 添加告警模板市场
10. 实现批量操作
11. 添加告警分组功能
12. 集成AI辅助配置

## 贡献指南

### 代码风格
- 使用TypeScript strict模式
- 遵循ESLint规则
- 使用Prettier格式化代码
- 组件命名使用PascalCase
- 文件命名使用kebab-case

### 提交规范
```
feat: 添加新功能
fix: 修复bug
docs: 更新文档
style: 代码格式调整
refactor: 重构代码
test: 添加测试
chore: 构建或辅助工具变动
```

## 参考资料

- [TanStack Query文档](https://tanstack.com/query/latest)
- [React Hook Form文档](https://react-hook-form.com/)
- [Shadcn UI文档](https://ui.shadcn.com/)
- [Zod文档](https://zod.dev/)
- [Framer Motion文档](https://www.framer.com/motion/)

## 联系方式

如有问题或建议,请查看:
- 项目README: `/README.md`
- 后端API文档: `/doc/ALERT_API_GUIDE.md`
- 详细实现文档: `/web/ALERT_UI_IMPLEMENTATION.md`

---
**最后更新**: 2024年
**版本**: 1.0.0
**状态**: 部分完成,核心功能可用
