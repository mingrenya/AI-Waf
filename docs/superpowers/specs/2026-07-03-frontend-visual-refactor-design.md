# 前端样式、动画、组件重构设计

## 概述

将 AI-Waf 前端从"过度 Glassmorphism"转型为**"克制化 Glassmorphism + Ant Design 企业级专业风"**。参考 Ant Design 设计系统的色板算法、间距网格、阴影层级、圆角体系、动效曲线和 ProLayout 布局模式。

## 设计目标

1. Glassmorphism 保留但做减法——建立 Z 层深度体系（Z=0 实底 / Z=1 微玻璃 / Z=2 浮层 / Z=3 模态）
2. 浅色为主、深色为辅——ThemeProvider 控制，不再强制 dark
3. 参考 Ant Design 色板体系——每个语义色派生出完整色阶，Token 化引用
4. 动效收紧——入场动画仅首屏，悬停只保留 lift，路由切换极简淡入淡出
5. 排版规范化——字号 5 级、间距基于 8px 网格
6. **新增**：顶部 Header Bar（Logo + 状态 + 主题 + AI 助手 + 用户头像）
7. **新增**：Page Header（面包屑 + 操作按钮区）取代当前面包屑组件

---

## 第一节：Z 层深度体系

```
Z=-1  底座背景
      浅色：柔和冷灰渐变（#f5f6fa → #eef0f9）
      深色：紫蓝渐变，动画放缓到 30s
      → 无边框、无模糊，纯色底

Z=0   内容区 / 侧边栏
      浅色：不透明白底或极浅透
      深色：rgba(255,255,255,0.03~0.06) 微透
      → blur 10px，细边框，弱阴影

Z=1   信息容器（统计卡片 / 表格面板 / 图表箱）
      微玻璃：backdrop-blur 16-20px
      浅色：rgba(255,255,255,0.65)
      深色：rgba(255,255,255,0.08)
      → 顶部高光线（border-top 更亮模拟光源）

Z=2   浮层（下拉菜单 / Popover / Tooltip / 选择列表）
      中玻璃：backdrop-blur 24px
      → 更强的边框 + 更大的阴影扩散

Z=3   模态（Dialog / 抽屉 / 大弹窗）
      强玻璃：backdrop-blur 32px
      → 最明显的三维抬起感 + 背景遮罩
```

### 阴影体系（Ant Design 风格，替代均匀 blur）

| Token | 对应层 | 值 |
|-------|--------|-----|
| `--shadow-card` | Z=1 | `0 2px 8px rgba(0,0,0,0.06)` |
| `--shadow-floating` | Z=2 | `0 6px 16px rgba(0,0,0,0.08)` |
| `--shadow-modal` | Z=3 | `0 12px 32px rgba(0,0,0,0.12)` |

### 圆角体系（Ant Design 风格，细化为 4 级）

| Token | 值 | 用途 |
|-------|-----|------|
| `--radius-sm` | 6px | Tag / Badge / Tooltip |
| `--radius-base` | 8px | Button / Input / Select 交互控件 |
| `--radius-lg` | 12px | Card 卡片容器 |
| `--radius-xl` | 16px | Modal / Drawer 大容器 |

### 语义化 CSS 类名

| 旧类名 | 新类名 | 用途 |
|------|------|------|
| `glass`, `glass-nav` | `surface-root` | Z=0 侧边栏/内容底 |
| `glass-card`, `glass-card-light` | `surface-card` | Z=1 微玻璃卡片 |
| — | `surface-card-raised` | Z=1 hover 高一级 |
| `glass-card-emphasis` | `surface-floating` | Z=2 下拉/浮层 |
| 内联 style | `surface-modal` | Z=3 Dialog/抽屉 |
| `glass-input` | 删除（input 用原生样式 + 细边框） | — |
| `glass-btn` | 删除（button 用 Radix variant） | — |

---

## 第二节：颜色体系（Ant Design 色板模式）

### Brand（品牌紫蓝，10 级色阶）

```
primary-1  #ede9fe (最浅背景)
primary-2  #d4c9fc
primary-3  #b7a4f8
primary-4  #9b7ff3
primary-5  #667eea (基准色)
primary-6  #5a6fd6
primary-7  #4d5cc2
primary-8  #3b3fa5
primary-9  #4c1d95 (最深 hover/active)
primary-10 #3a1078
```

### Neutral（冷灰系，13 级灰度）

文字层级：`--text-primary` / `--text-secondary` / `--text-muted` / `--text-dim`
边框层级：浅色 `rgba(31,38,135, 0.06~0.12)` / 深色 `rgba(255,255,255, 0.08~0.18)`
背景层级：page-bg → section-bg → card-bg → overlay-bg

### Semantic（仅用于状态标记，4 级色阶）

