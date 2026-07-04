# 前端视觉重构实施计划

> **For agentic workers:** 使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实施。步骤使用 `- [ ]` checkbox 跟踪。

**目标：** 将前端从"过度 Glassmorphism"转型为克制化 Glassmorphism + Ant Design 企业级风格，新增 ProLayout 顶部 Header Bar。

**架构：** 先建 Design Token 底座（CSS 变量），再建布局骨架（Header Bar / Page Header / Sidebar），最后逐页面替换旧 class 并清理硬编码颜色。每个 Task 独立可测、可 Review。

**技术栈：** React 18 + TypeScript + Tailwind CSS 3.4 + Motion 12.6 + Zustand 5 + Radix UI

## 全局约束

- 不引入新 NPM 依赖
- 不改变组件 API（props 签名不变）
- 不改变路由结构
- 不做移动端适配（desktop only）
- 保留双主题切换（ThemeProvider + ThemeStore 已存在，通过 Header Bar 暴露入口）
- 所有渐变动效用 `cubic-bezier(0.23, 0.23, 0.23, 0.96)` 替代 `ease-out`
- 间距基准 8px 网格

---

### Task 1: 重构全局 CSS Design Tokens

**文件：**
- 修改：`web/src/index.css`（完整重写）

**接口：**
- 产出：CSS 变量 `--surface-root-bg`, `--surface-card-bg`, `--surface-floating-bg`, `--surface-modal-bg` 及其深/浅色值
- 产出：CSS 变量 `--shadow-card`, `--shadow-floating`, `--shadow-modal`
- 产出：CSS 变量 `--radius-sm` (6px), `--radius-base` (8px), `--radius-lg` (12px), `--radius-xl` (16px)
- 产出：CSS 变量 `--color-primary-1` ~ `--color-primary-10` (品牌紫蓝色板)
- 产出：CSS 变量 `--text-primary`, `--text-secondary`, `--text-muted`, `--text-dim`
- 产出：5 个语义类 `.surface-root`, `.surface-card`, `.surface-card-raised`, `.surface-floating`, `.surface-modal`
- 产出：CSS 变量 `--ease-natural` = `cubic-bezier(0.23, 0.23, 0.23, 0.96)`

- [ ] **Step 1: 完整重写 `web/src/index.css`**

用以下内容替换整个文件：

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

/* ============================================
   Design Tokens — 浅色模式
   ============================================ */