- 红 danger：告警/阻断/严重
- 橙 warning：警告/中等风险
- 绿 success：正常/在线
- 蓝 info：信息提示

### 改动影响

- 登录页 features 卡片颜色从 `text-indigo-300/emerald-300/sky-300/amber-300` → 统一 brand 紫色系
- AttackChainTimeline stages 保留色标语义但降低饱和度
- 各处 `text-white/50`, `text-white/40` → 统一 `--text-muted` / `--text-dim`
- `dark:text-shadow-glow-white` → 去掉（玻璃卡片本身已有足够对比度）

---

## 第三节：排版 & 间距（Ant Design 8px 基准网格）

### 字号阶梯（5 级）

| Token | 用途 | 值 |
|-------|------|-----|
| `data-2xl` | 页面主数值 | 32px / font-bold / mono |
| `data-xl` | 卡片统计数 | 24px / font-semibold / mono |
| `heading-lg` | 页面标题 | 20px / font-semibold / sans |
| `heading-sm` | 卡片/区块标题 | 14px / font-medium / sans |
| `body` | 正文/描述 | 14px / sans |
| `caption` | 辅助信息/时间戳/标签 | 12px / sans |

### 字体家族（保留现有）

- 标题/导航/正文 → Inter（系统 UI sans）
- 数据/指标/IP → JetBrains Mono（等宽，tnum + cv01）

### 间距阶梯（8px 基准网格）

- `gap-xs` 8px（图标间距 / Badge 贴紧）
- `gap-sm` 12px（表格列 / 小卡片 padding）
- `gap-md` 16px（卡片内 sections）
- `gap-lg` 24px（统计卡间距，页面常见 margin）
- `gap-xl` 32px（页面 section 间隔）
- `gap-2xl` 40px（页面顶部大间隔）

---

## 第四节：动效体系

### 动效三原则

1. **入场动画 → 仅首屏**：stagger 入场只在首次加载播一次；路由切换用 opacity 交叉淡入淡出（120ms）
2. **悬停动画 → 只保留 lift**：去掉 scale/glow/ripple；统一用 `translateY(-1px)` + 阴影微增强
3. **反馈动画 → 系统交互专用**：按钮点击无动画；告警 severity 标签保留脉冲；加载态保留骨架屏 shimmer

### 动效曲线（Ant Design "自然" 原则）

所有渐变动画使用 `cubic-bezier(0.23, 0.23, 0.23, 0.96)` 替代 `ease-out`，终点更平缓自然。

### 删除的 CSS 类

`.hover-scale`、`.hover-glow`、`.ripple-effect`、`.border-breathe`、`.number-pop`、`.stagger-1` ~ `.stagger-8`

### 保留的动画

- 登录页 `glass-particle` / `float-particle`（登录页专用特色）
- 状态点脉冲 `status-dot-*`（有功能意义）
- 骨架屏 shimmer（加载反馈）
- `hover-lift`（但 translateY -4px → -1px，阴影幅度减小）

---

## 第五节：ProLayout 布局重构（新增）

### 当前布局

```
┌─────────────────────────────────────┐
│ 侧边栏 │ Breadcrumb                  │
│ (Logo  │ Content                     │
│  菜单  │                             │
│  退出) │                             │
└────────┴─────────────────────────────┘
```

### 新布局（ProLayout 风格）

```
┌──────────────────────────────────────────────────────────┐
│  🔒 Logo  │  系统状态  │  主题 ○●  AI 🤖  用户 👤       │  ← Header Bar (48px)
├────────────┼─────────────────────────────────────────────┤
│ 侧边栏     │  📄 监控 / 概览       [刷新] [导出报表]      │  ← Page Header (40px)
│ (仅菜单)   │  ─────────────────────────────────────────── │
│            │  Content (Z=0 纯底)                          │
│            │                                              │
└────────────┴──────────────────────────────────────────────┘
```

### 5.1 Header Bar（新增组件 `components/layout/header-bar.tsx`）

固定高度 48px，横跨顶部。

| 位置 | 元素 | 说明 |
|------|------|------|
| 左 | Logo + 产品名 | 从 sidebar 顶部移到这里，点击可收起/展开侧边栏 |
| 中 | 系统状态指示 | 绿色状态点 + "系统正常" 文字 |
| 右 | 主题切换 | 太阳/月亮图标按钮，调用 ThemeProvider.setTheme |
| 右 | AI 助手 | 🤖 图标按钮，点击导航到 `/ai-analyzer/assistant` |
| 右 | 用户头像 | 头像圆圈 + Radix DropdownMenu（用户名/角色/退出） |

- 背景：实底（无玻璃），`surface-root` token，底部细边框分隔
- 不参与页面滚动（sticky top）