:root {
  /* 底座背景 */
  --bg-page: linear-gradient(135deg, #f5f6fa 0%, #eef0f9 100%);

  /* Surface 背景 (Z=0 实底) */
  --surface-root-bg: #ffffff;
  --surface-root-border: rgba(31, 38, 135, 0.06);

  /* Surface 卡片 (Z=1 微玻璃) */
  --surface-card-bg: rgba(255, 255, 255, 0.65);
  --surface-card-bg-hover: rgba(255, 255, 255, 0.82);
  --surface-card-border: rgba(31, 38, 135, 0.08);
  --surface-card-border-top: rgba(255, 255, 255, 0.7);

  /* Surface 浮层 (Z=2 中玻璃) */
  --surface-floating-bg: rgba(255, 255, 255, 0.8);
  --surface-floating-border: rgba(31, 38, 135, 0.12);

  /* Surface 模态 (Z=3 强玻璃) */
  --surface-modal-bg: rgba(255, 255, 255, 0.9);
  --surface-modal-border: rgba(31, 38, 135, 0.15);

  /* 阴影 (Ant Design 风格，级联) */
  --shadow-card: 0 2px 8px rgba(0, 0, 0, 0.06);
  --shadow-card-hover: 0 4px 12px rgba(0, 0, 0, 0.08);
  --shadow-floating: 0 6px 16px rgba(0, 0, 0, 0.08);
  --shadow-modal: 0 12px 32px rgba(0, 0, 0, 0.12);

  /* 圆角 */
  --radius-sm: 6px;
  --radius-base: 8px;
  --radius-lg: 12px;
  --radius-xl: 16px;

  /* 品牌色板 (Primary — 紫蓝) */
  --color-primary-1: #ede9fe;
  --color-primary-2: #d4c9fc;
  --color-primary-3: #b7a4f8;
  --color-primary-4: #9b7ff3;
  --color-primary-5: #667eea;
  --color-primary-6: #5a6fd6;
  --color-primary-7: #4d5cc2;
  --color-primary-8: #3b3fa5;
  --color-primary-9: #4c1d95;
  --color-primary-10: #3a1078;

  /* 文本层级 */
  --text-primary: #0f1133;
  --text-secondary: rgba(15, 17, 51, 0.65);
  --text-muted: rgba(15, 17, 51, 0.45);
  --text-dim: rgba(15, 17, 51, 0.25);

  /* 动效 */
  --ease-natural: cubic-bezier(0.23, 0.23, 0.23, 0.96);

  /* 保留 — Tailwind 主题变量兼容 */
  --background: 240 20% 97%;
  --foreground: 240 10% 10%;
  --card: 255 255 255 / 0.6;
  --card-foreground: 240 10% 10%;
  --popover: 240 20% 99%;
  --popover-foreground: 240 10% 10%;
  --primary: 262 83% 58%;
  --primary-foreground: 0 0% 100%;
  --secondary: 240 10% 94%;
  --secondary-foreground: 240 6% 20%;
  --muted: 240 10% 94%;
  --muted-foreground: 240 4% 50%;
  --accent: 262 83% 58%;
  --accent-foreground: 0 0% 100%;
  --destructive: 0 72% 51%;
  --destructive-foreground: 0 0% 100%;
  --success: 142 71% 40%;
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 45%;
  --warning-foreground: 0 0% 100%;
  --border: 240 6% 88%;
  --input: 240 6% 92%;
  --ring: 262 83% 58%;
  --radius: 0.75rem;
  --chart-1: 262 83% 58%;
  --chart-2: 217 91% 60%;
  --chart-3: 330 81% 60%;
  --chart-4: 142 71% 40%;
  --chart-5: 38 92% 45%;
  --scrollbar-track: 240 6% 90%;
  --scrollbar-thumb: 240 4% 75%;
  --scrollbar-thumb-hover: 240 4% 55%;
  --font-sans: 'Inter', ui-sans-serif, system-ui, sans-serif;
  --font-mono: 'JetBrains Mono', ui-monospace, monospace;
}

/* ============================================
   Design Tokens — 深色模式
   ============================================ */
.dark {
  --bg-page: linear-gradient(135deg, #1a1040 0%, #0f0728 50%, #1a1040 100%);
  background-size: 400% 400%;
  animation: gradientShift 30s ease infinite;

  --surface-root-bg: rgba(255, 255, 255, 0.03);
  --surface-root-border: rgba(255, 255, 255, 0.08);

  --surface-card-bg: rgba(255, 255, 255, 0.06);
  --surface-card-bg-hover: rgba(255, 255, 255, 0.1);
  --surface-card-border: rgba(255, 255, 255, 0.1);
  --surface-card-border-top: rgba(255, 255, 255, 0.14);

  --surface-floating-bg: rgba(255, 255, 255, 0.1);
  --surface-floating-border: rgba(255, 255, 255, 0.15);

  --surface-modal-bg: rgba(255, 255, 255, 0.12);
  --surface-modal-border: rgba(255, 255, 255, 0.18);

  --shadow-card: 0 2px 8px rgba(0, 0, 0, 0.2);
  --shadow-card-hover: 0 4px 12px rgba(0, 0, 0, 0.3);
  --shadow-floating: 0 6px 16px rgba(0, 0, 0, 0.25);
  --shadow-modal: 0 12px 32px rgba(0, 0, 0, 0.4);

  --text-primary: #ffffff;
  --text-secondary: rgba(255, 255, 255, 0.65);
  --text-muted: rgba(255, 255, 255, 0.45);
  --text-dim: rgba(255, 255, 255, 0.25);

  /* Tailwind 兼容变量 */
  --background: 240 10% 3.9%;
  --foreground: 0 0% 98%;
  --card: 255 255 255 / 0.1;
  --card-foreground: 0 0% 100%;
  --popover: 240 10% 3.9%;
  --popover-foreground: 0 0% 98%;
  --primary: 217 91% 60%;
  --primary-foreground: 0 0% 100%;
  --secondary: 240 4% 16%;
  --secondary-foreground: 0 0% 98%;
  --muted: 240 4% 16%;
  --muted-foreground: 240 5% 65%;
  --accent: 262 83% 58%;
  --accent-foreground: 0 0% 100%;
  --destructive: 0 63% 31%;
  --destructive-foreground: 0 0% 100%;
  --success: 142 71% 45%;
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 50%;
  --warning-foreground: 0 0% 100%;
  --border: 0 0% 100% / 0.1;
  --input: 0 0% 100% / 0.05;
  --ring: 217 91% 60%;
  --chart-1: 217 91% 60%;
  --chart-2: 262 83% 58%;
  --chart-3: 330 81% 60%;
  --chart-4: 142 71% 45%;
  --chart-5: 38 92% 50%;
  --scrollbar-track: 240 6% 15%;
  --scrollbar-thumb: 240 4% 35%;
  --scrollbar-thumb-hover: 240 4% 55%;
}

/* ============================================
   Base Layer
   ============================================ */
@layer base {
  * {
    @apply border-border;
  }

  html,
  body {
    color: var(--text-primary);
    font-family: 'Inter', ui-sans-serif, system-ui, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    height: 100%;
    overflow: hidden;
  }

  #root {
    height: 100%;
    overflow: hidden;
  }
}

/* ============================================
   Surface Utility Classes
   ============================================ */

/* Z=0 实底 — 侧边栏 / 内容区底 */
.surface-root {
  background: var(--surface-root-bg);
  border-color: var(--surface-root-border);
}

/* Z=1 微玻璃卡片 — 统计卡片 / 图表容器 / 表格面板 */
.surface-card {
  background: var(--surface-card-bg);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border: 1px solid var(--surface-card-border);
  border-top: 1px solid var(--surface-card-border-top);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  transition: all 0.2s var(--ease-natural);
}
.surface-card:hover {
  background: var(--surface-card-bg-hover);
  box-shadow: var(--shadow-card-hover);
  transform: translateY(-1px);
}

/* Z=1 hover 提升态 */
.surface-card-raised {
  background: var(--surface-card-bg-hover);
  box-shadow: var(--shadow-card-hover);
}

/* Z=2 浮层 — 下拉菜单 / Popover / 选择 */
.surface-floating {
  background: var(--surface-floating-bg);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--surface-floating-border);
  border-radius: var(--radius-base);
  box-shadow: var(--shadow-floating);
}

/* Z=3 强玻璃模态 — Dialog / 抽屉 */
.surface-modal {
  background: var(--surface-modal-bg);
  backdrop-filter: blur(32px);
  -webkit-backdrop-filter: blur(32px);
  border: 1px solid var(--surface-modal-border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-modal);
}

/* ============================================
   滚动条
   ============================================ */
.scrollbar-custom {
  scrollbar-width: thin;
  scrollbar-color: hsl(var(--scrollbar-thumb)) hsl(var(--scrollbar-track));
}
.scrollbar-custom::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
.scrollbar-custom::-webkit-scrollbar-track {
  background: hsl(var(--scrollbar-track));
  border-radius: 4px;
}
.scrollbar-custom::-webkit-scrollbar-thumb {
  background: hsl(var(--scrollbar-thumb));
  border-radius: 4px;
}
.scrollbar-custom::-webkit-scrollbar-thumb:hover {
  background: hsl(var(--scrollbar-thumb-hover));
}

/* 玻璃风格滚动条 */
.scrollbar-glass::-webkit-scrollbar { width: 6px; }
.scrollbar-glass::-webkit-scrollbar-track { background: rgba(255, 255, 255, 0.05); }
.scrollbar-glass::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}
.scrollbar-glass::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* ============================================
   保留 — 登录页专用动画
   ============================================ */