### 5.2 侧边栏调整（`sidebar.tsx`）

- **移除**：顶部 Logo 区域、底部 logout 按钮、version 文字、折叠按钮
- **保留**：导航菜单列表
- 收缩态宽度 48px → 图标居中，hover 时 Tooltip 浮出菜单名
- 展开态宽度 220px → 图标 + 文字，和当前一致
- active 态：从 indigo-300 光点 → 左侧 3px 竖线指示（Ant Design Menu 风格）
- 背景：`surface-root` token

### 5.3 Page Header（升级 `components/layout/page-header.tsx` 替代当前 `breadcrumb.tsx`）

高度 40px，内容区顶部，带面包屑路径。

```
┌────────────────────────────────────────────┐
│ 监控 / 概览              [刷新] [导出报表]   │
└────────────────────────────────────────────┘
   面包屑                      操作按钮插槽
```

- 面包屑从已有 `useBreadcrumbMap()` 自动计算
- 右侧 `actions` 插槽：各页面通过 Context 或 props 注入操作按钮（可选）
- 背景透明，不玻璃

### 5.4 RootLayout 改动（`root-layout.tsx`）

```
删除：
  - className="dark"（改为由 ThemeProvider 控制）
  - AnimatePresence mode="wait"（改为简单 fade）
  - Breadcrumb 组件引用

新增：
  - HeaderBar 组件
  - PageHeader 组件（替代 Breadcrumb）
  - 背景渐变移入 CSS 变量控制
```

新结构：
```tsx
<div className="flex h-screen">
  <Sidebar />
  <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
    <HeaderBar />
    <PageHeader actions={...} />
    <main className="flex-1 overflow-auto p-6">
      <Outlet />
    </main>
  </div>
</div>
```

---

## 第六节：逐组件改动清单

### 6.1 全局 CSS（`index.css`）

- **当前**：718 行，`.glass-*` 类 + 浅色/深色覆写 + 动画类 + 工具类混杂
- **目标**：~400 行，拆分为 3 块：
  - ① Design Tokens（CSS 变量：色板 10 级 / 阴影 3 级 / 圆角 4 级 / Z 层背景值）
  - ② Base Layer（reset / body / #root / 滚动条）
  - ③ Surface Utility Classes（`.surface-root` / `.surface-card` / `.surface-card-raised` / `.surface-floating` / `.surface-modal`）

### 6.2 登录页（`login.tsx`）

- 保留渐变背景 + 浮动粒子
- features 卡片颜色统一 brand 紫色系
- 顶部状态栏（玻璃 nav）不变

### 6.3 态势感知（`situation/`）

| 组件 | 改动 |
|------|------|
| SituationDashboard | `glass-card-light` → `surface-card`；去掉 text-shadow-glow |
| AttackChainTimeline | Stage 色标饱和度降低；hover 用 `surface-card-raised` |
| AttackerDrawer | Dialog 内联 style → `surface-modal`；去掉 icon-neon |
| QuickActionToolbar | 去掉 `dark:icon-neon` |

### 6.4 监控仪表盘（`monitor/`）

- MetricCard → `surface-card`；数值字号统一 `data-xl`；去除多余 hover
- StatusBar → 保留状态点脉冲

### 6.5 安全大屏（`security-dashboard/`）

- Globe3DMap 背景保持深色（地球仪特例，路由级覆盖）
- StatCard / AttackIPList / RealtimeAttackList → `surface-card`
- QPS Chart 容器 → `surface-card`

### 6.6 AI 分析器（`ai-analyzer/`）

- 8 个组件：卡片统一 `surface-card`，Dialog 统一 `surface-modal`
- AttackPatternTable 表头/行用实底色，不做玻璃

### 6.7 告警 / 日志 / 规则页面

- AlertStatsCards → `surface-card`
- 表格容器 → 实底白/灰 + 细边框，不玻璃
- ChannelDialog / RuleDialog 等 → `surface-modal`

---

## 技术约束

- 不引入新的 NPM 依赖
- 不改变组件 API（props 签名不变）
- 不改变路由结构
- 保留双主题切换能力（ThemeProvider 已存在，只是从未启用，现通过 Header Bar 真正可用）
- Tailwind 配置中的 keyframes 和 animation 定义保留，但部分移除
- 不做移动端适配（desktop only）

## 浅色模式适配

当前核心问题：`root-layout.tsx` 强制 `className="dark"`，导致所有页面的浅色模式 CSS 从未被测试过。改动后：

1. 移除强制 dark
2. `surface-*` 类使用 CSS 变量自动适配双主题
3. 各页面移除 hardcode 的 `text-white`、`text-white/50` 等，改用语义 token
4. Header Bar 上的主题切换按钮是浅色模式的入口