.glass-bg-animated {
  background: linear-gradient(135deg, #667eea, #764ba2, #f093fb, #4facfe);
  background-size: 400% 400%;
  animation: gradientShift 15s ease infinite;
}
@keyframes gradientShift {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.glass-particle {
  position: absolute;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.2), rgba(255, 255, 255, 0.05));
  animation: float-particle 20s ease-in-out infinite;
  pointer-events: none;
}
@keyframes float-particle {
  0%, 100% { transform: translate(0, 0) scale(1); }
  25% { transform: translate(30px, -40px) scale(1.1); }
  50% { transform: translate(-20px, -80px) scale(0.9); }
  75% { transform: translate(-40px, -40px) scale(1.05); }
}

/* ============================================
   保留 — 状态指示动画
   ============================================ */
.status-dot-online {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #22c55e;
  animation: statusPulse 2s ease-in-out infinite;
}
.status-dot-warning {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #eab308;
  animation: statusPulse 1.5s ease-in-out infinite;
}
.status-dot-critical {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #ef4444;
  animation: statusPulse 1s ease-in-out infinite;
}
@keyframes statusPulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 0 0 currentColor; }
  50% { opacity: 0.6; box-shadow: 0 0 0 6px transparent; }
}

/* ============================================
   保留 — 骨架屏
   ============================================ */
.skeleton-shimmer {
  background: linear-gradient(90deg,
    rgba(255,255,255,0.06) 25%,
    rgba(255,255,255,0.12) 50%,
    rgba(255,255,255,0.06) 75%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
}
@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

- [ ] **Step 2: 验证 CSS 文件无语法错误**

```bash
cd /Users/duheling/AI-Waf/web && npx tailwindcss -i src/index.css -o /dev/null --dry-run 2>&1 | tail -5
```

预期：无报错输出，tailwind 正常编译。

- [ ] **Step 3: 提交**

```bash
git add web/src/index.css
git commit -m "refactor(ui): 重写 CSS Design Tokens — 建立 Z 层深度体系 + Ant Design 色板"
```

---

### Task 2: 清理 Tailwind 配置

**文件：**
- 修改：`web/tailwind.config.ts`

**接口：**
- 使用：Task 1 中定义的 CSS 变量 `--radius-sm/base/lg/xl`
- 产出：移除不再使用的 keyframes / animation 定义

- [ ] **Step 1: 删除废弃的 keyframes 和 animation 定义**

将以下行从 keyframes 中删除：
- `"ripple"`, `"breathe"`, `"number-pop"`, `"spin-slow"`, `"icon-shake"`, `"wiggle"`（这些效果已被移出 CSS）
- 删除 `"pulse-green"` keyframe

对应 animation 中删除：
- `"icon-shake"`, `"wiggle"`, `"pulse-green"`, `"ping-slow"`

保留以下 keyframes（仍在 CSS 或组件中使用）：
- `"accordion-down"`, `"accordion-up"`, `"fade-in"`, `"slide-up"`, `"slide-in-right"`, `"scale-in"`, `"fade-in-up"`, `"pulse-glow"`, `"gradient-shift"`

```typescript
// 从 config.theme.extend.keyframes 中删除：
"icon-shake": {
    "0%": { transform: "rotate(0deg)" },
    "25%": { transform: "rotate(-12deg)" },
    "50%": { transform: "rotate(10deg)" },
    "75%": { transform: "rotate(-6deg)" },
    "85%": { transform: "rotate(3deg)" },
    "92%": { transform: "rotate(-2deg)" },
    "100%": { transform: "rotate(0deg)" },
},
"wiggle": {
    "0%, 100%": { transform: "rotate(-2deg)" },
    "50%": { transform: "rotate(2deg)" },
},
"pulse-green": {
    "0%, 100%": { opacity: "1" },
    "50%": { opacity: "0.4" },
},
"ping-slow": {
    "0%": { transform: "scale(1)", opacity: "1" },
    "100%": { transform: "scale(1.5)", opacity: "0" },
},

// 从 config.theme.extend.animation 中删除对应项
```

- [ ] **Step 2: 更新圆角 token 引用为 CSS 变量**

```typescript
// tailwind.config.ts — theme.extend.borderRadius 修改为：
borderRadius: {
    lg: 'var(--radius-lg)',
    md: 'var(--radius-base)',
    sm: 'var(--radius-sm)',
},
```

- [ ] **Step 3: 删除 boxShadow 自定义定义**

```typescript
// 删除 config.theme.extend.boxShadow 块：
boxShadow: {
    "glass": "0 8px 32px 0 rgba(31,38,135,0.37)",
    "glass-hover": "0 12px 48px 0 rgba(31,38,135,0.45)",
    "glass-light": "0 4px 16px 0 rgba(31,38,135,0.2)",
},
```
（阴影现在由 CSS 变量 `--shadow-*` 管理）

- [ ] **Step 4: 验证 TypeScript 编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | head -20
```

预期：无与 tailwind 配置相关的类型错误。

- [ ] **Step 5: 提交**

```bash
git add web/tailwind.config.ts
git commit -m "refactor(ui): 清理 Tailwind 配置 — 删除废弃 keyframes，rounding/shadow 改为 CSS 变量引用"
```

---

### Task 3: 新建 Header Bar 组件

**文件：**
- 创建：`web/src/components/layout/header-bar.tsx`

**接口：**
- 使用：`useThemeStore` from `@/store` (setTheme, resolvedTheme)
- 使用：`useAuthStore` from `@/store` (user, logout)
- 使用：`useNavigate` from `react-router`
- 使用：Radix `DropdownMenu` from `@/components/ui/dropdown-menu`
- 使用：`Sun`, `Moon`, `Bot`, `Shield` from `lucide-react`
- 使用：`useTranslation` from `react-i18next`
- 产出：固定高度 48px 的顶部栏组件

- [ ] **Step 1: 创建 `web/src/components/layout/header-bar.tsx`**

```tsx
import { useNavigate } from 'react-router';
import { Sun, Moon, Bot, Shield, LogOut, User } from 'lucide-react';
import { useThemeStore, useAuthStore } from '@/store';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useTranslation } from 'react-i18next';
import { ROUTES } from '@/routes/constants';

export function HeaderBar() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { resolvedTheme, setTheme } = useThemeStore();
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header
      className="h-12 flex items-center justify-between px-4 flex-shrink-0 surface-root border-b z-10"
      style={{ borderBottom: '1px solid var(--surface-root-border)' }}
    >
      {/* 左：Logo + 产品名 */}
      <div className="flex items-center gap-2">
        <div
          className="w-7 h-7 rounded-md flex items-center justify-center"
          style={{ background: 'linear-gradient(135deg, var(--color-primary-5), var(--color-primary-9))' }}
        >
          <Shield className="w-4 h-4 text-white" />
        </div>
        <span className="font-bold text-sm" style={{ color: 'var(--text-primary)' }}>
          MRYa WAF
        </span>
      </div>

      {/* 中：系统状态 */}
      <div className="flex items-center gap-2">
        <span className="status-dot-online" />
        <span className="text-xs" style={{ color: 'var(--text-muted)' }}>
          {t('system.status', { defaultValue: 'System Normal' })}
        </span>
      </div>

      {/* 右：操作区 */}
      <div className="flex items-center gap-1">
        {/* 主题切换 */}
        <button
          onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
          className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-black/5 dark:hover:bg-white/10 transition-colors"
          title={resolvedTheme === 'dark' ? 'Switch to light' : 'Switch to dark'}
          style={{ color: 'var(--text-secondary)' }}
        >
          {resolvedTheme === 'dark' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
        </button>

        {/* AI 助手 */}
        <button
          onClick={() => navigate(ROUTES.AI_ANALYZER + '/assistant')}
          className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-black/5 dark:hover:bg-white/10 transition-colors"
          title={t('sidebar.aiAssistant', { defaultValue: 'AI Assistant' })}
          style={{ color: 'var(--text-secondary)' }}
        >
          <Bot className="w-4 h-4" />
        </button>

        {/* 用户头像 */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium text-white ml-1"
              style={{ background: 'linear-gradient(135deg, var(--color-primary-5), var(--color-primary-9))' }}
            >
              {user?.username?.charAt(0).toUpperCase() ?? 'U'}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48 surface-floating">
            <DropdownMenuLabel>
              <div className="flex flex-col gap-1">
                <span style={{ color: 'var(--text-primary)' }} className="text-sm">{user?.username}</span>
                <span style={{ color: 'var(--text-muted)' }} className="text-xs font-normal">{user?.role}</span>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => navigate('/settings/user')} className="cursor-pointer">
              <User className="w-4 h-4 mr-2" />
              {t('auth.profile', { defaultValue: 'Profile' })}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleLogout} className="cursor-pointer text-red-500">
              <LogOut className="w-4 h-4 mr-2" />
              {t('sidebar.logout')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
```

- [ ] **Step 2: 验证 TypeScript 编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | grep -i "header-bar\|error" | head -10
```

预期：无 header-bar 相关的类型错误。

- [ ] **Step 3: 提交**

```bash
git add web/src/components/layout/header-bar.tsx
git commit -m "feat(ui): 新增 Header Bar 组件 — Logo + 状态 + 主题切换 + AI 助手 + 用户菜单"
```

---

### Task 4: 升级 Page Header（替代 Breadcrumb）

**文件：**
- 修改：`web/src/components/layout/breadcrumb.tsx`（重写为 page-header）
- 重命名：不变，保持文件路径兼容

**接口：**
- 使用：`useBreadcrumbMap` from `@/routes/config`
- 使用：`useLocation`, `Link` from `react-router`
- 使用：`ChevronRight` from `lucide-react`
- 产出：`PageHeader` 组件，面包屑路径 + 可选 `actions` 插槽

- [ ] **Step 1: 重写 `web/src/components/layout/breadcrumb.tsx`**

```tsx
import { Link, useLocation } from 'react-router';
import { ChevronRight } from 'lucide-react';
import { type RoutePath, useBreadcrumbMap, type BreadcrumbConfig } from '@/routes/config';
import type { ReactNode } from 'react';

interface PageHeaderProps {
  actions?: ReactNode;
}

export function PageHeader({ actions }: PageHeaderProps) {
  const location = useLocation();
  const breadcrumbMap = useBreadcrumbMap();

  const segments = location.pathname.split('/').filter(Boolean);
  if (segments.length === 0) return null;

  const parentPath = ('/' + segments[0]) as RoutePath;
  const config: BreadcrumbConfig | undefined = breadcrumbMap[parentPath];
  const currentItem = config?.items?.find((item) => item.path === segments[1]);
  const parentLabel = getParentLabel(parentPath);

  return (
    <nav
      className="h-10 px-6 flex items-center gap-2 text-sm flex-shrink-0"
      style={{
        borderBottom: '1px solid var(--surface-root-border)',
        background: 'var(--surface-root-bg)',
      }}
    >
      <Link
        to={parentPath}
        style={{ color: 'var(--text-muted)' }}
        className="hover:underline transition-colors"
      >
        {parentLabel}
      </Link>
      {currentItem && (
        <>
          <ChevronRight className="w-3 h-3" style={{ color: 'var(--text-dim)' }} />
          <span className="font-medium" style={{ color: 'var(--text-primary)' }}>
            {currentItem.title}
          </span>
        </>
      )}
      {actions && (
        <div className="ml-auto flex items-center gap-2">
          {actions}
        </div>
      )}
    </nav>
  );
}

function getParentLabel(path: string): string {
  const labels: Record<string, string> = {
    '/monitor': '监控',
    '/logs': '日志',
    '/rules': '规则',
    '/alerts': '告警',
    '/settings': '设置',
    '/ai-analyzer': 'AI 分析器',
    '/nuclei': 'Nuclei',
    '/capture': '流量捕获',
    '/situation': '态势感知',
    '/ftw': 'FTW 测试',
  };
  return labels[path] ?? path;
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | grep -i "breadcrumb\|page-header\|error" | head -10
```

- [ ] **Step 3: 提交**

```bash
git add web/src/components/layout/breadcrumb.tsx
git commit -m "refactor(ui): 面包屑升级为 Page Header — 支持 actions 插槽 + 文本 Token 化"
```

---

### Task 5: 重构侧边栏

**文件：**
- 修改：`web/src/components/layout/sidebar.tsx`

**接口：**
- 使用：`useLocation`, `useNavigate` from `react-router`
- 使用：Task 1 的 `surface-root` CSS 类及文本 Token
- 使用：`useTranslation` from `react-i18next`
- 产出：移除 Logo / logout / version，保留菜单导航，active 态改为左侧竖线

- [ ] **Step 1: 重写 `web/src/components/layout/sidebar.tsx`**

```tsx
import { Link, useLocation } from 'react-router';
import { cn } from '@/lib/utils';
import { type ComponentType } from 'react';
import {
  Settings, Shield, BarChart2, FileText, Bell, Brain,
  ScanSearch, Crosshair, Radio,
} from 'lucide-react';
import { ROUTES } from '@/routes/constants';
import { useTranslation } from 'react-i18next';

interface SidebarItem {
  title: string;
  icon: ComponentType<{ className?: string }>;
  href: string;
}

function createSidebarItems(t: (key: string) => string): SidebarItem[] {
  return [
    { title: t('sidebar.monitor'), icon: BarChart2, href: ROUTES.MONITOR },
    { title: t('sidebar.logs'), icon: FileText, href: ROUTES.LOGS },
    { title: t('sidebar.rules'), icon: Shield, href: ROUTES.RULES },
    { title: t('sidebar.alerts'), icon: Bell, href: ROUTES.ALERTS },
    { title: t('sidebar.situation'), icon: ScanSearch, href: ROUTES.SITUATION },
    { title: t('sidebar.nuclei'), icon: Crosshair, href: ROUTES.NUCLEI },
    { title: t('sidebar.capture'), icon: Radio, href: ROUTES.CAPTURE },
    { title: t('sidebar.aiAnalyzer'), icon: Brain, href: ROUTES.AI_ANALYZER },
    { title: t('sidebar.settings'), icon: Settings, href: ROUTES.SETTINGS },
  ];
}

interface SidebarProps {
  allowedItems?: string[];
}

export function Sidebar({ allowedItems }: SidebarProps) {
  const location = useLocation();
  const { t } = useTranslation();

  const currentFirstLevelPath = '/' + location.pathname.split('/')[1];

  const items = createSidebarItems(t).filter((item) => {
    if (!allowedItems) return true;
    return allowedItems.includes(item.href);
  });

  return (
    <aside
      className="h-full flex flex-col surface-root"
      style={{
        width: '220px',
        minWidth: '220px',
        borderRight: '1px solid var(--surface-root-border)',
      }}
    >
      {/* 导航菜单 */}
      <nav className="flex-1 py-4 space-y-1 overflow-y-auto scrollbar-glass">
        {items.map((item) => {
          const isActive = currentFirstLevelPath === item.href;
          return (
            <Link
              key={item.href}
              to={item.href}
              className={cn(
                'flex items-center gap-3 px-4 py-2.5 mx-2 rounded-md text-sm font-medium transition-colors duration-150 relative',
              )}
              style={{
                color: isActive
                  ? 'var(--color-primary-5)'
                  : 'var(--text-secondary)',
                background: isActive ? 'var(--color-primary-1)' : 'transparent',
              }}
            >
              {/* Active 指示竖线 */}
              {isActive && (
                <div
                  className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full"
                  style={{ background: 'var(--color-primary-5)' }}
                />
              )}
              <item.icon className="w-5 h-5 flex-shrink-0" />
              <span>{item.title}</span>
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
```

- [ ] **Step 2: 在所有引用处更新 import**

检查以下文件的 import：
- `web/src/components/layout/root-layout.tsx`（Task 6 会改）
- `web/src/pages/*`（如果有直接引用 sidebar 的——通过 root-layout 间接引用即可）

```bash
cd /Users/duheling/AI-Waf/web && grep -r "from.*sidebar\|import.*Sidebar" src/ --include="*.tsx" --include="*.ts"
```

确认只有 `root-layout.tsx` 引用 Sidebar。

- [ ] **Step 3: 验证编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 4: 提交**

```bash
git add web/src/components/layout/sidebar.tsx
git commit -m "refactor(ui): 侧边栏精简 — 移除 Logo/logout，Ant Design 竖线 active 指示，token 化颜色"
```

---

### Task 6: 重构 RootLayout

**文件：**
- 修改：`web/src/components/layout/root-layout.tsx`

**接口：**
- 使用：Task 3 的 `HeaderBar`
- 使用：Task 4 的 `PageHeader`
- 使用：Task 5 的新 `Sidebar`
- 使用：`useAuthStore` from `@/store`
- 使用：`Outlet`, `useLocation` from `react-router`
- 使用：`motion` from `motion/react`（仅保留简单 fade）
- 产出：ProLayout 三段式结构（Header / PageHeader / Content）

- [ ] **Step 1: 重写 `web/src/components/layout/root-layout.tsx`**

```tsx
import { Outlet } from 'react-router';
import { Sidebar } from './sidebar';
import { HeaderBar } from './header-bar';
import { PageHeader } from './breadcrumb';
import { useAuthStore } from '@/store/auth';
import { hasAnyPermission } from '@/lib/permissions';
import { ROUTES } from '@/routes/constants';
import { useMemo } from 'react';

export function RootLayout() {
  const user = useAuthStore((state) => state.user);

  const allowedSidebarItems = useMemo(() => {
    if (!user) return [];

    const mapping: { route: string; permissions: string[] }[] = [
      { route: ROUTES.MONITOR, permissions: ['system:status'] },
      { route: ROUTES.LOGS, permissions: ['waf:log:read'] },
      { route: ROUTES.RULES, permissions: ['config:read'] },
      { route: ROUTES.ALERTS, permissions: ['alert:channel:read'] },
      { route: ROUTES.SITUATION, permissions: ['waf:log:read'] },
      { route: ROUTES.NUCLEI, permissions: ['config:read'] },
      { route: ROUTES.CAPTURE, permissions: ['config:read'] },
      { route: ROUTES.AI_ANALYZER, permissions: ['waf:log:read'] },
      { route: ROUTES.SETTINGS, permissions: ['config:read'] },
    ];

    return mapping
      .filter((m) => hasAnyPermission(user, m.permissions))
      .map((m) => m.route);
  }, [user]);

  return (
    <div
      className="flex h-screen"
      style={{
        background: 'var(--bg-page)',
        backgroundAttachment: 'fixed',
      }}
    >
      <Sidebar allowedItems={allowedSidebarItems} />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <HeaderBar />
        <PageHeader />
        <main
          className="flex-1 overflow-auto p-6"
          style={{ background: 'transparent' }}
        >
          <Outlet />
        </main>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 3: 提交**

```bash
git add web/src/components/layout/root-layout.tsx
git commit -m "refactor(ui): RootLayout 升级为 ProLayout — Header Bar + Page Header + Sidebar 三段式"
```

---

### Task 7: 态势感知组件样式替换

**文件：**
- 修改：`web/src/feature/situation/components/SituationDashboard.tsx`
- 修改：`web/src/feature/situation/components/AttackChainTimeline.tsx`
- 修改：`web/src/feature/situation/components/AttackerDrawer.tsx`
- 修改：`web/src/feature/situation/components/QuickActionToolbar.tsx`

**接口：**
- 旧 class `glass-card` / `glass-card-light` / `glass-card-emphasis` → 新 `surface-card` / `surface-modal`
- 硬编码 `text-white*` → CSS 变量 `var(--text-*)`
- 删除 `dark:text-shadow-glow-white`, `dark:icon-neon`

- [ ] **Step 1: 更新 `SituationDashboard.tsx`**

```tsx
// 将所有 "text-white/50" 替换为样式：
// style={{ color: 'var(--text-muted)' }}
// 
// 将所有 "text-white" 替换为：
// style={{ color: 'var(--text-primary)' }}
//
// 将所有 "text-white/40" 替换为：
// style={{ color: 'var(--text-dim)' }}
//
// 将所有 "text-white/60" 替换为：
// style={{ color: 'var(--text-secondary)' }}
//
// 将所有 "glass-card-light" 替换为 "surface-card"
// 将所有 "glass-card" 替换为 "surface-card"
// 删除所有 "dark:text-shadow-glow-white"
```

具体改动（用 Edit 逐段替换）：

StatCard 组件内：
- `className="glass-card-light hover-lift p-5 h-full"` → `className="surface-card p-5 h-full"`
- `<p className="text-xs font-medium text-white/50 mb-2">` → `<p className="text-xs font-medium mb-2" style={{color:'var(--text-muted)'}}>`
- `<div className="text-3xl font-bold font-mono text-white dark:text-shadow-glow-white">` → `<div className="text-3xl font-bold font-mono" style={{color:'var(--text-primary)'}}>`
- `<p className="mt-1 text-xs text-white/40">` → `<p className="mt-1 text-xs" style={{color:'var(--text-dim)'}}>`

OverviewStats 中的整体风险分卡片同理替换。
BreakdownSection 同理替换。

- [ ] **Step 2: 更新 `AttackChainTimeline.tsx`**

```tsx
// "glass-card p-5" → "surface-card p-5"
// "text-white" → style={{color:'var(--text-primary)'}}
// "text-white/40" → style={{color:'var(--text-dim)'}}
// "text-white/30" → style={{color:'var(--text-dim)'}}
// "text-white/50" → style={{color:'var(--text-muted)'}}
// "dark:text-shadow-glow-white" → 删除
// "bg-white/15" → style={{background:'var(--surface-root-border)'}}
// "bg-white/5" → hover:bg-black/5 (浅色通用)
// "border-[rgba(102,126,234,0.4)]" → 不变（stage 色标保留语义）
```

保留 STAGE_COLORS（攻击阶段色标有业务语义），但饱和度按 spec 降低（在常量里调整）。

- [ ] **Step 3: 更新 `AttackerDrawer.tsx`**

```tsx
// DialogContent 的 style 替换为 className="surface-modal" + 必要的 color:
<DialogContent className="surface-modal max-w-lg max-h-[90vh] p-0 gap-0 sm:rounded-lg" style={{color:'var(--text-primary)'}}>
  // 内部所有 "text-white" 替换为 style={{color:'var(--text-primary)'}}
  // "text-white/40" → style={{color:'var(--text-dim)'}}
  // "text-white/50" → style={{color:'var(--text-muted)'}}
  // "text-white/90" → style={{color:'var(--text-secondary)'}}
  // "text-white/80" → style={{color:'var(--text-secondary)'}}
  // "text-white/30" → style={{color:'var(--text-dim)'}}
  // "bg-white/5" → "bg-black/5 dark:bg-white/5"
  // "border-white/10" → "border-black/10 dark:border-white/10"
  // 去掉 "dark:text-shadow-glow-white"
</DialogContent>
```

- [ ] **Step 4: 更新 `QuickActionToolbar.tsx`**

```tsx
// 删除 "dark:icon-neon"
// <Shield className="h-3.5 w-3.5" />
// "text-white/50" → style={{color:'var(--text-muted)'}}
```

- [ ] **Step 5: 验证编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 6: 提交**

```bash
git add web/src/feature/situation/components/
git commit -m "refactor(ui): 态势感知组件 — surface-card/modal 替换 + 硬编码颜色 Token 化"
```

---

### Task 8: 告警/日志/规则页面样式替换

**文件：**
- 修改：`web/src/feature/alert/components/AlertStatsCards.tsx`
- 修改：`web/src/feature/alert/components/ChannelDialog.tsx`
- 修改：`web/src/feature/alert/components/RuleDialog.tsx`
- 修改：`web/src/feature/alert/components/HistoryDetailDialog.tsx`
- 修改：`web/src/feature/log/components/AttackDetailDialog.tsx`
- 修改：`web/src/feature/log/components/AttackLogFilter.tsx`
- 修改：`web/src/feature/log/components/AttackEventFilter.tsx`
- 修改：`web/src/feature/log/components/LokiSearch.tsx`
- 修改：`web/src/feature/log/components/LokiStatsPanel.tsx`

**接口：**
- Dialog 内联 style → `surface-modal` class
- 统计卡片 → `surface-card`
- 表格容器 → 实底白/灰 + 细边框

- [ ] **Step 1: 逐个文件创建更改**

每个文件执行以下模式：

**Dialog 文件** (ChannelDialog, RuleDialog, HistoryDetailDialog, AttackDetailDialog)：
```
查找 DialogContent 上的 style 属性
替换为：className="surface-modal ...(保留其他 className)" style={{color:'var(--text-primary)'}}
文本色硬编码替换同 Task 7 模式
```

**StatsCards 文件** (AlertStatsCards)：
```
确认 StatsCard 子组件的样式（可能不需要改动，取决于 StatsCard 的实现）
```

**Filter/Loki 组件**：
```
"glass-card" → "surface-card"
硬编码 text-white* → CSS 变量
"bg-white/5" → "bg-black/5 dark:bg-white/5"
"border-white/10" → "border-black/10 dark:border-white/10"
```

- [ ] **Step 2: 验证编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 3: 提交**

```bash
git add web/src/feature/alert/components/ web/src/feature/log/components/
git commit -m "refactor(ui): 告警/日志组件 — surface-card/modal + 颜色 Token 化"
```

---

### Task 9: 安全大屏 & 监控仪表盘样式替换

**文件：**
- 修改：`web/src/feature/security-dashboard/component/StatCard.tsx`
- 修改：`web/src/feature/security-dashboard/component/AttackIPList.tsx`
- 修改：`web/src/feature/security-dashboard/component/RealtimeAttackList.tsx`
- 修改：`web/src/feature/security-dashboard/component/DashboardQPSChart.tsx`
- 修改：`web/src/feature/monitor/components/MetricCard.tsx`
- 修改：`web/src/feature/monitor/components/StatusBar.tsx`
- 修改：`web/src/feature/stats/components/StatsCard.tsx`

**接口：**
- 所有卡片 → `surface-card`
- StatusBar 保留脉冲动画
- Globe3DMap 背景强制深色（路由级例外）

- [ ] **Step 1: 逐个文件替换 class 和颜色**

`StatCard.tsx` / `MetricCard.tsx` / `StatsCard.tsx`:
```
"glass-card*" → "surface-card"
"text-white" → style={{color:'var(--text-primary)'}}
"text-white/50" → style={{color:'var(--text-muted)'}}
"text-white/60" → style={{color:'var(--text-secondary)'}}
删除 "dark:text-shadow-glow-white"
数值字号：text-3xl 保持不变（已接近 data-xl 规范 24px）
```

`AttackIPList.tsx` / `RealtimeAttackList.tsx` / `DashboardQPSChart.tsx`:
```
容器 "glass-card*" → "surface-card"
内部表格行列色值替换
```

`StatusBar.tsx`:
```
保留 status-dot-online
容器改为实底不玻璃
```

- [ ] **Step 2: Globe3DMap 深色背景保留**

```tsx
// 确认 SecurityDashboardLayout.tsx 中性化——移除硬编码深色，由路由级控制
// 如果 Globe3D 需要深色背景，在 SecurityDashboardLayout 容器上加 dark class
```

- [ ] **Step 3: 验证编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 4: 提交**

```bash
git add web/src/feature/security-dashboard/ web/src/feature/monitor/ web/src/feature/stats/
git commit -m "refactor(ui): 安全大屏 & 监控仪表盘 — surface-card + Globe3D 深色保留"
```

---

### Task 10: AI 分析器组件样式替换

**文件：**
- 修改：`web/src/feature/ai-analyzer/components/AttackPatternTable.tsx`
- 修改：`web/src/feature/ai-analyzer/components/GeneratedRuleTable.tsx`
- 修改：`web/src/feature/ai-analyzer/components/PatternDetectionCard.tsx`
- 修改：`web/src/feature/ai-analyzer/components/RuleGenerationCard.tsx`
- 修改：`web/src/feature/ai-analyzer/components/AIConfigCard.tsx`
- 修改：`web/src/feature/ai-analyzer/components/AttackPatternDetailDialog.tsx`
- 修改：`web/src/feature/ai-analyzer/components/RuleDetailDialog.tsx`
- 修改：`web/src/feature/ai-analyzer/components/RuleReviewDialog.tsx`

**接口：**
- 卡片 → `surface-card`，Dialog → `surface-modal`
- AttackPatternTable 表头/行用实底色

- [ ] **Step 1: 表格组件实底色化**

`AttackPatternTable.tsx` / `GeneratedRuleTable.tsx`:
```
表格容器 → 透明底 + 细边框
表头行：bg-black/5 dark:bg-white/5
数据行：透明 hover:bg-black/5 dark:hover:bg-white/5
列标题：style={{color:'var(--text-muted)'}}
```

- [ ] **Step 2: 卡片和 Dialog 替换**

PatternDetectionCard / RuleGenerationCard / AIConfigCard:
```
"glass-card*" → "surface-card"
硬编码色 → CSS 变量
```

AttackPatternDetailDialog / RuleDetailDialog / RuleReviewDialog:
```
DialogContent style → className="surface-modal ..." + color token
```

- [ ] **Step 3: 验证编译 & 提交**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | head -20
git add web/src/feature/ai-analyzer/components/
git commit -m "refactor(ui): AI 分析器组件 — 表格实底化 + 卡片/Dialog Token 化"
```

---

### Task 11: 登录页 & 其余页面收尾

**文件：**
- 修改：`web/src/pages/auth/login.tsx`
- 修改：`web/src/pages/auth/reset-password.tsx`
- 修改：`web/src/pages/forbidden/page.tsx`
- 修改：`web/src/components/common/loading-fallback.tsx`
- 修改：所有剩余含 `text-white` 硬编码的页面文件

**接口：**
- 登录页保留渐变背景 + 浮动粒子
- features 卡片颜色统一品牌紫
- 所有 Dialog / 弹窗统一 surface-modal

- [ ] **Step 1: 登录页调整**

`login.tsx`:
```
// features 数组中的 color 从多色改为品牌紫系：
const features = [
  { icon: Shield, label: 'WAF Engine', value: 'Coraza 3.0', color: 'text-primary-400' },
  { icon: Activity, label: 'Threats Blocked', value: '12.4K', color: 'text-primary-300' },
  { icon: Globe, label: 'Protected Sites', value: '8', color: 'text-primary-300' },
  { icon: Server, label: 'Uptime', value: '99.9%', color: 'text-primary-300' },
];

// 登录表单卡片 borderRadius 从 glass-card-emphasis 改为更小的圆角
"glass-card-emphasis p-8" → className="surface-card p-8" style={{borderRadius:'var(--radius-lg)'}}
// 文本色改为 CSS 变量
"text-white" → style={{color:'var(--text-primary)'}}
"text-white/50" → style={{color:'var(--text-muted)'}}
"text-white/70" → style={{color:'var(--text-secondary)'}}
```

- [ ] **Step 2: 批量替换剩余文件**

遍历 `grep -rl 'text-white\|glass-card\|glass-nav\|glass-btn\|glass-input' web/src/` 中尚未处理的文件：
```
每个文件执行相同的替换模式：
- glass-* 类 → surface-* 类
- text-white → style={{color:'var(--text-primary)'}}
- text-white/30 等 → 对应的 CSS 变量
- 内联 background/backdropFilter → surface-* 类
```

- [ ] **Step 3: 验证编译 & 提交**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | head -20
git add web/src/pages/ web/src/components/common/
git commit -m "refactor(ui): 登录页 & 剩余页面收尾 — 颜色统一 + class 全量替换"
```

---

### Task 12: i18n 翻译键补充

**文件：**
- 修改：`web/src/i18n.ts`（或对应的翻译 JSON 文件）

**接口：**
- 新增翻译键：`system.status`, `sidebar.aiAssistant`, `auth.profile`
- 供 Header Bar 和 Page Header 使用

- [ ] **Step 1: 查找翻译文件位置**

```bash
find /Users/duheling/AI-Waf/web/src -name '*.json' -path '*/i18n/*' -o -name '*.json' -path '*/locales/*' | head -10
```

- [ ] **Step 2: 添加翻译键**

在翻译 JSON 中增加：
```json
{
  "system": {
    "status": "系统正常运行"
  },
  "sidebar": {
    "aiAssistant": "AI 智能助手"
  },
  "auth": {
    "profile": "个人设置"
  }
}
```

- [ ] **Step 3: 提交**

```bash
git add web/src/i18n* web/src/locales/
git commit -m "feat(i18n): 添加 Header Bar 相关翻译键"
```

---

### Task 13: 端到端验证

**目标：** 确保 dev server 正常启动，所有页面无白屏。

- [ ] **Step 1: 启动 dev server**

```bash
cd /Users/duheling/AI-Waf/web && npx vite --host 0.0.0.0 &
sleep 5
```

- [ ] **Step 2: 检查基本编译产物**

```bash
curl -s http://localhost:5173 | head -20
```

预期：返回 HTML 页面（含 `<div id="root">`）。

- [ ] **Step 3: 检查 TypeScript 编译**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit 2>&1 | tail -20
```

预期：无类型错误。

- [ ] **Step 4: 关闭 dev server**

```bash
kill %1 2>/dev/null
```

- [ ] **Step 5: 提交（如有遗漏修复）**

```bash
git add -A
git commit -m "chore(ui): 端到端验证修复 — 编译 & 类型检查通过"
```

---

### Task 14: 浅色模式验证 & 修复

**目标：** 在 localStorage 中手动设置 theme=light，验证浅色模式在各页面的表现。

- [ ] **Step 1: 启动 dev server 并设置浅色模式**

```bash
cd /Users/duheling/AI-Waf/web && npx vite --host 0.0.0.0 &
# 在浏览器 console 中执行：localStorage.setItem('theme-storage', JSON.stringify({state:{theme:'light'}}))
```

- [ ] **Step 2: 逐页面检查**

检查以下页面的浅色模式渲染：
1. 登录页（背景渐变应转为浅色柔和色）
2. 监控概览（统计卡片应白底半透）
3. 日志页（表格应白底）
4. 态势感知（卡片/Dialog 玻璃效果正常）
5. AI 分析器（表格表头/行可读）
6. 安全大屏（Globe3D 除外，保持深色）

- [ ] **Step 3: 修复浅色模式问题**

记录并修复发现的对齐问题（例如某些元素在浅色下缺少边框、文字看不清等）。

- [ ] **Step 4: 提交**

```bash
git add -A
git commit -m "fix(ui): 浅色模式验证修复 — 双主题兼容性调整"
```

---

## 任务依赖图

```
Task 1 (CSS Tokens)
  ├→ Task 2 (Tailwind 清理)
  ├→ Task 3 (Header Bar)
  ├→ Task 4 (Page Header)
  ├→ Task 5 (Sidebar)
  │    └→ Task 6 (RootLayout) ← 依赖 3,4,5
  ├→ Task 7 (态势感知)
  ├→ Task 8 (告警/日志)
  ├→ Task 9 (安全大屏/监控)
  ├→ Task 10 (AI 分析器)
  ├→ Task 11 (登录页 & 收尾)
  ├→ Task 12 (i18n)
  └→ Task 13 (端到端验证) ← 依赖所有
       └→ Task 14 (浅色模式修复)
```

**Task 3-5 可并行，Task 7-12 可并行，均依赖 Task 1 完成。**
