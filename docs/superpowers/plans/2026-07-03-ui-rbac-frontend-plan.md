# AI-Waf 前端 UI 重构 + 权限体系贯通 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将前端从当前半成品状态升级为专业级 WAF 控制台——基于 uitest 模板的设计语言重构 UI 样式/动效，同时将后端已有的四角色+24权限体系完整贯通到前端的路由守卫、侧边栏、页面和操作按钮。

**Architecture:** 设计层面采用「深色监控仪表盘」为主风格（融合 #31 Real-Time Monitoring + #39 Bento Box + #15 Motion-Driven），权限层面基于现有的 Zustand auth store 和 permissions.ts，增强 ProtectedRoute、Sidebar、各页面组件的权限感知能力，实现与后端 RBAC 体系的无缝对接。

**Tech Stack:** React 18 + TypeScript 5.6 + TailwindCSS 3.4 + shadcn/ui (Radix) + Motion (Framer Motion) 12.6 + Zustand 5.0 + React Router 7

**设计参考模板:** `/Users/duheling/Downloads/uitest-main/styles/`
- #31 Real-Time Monitoring — 深色监控仪表盘主基调
- #39 Bento Box — 非对称网格布局
- #15 Motion-Driven — 页面过渡与微交互
- #41 Cyberpunk — 状态指示灯/描边发光
- #43 AI-Native — AI 对话界面
- #03 Glassmorphism — 毛玻璃卡片叠加

## Global Constraints

- 所有后端 API 不变 — 仅前端代码修改，不触及 `server/` 目录
- 保持现有 TailwindCSS + shadcn/ui + Motion 技术栈，不引入新 UI 框架
- 颜色系统使用 CSS 变量定义，支持暗色/亮色双主题
- 权限判断统一使用 `@/lib/permissions` 中的 `hasPermission()` / `hasAnyPermission()`
- 用户状态从 `useAuthStore((state) => state.user)` 获取
- 侧边栏可见性通过 `displayConfig` prop 控制，在 `root-layout.tsx` 中根据用户权限动态计算
- 路由守卫在 `ProtectedRoute` 中增加权限检查
- 页面级权限：无权限访问的路由重定向到 403 页面
- 操作级权限：按钮/表单根据 `hasPermission` 条件渲染
- 文件路径相对 `/Users/duheling/AI-Waf/web/`
- 提交使用 Conventional Commits 格式，中文 commit message
- 每个 Task 结束需通过 `npx tsc --noEmit && npm run build` 验证

---

## File Structure

```
web/src/
├── styles/
│   └── theme.css                    # (Create) CSS 变量主题系统重构
├── tailwind.config.ts               # (Modify) 扩展动画/颜色/霓虹灯配置
├── index.css                        # (Modify) 全局样式重构
├── lib/
│   └── permissions.ts               # (Modify) 增强权限判断工具
├── feature/
│   └── auth/
│       └── components/
│           ├── ProtectedRoute.tsx   # (Modify) 增加路由级权限检查
│           └── RoleBasedRoute.tsx   # (Create) 基于角色/权限的路由组件
├── components/
│   ├── layout/
│   │   ├── root-layout.tsx          # (Modify) 根据用户权限动态计算 sidebar displayConfig
│   │   ├── sidebar.tsx              # (Modify) 重构样式(设计升级) + 接收权限配置
│   │   └── breadcrumb.tsx           # (Modify) 重构样式
│   └── common/
│       └── forbidden.tsx            # (Create) 403 无权限页面
├── pages/
│   ├── auth/
│   │   └── login.tsx                # (Modify) 重构登录页样式
│   ├── setting/
│   │   └── pages/
│   │       └── user/
│   │           └── page.tsx         # (Modify) 增强用户管理页
│   └── forbidden/
│       └── page.tsx                 # (Create) 403 页面路由
├── feature/
│   └── user-management/
│       └── components/
│           ├── UserManagementTable.tsx  # (Modify) 重构样式
│           └── UserDialog.tsx           # (Modify) 增加角色选择器说明
├── routes/
│   ├── config.tsx                   # (Modify) 增加权限感知路由 + 403 路由
│   └── constants.ts                 # (Modify) 添加 FORBIDDEN 路由常量
├── i18n/ (public/locales/)
│   ├── en/translation.json          # (Modify) 新增翻译键
│   └── zh/translation.json          # (Modify) 新增翻译键
```

---

## Phase 1: 设计系统重构（UI Foundation）

### Task 1: CSS 主题变量系统重构

**Files:**
- Modify: `web/src/index.css:1-80`（CSS 变量定义区域）
- Modify: `web/tailwind.config.ts:1-60`（动画扩展）

**Interfaces:**
- Consumes: 现有 CSS 变量体系（`:root` / `.dark`）
- Produces: 新的「深色监控仪表盘」主题变量系统，基于 uitest #31 + #39 的配色方案

**What:** 将当前紫色霓虹灯主题升级为专业的深色监控仪表盘风格。保留暗色模式的霓虹灯元素作为点缀，但整体基调从"花哨"转向"专业数据驱动"。

**设计决策:**
- **主色**: 从紫色渐变 → 暗色 slate-900/950 底色 + cyan/blue 数据高亮
- **字体**: Inter（UI 文本） + JetBrains Mono（数据/代码/monospace 数值）
- **卡片**: 深色半透明背景 + 细边框 + 微模糊
- **状态指示**: 绿色脉冲点（系统正常）/ 红色脉冲点（告警）
- **数据高亮**: cyan-400 / blue-400 / green-400 / yellow-400 / red-400
- **保留**: 现有暗色模式 CSS 变量结构，替换色值

- [ ] **Step 1: 更新 `web/src/index.css` 中的 CSS 变量**

将 `@layer base` 中的 `:root` 和 `.dark` 选择器的 CSS 变量替换为：

```css
@layer base {
  :root {
    /* 浅色模式 - 干净白色基调 */
    --background: 210 40% 98%;
    --foreground: 222 47% 11%;
    --card: 0 0% 100%;
    --card-foreground: 222 47% 11%;
    --popover: 0 0% 100%;
    --popover-foreground: 222 47% 11%;
    --primary: 199 89% 48%;
    --primary-foreground: 0 0% 100%;
    --secondary: 210 40% 96%;
    --secondary-foreground: 222 47% 11%;
    --muted: 210 40% 96%;
    --muted-foreground: 215 16% 47%;
    --accent: 199 89% 48%;
    --accent-foreground: 0 0% 100%;
    --destructive: 0 84% 60%;
    --destructive-foreground: 0 0% 100%;
    --success: 142 71% 45%;
    --success-foreground: 0 0% 100%;
    --warning: 38 92% 50%;
    --warning-foreground: 0 0% 100%;
    --border: 214 32% 91%;
    --input: 214 32% 91%;
    --ring: 199 89% 48%;
    --radius: 0.625rem;
    /* 图表色 */
    --chart-1: 199 89% 48%;
    --chart-2: 142 71% 45%;
    --chart-3: 38 92% 50%;
    --chart-4: 262 83% 58%;
    --chart-5: 0 84% 60%;
    /* 侧边栏 */
    --sidebar-bg: 222 47% 11%;
    --sidebar-fg: 210 40% 98%;
    --sidebar-accent: 199 89% 48%;
    --sidebar-border: 217 33% 17%;
    /* 状态指示 */
    --status-online: 142 71% 45%;
    --status-warning: 38 92% 50%;
    --status-critical: 0 84% 60%;
    /* 字体 */
    --font-sans: 'Inter', ui-sans-serif, system-ui;
    --font-mono: 'JetBrains Mono', ui-monospace, monospace;
  }

  .dark {
    --background: 222 47% 7%;
    --foreground: 210 40% 98%;
    --card: 217 33% 12%;
    --card-foreground: 210 40% 98%;
    --popover: 217 33% 12%;
    --popover-foreground: 210 40% 98%;
    --primary: 199 89% 48%;
    --primary-foreground: 0 0% 100%;
    --secondary: 217 33% 17%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217 33% 17%;
    --muted-foreground: 215 20% 65%;
    --accent: 199 89% 48%;
    --accent-foreground: 0 0% 100%;
    --destructive: 0 62% 50%;
    --destructive-foreground: 0 0% 100%;
    --success: 142 69% 40%;
    --success-foreground: 0 0% 100%;
    --warning: 38 92% 45%;
    --warning-foreground: 0 0% 100%;
    --border: 217 33% 20%;
    --input: 217 33% 20%;
    --ring: 199 89% 48%;
    --radius: 0.625rem;
    --chart-1: 199 89% 52%;
    --chart-2: 142 69% 44%;
    --chart-3: 38 92% 48%;
    --chart-4: 262 83% 62%;
    --chart-5: 0 62% 54%;
    --sidebar-bg: 222 47% 7%;
    --sidebar-fg: 210 40% 98%;
    --sidebar-accent: 199 89% 48%;
    --sidebar-border: 217 33% 17%;
    --status-online: 142 69% 40%;
    --status-warning: 38 92% 45%;
    --status-critical: 0 62% 50%;
  }
}
```

同时更新 `@layer base` 中的全局基础样式：

```css
@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
    font-family: 'Inter', ui-sans-serif, system-ui, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
  /* JetBrains Mono for data/metrics */
  .font-mono, [data-mono] {
    font-family: 'JetBrains Mono', ui-monospace, monospace;
    font-feature-settings: 'tnum' 1, 'cv01' 1;
  }
}
```

删除或缩减旧的霓虹灯/紫色发光 CSS 类（保留 `.scrollbar-custom`、`.scrollbar-none` 等实用工具类），移除 `.card-neon`、`.button-neon`、`.text-shadow-glow-purple`、`.text-shadow-glow-blue` 等样式。

- [ ] **Step 2: 更新 `web/tailwind.config.ts` 中的动画和扩展配置**

在 `tailwind.config.ts` 的 `theme.extend` 中，更新动画定义：

```typescript
theme: {
  extend: {
    // ... 保留现有 borderRadius, colors
    keyframes: {
      'accordion-down': { from: { height: '0' }, to: { height: 'var(--radix-accordion-content-height)' } },
      'accordion-up': { from: { height: 'var(--radix-accordion-content-height)' }, to: { height: '0' } },
      // 新增：监控仪表盘专用动画
      'fade-in': { '0%': { opacity: '0' }, '100%': { opacity: '1' } },
      'slide-up': { '0%': { opacity: '0', transform: 'translateY(16px)' }, '100%': { opacity: '1', transform: 'translateY(0)' } },
      'slide-in-right': { '0%': { opacity: '0', transform: 'translateX(16px)' }, '100%': { opacity: '1', transform: 'translateX(0)' } },
      'scale-in': { '0%': { opacity: '0', transform: 'scale(0.95)' }, '100%': { opacity: '1', transform: 'scale(1)' } },
      'pulse-green': { '0%, 100%': { opacity: '1' }, '50%': { opacity: '0.4' } },
      'pulse-red': { '0%, 100%': { opacity: '1' }, '50%': { opacity: '0.4' } },
      'ping-slow': { '0%': { transform: 'scale(1)', opacity: '1' }, '100%': { transform: 'scale(1.5)', opacity: '0' } },
      'glow-pulse': {
        '0%, 100%': { boxShadow: '0 0 4px var(--tw-shadow-color)' },
        '50%': { boxShadow: '0 0 12px var(--tw-shadow-color)' },
      },
      'number-tick': { '0%': { transform: 'translateY(8px)', opacity: '0' }, '100%': { transform: 'translateY(0)', opacity: '1' } },
      'border-pulse': {
        '0%, 100%': { borderColor: 'rgba(var(--primary), 0.3)' },
        '50%': { borderColor: 'rgba(var(--primary), 0.8)' },
      },
    },
    animation: {
      'accordion-down': 'accordion-down 0.2s ease-out',
      'accordion-up': 'accordion-up 0.2s ease-out',
      'fade-in': 'fade-in 0.4s ease-out',
      'slide-up': 'slide-up 0.5s ease-out',
      'slide-in-right': 'slide-in-right 0.4s ease-out',
      'scale-in': 'scale-in 0.3s ease-out',
      'pulse-green': 'pulse-green 2s ease-in-out infinite',
      'pulse-red': 'pulse-red 1.5s ease-in-out infinite',
      'ping-slow': 'ping-slow 2s ease-out infinite',
      'glow-pulse': 'glow-pulse 2s ease-in-out infinite',
      'number-tick': 'number-tick 0.3s ease-out',
      'border-pulse': 'border-pulse 3s ease-in-out infinite',
    },
    // 新增：字体家族
    fontFamily: {
      sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      mono: ['JetBrains Mono', 'ui-monospace', 'monospace'],
    },
  },
},
```

移除旧的 `aurora`、`float`、`float-reverse`、`gradient-shift`、`sidebar-neon`、`logo-pulse`、`neon-pulse`、`text-flicker`、`border-neon-flow` 等动画。移除旧的霓虹灯 Tailwind 插件（文本阴影/边框发光等）。

- [ ] **Step 3: 在 `web/index.html` 中添加 Google Fonts 引用**

在 `<head>` 中添加：

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
```

- [ ] **Step 4: 新增 shadcn 动画组件 `web/src/components/ui/animation/page-transition.tsx`**

创建一个可复用的页面进入动画包装组件，基于 Motion (Framer Motion)：

```typescript
'use client'

import { motion, type Variants } from 'motion/react'
import type { ReactNode } from 'react'

const pageVariants: Variants = {
  initial: { opacity: 0, y: 12 },
  animate: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.35, ease: [0.25, 0.46, 0.45, 0.94] },
  },
  exit: { opacity: 0, y: -8, transition: { duration: 0.2 } },
}

export function PageTransition({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <motion.div
      variants={pageVariants}
      initial="initial"
      animate="animate"
      exit="exit"
      className={className}
    >
      {children}
    </motion.div>
  )
}
```

创建 `web/src/components/ui/animation/stagger-list.tsx`：

```typescript
'use client'

import { motion, type Variants } from 'motion/react'
import type { ReactNode } from 'react'

const containerVariants: Variants = {
  initial: {},
  animate: { transition: { staggerChildren: 0.06, delayChildren: 0.05 } },
}

const itemVariants: Variants = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0, transition: { duration: 0.3 } },
}

export function StaggerList({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <motion.div variants={containerVariants} initial="initial" animate="animate" className={className}>
      {Array.isArray(children)
        ? children.map((child, i) => (
            <motion.div key={i} variants={itemVariants}>
              {child}
            </motion.div>
          ))
        : children}
    </motion.div>
  )
}
```

创建 `web/src/components/ui/animation/number-ticker.tsx`：

```typescript
'use client'

import { useEffect, useRef, useState } from 'react'
import { motion, useSpring, useTransform } from 'motion/react'

interface NumberTickerProps {
  value: number
  duration?: number
  className?: string
}

export function NumberTicker({ value, duration = 1.5, className }: NumberTickerProps) {
  const ref = useRef<HTMLSpanElement>(null)
  const spring = useSpring(0, { stiffness: 80, damping: 20, duration: duration * 1000 })

  useEffect(() => {
    spring.set(value)
  }, [spring, value])

  const display = useTransform(spring, (latest) => Math.round(latest).toLocaleString())

  return <motion.span ref={ref} className={className}>{display}</motion.span>
}
```

更新 `web/src/components/ui/animation/index.ts` 导出新增组件。

- [ ] **Step 5: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

Expected: 无 TypeScript 错误，Vite 构建成功。

- [ ] **Step 6: Commit**

```bash
git add web/src/index.css web/tailwind.config.ts web/index.html web/src/components/ui/animation/
git commit -m "feat(ui): 重构设计系统——深色监控仪表盘主题、CSS变量、动画系统升级

- 将紫色霓虹灯主题替换为专业的深色监控仪表盘风格(slate底色 + cyan数据高亮)
- 引入 Inter 字体(UI) + JetBrains Mono(数据)
- 新增8个仪表盘专用动画(fade-in/slide-up/number-tick/glow-pulse等)
- 新增 PageTransition/StaggerList/NumberTicker 动画组件
- 移除旧的霓虹灯紫色发光 CSS 类和动画

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: 登录页重构

**Files:**
- Modify: `web/src/pages/auth/login.tsx`
- Modify: `web/src/feature/auth/components/LoginForm.tsx`

**Interfaces:**
- Consumes: 现有 `useLogin()` hook, `LoginForm` 组件, 新的 CSS 变量主题
- Produces: 重构后的登录页，风格匹配深色监控仪表盘

**What:** 将登录页从当前的紫色极光渐变风格重构为专业深色仪表盘风格，参考 uitest #31 Real-Time Monitoring 的 dark 底色调。

**设计:**
- 背景: `bg-slate-950` 深色全屏
- 左侧: 产品介绍区域（Bento 网格展示 WAF 核心指标卡片）
- 右侧: 登录表单卡片（glassmorphism 效果）
- 顶部状态栏: "MRYa WAF System — Secure Access"
- 底部: 系统状态指示器

- [ ] **Step 1: 重构 `web/src/pages/auth/login.tsx`**

将文件替换为新的登录页布局：

```typescript
import { Suspense } from 'react'
import { LoginForm } from '@/feature/auth/components/LoginForm'
import { LoadingFallback } from '@/components/common/loading-fallback'
import { Shield, Activity, Globe, Server } from 'lucide-react'

// 装饰性 Bento 卡片数据
const features = [
  { icon: Shield, label: 'WAF Engine', value: 'Coraza 3.0', color: 'text-cyan-400' },
  { icon: Activity, label: 'Threats Blocked', value: '12.4K', color: 'text-green-400' },
  { icon: Globe, label: 'Protected Sites', value: '8', color: 'text-blue-400' },
  { icon: Server, label: 'Uptime', value: '99.9%', color: 'text-yellow-400' },
]

export default function LoginPage() {
  return (
    <div className="min-h-screen bg-slate-950 flex flex-col">
      {/* 状态栏 */}
      <div className="bg-emerald-600 text-white text-center py-1.5 text-sm font-medium flex items-center justify-center gap-2">
        <span className="w-2 h-2 bg-white rounded-full animate-pulse-green" />
        MRYa WAF System — Secure Access
      </div>

      <div className="flex-1 flex items-center justify-center p-6">
        <div className="w-full max-w-5xl grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* 左侧：产品展示 */}
          <div className="hidden lg:flex flex-col justify-center">
            <div className="mb-8">
              <h1 className="text-4xl font-bold text-white mb-3 font-mono tracking-tight">
                MRYa<span className="text-cyan-400">WAF</span>
              </h1>
              <p className="text-slate-400 text-lg">
                智能 Web 应用防火墙管理平台
              </p>
            </div>
            <div className="grid grid-cols-2 gap-4">
              {features.map((f) => (
                <div key={f.label} className="bg-slate-900 border border-slate-800 rounded-xl p-4 animate-fade-in">
                  <f.icon className={`w-5 h-5 ${f.color} mb-2`} />
                  <p className="text-xs text-slate-500 mb-1">{f.label}</p>
                  <p className={`text-xl font-bold font-mono ${f.color}`}>{f.value}</p>
                </div>
              ))}
            </div>
          </div>

          {/* 右侧：登录表单 */}
          <div className="flex items-center">
            <div className="w-full bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-8 shadow-2xl animate-scale-in">
              <div className="text-center mb-6">
                <h2 className="text-2xl font-bold text-white">登录</h2>
                <p className="text-slate-400 text-sm mt-1">输入凭据以访问控制台</p>
              </div>
              <Suspense fallback={<LoadingFallback />}>
                <LoginForm />
              </Suspense>
            </div>
          </div>
        </div>
      </div>

      {/* 底部状态 */}
      <div className="py-3 text-center text-xs text-slate-600 border-t border-slate-800">
        System Status: <span className="text-emerald-500">All Systems Operational</span>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 重构 `web/src/feature/auth/components/LoginForm.tsx`**

将表单样式更新为匹配新的深色仪表盘主题：

```typescript
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { User, LockKeyhole, LogIn } from 'lucide-react'
import { useLogin } from '../hooks'
import { loginSchema, type LoginFormData } from '@/validation/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Form, FormControl, FormField, FormItem, FormMessage } from '@/components/ui/form'

export function LoginForm() {
  const { login, isPending } = useLogin()

  const form = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: '', password: '' },
  })

  const onSubmit = (data: LoginFormData) => login(data)

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                  <Input
                    {...field}
                    placeholder="用户名"
                    autoComplete="username"
                    className="pl-10 bg-slate-800 border-slate-700 text-white placeholder:text-slate-500 h-11 rounded-lg focus-visible:ring-cyan-500"
                  />
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <div className="relative">
                  <LockKeyhole className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                  <Input
                    {...field}
                    type="password"
                    placeholder="密码"
                    autoComplete="current-password"
                    className="pl-10 bg-slate-800 border-slate-700 text-white placeholder:text-slate-500 h-11 rounded-lg focus-visible:ring-cyan-500"
                  />
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button
          type="submit"
          disabled={isPending}
          className="w-full h-11 bg-cyan-600 hover:bg-cyan-500 text-white font-medium rounded-lg transition-all duration-200 flex items-center justify-center gap-2"
        >
          {isPending ? (
            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
          ) : (
            <LogIn className="w-4 h-4" />
          )}
          {isPending ? '登录中...' : '登录'}
        </Button>
      </form>
    </Form>
  )
}
```

- [ ] **Step 3: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add web/src/pages/auth/login.tsx web/src/feature/auth/components/LoginForm.tsx
git commit -m "feat(ui): 重构登录页——深色仪表盘风格、Bento特性卡片、毛玻璃登录表单

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: 侧边栏重构

**Files:**
- Modify: `web/src/components/layout/sidebar.tsx`（完整重写）
- Modify: `web/src/components/layout/root-layout.tsx:5-25`

**Interfaces:**
- Consumes: `useAuthStore` (user), `hasPermission` from `@/lib/permissions`, 新的 CSS 变量
- Produces: 重构后的侧边栏（深色仪表盘风格） + `root-layout.tsx` 中根据用户权限动态计算 `displayConfig`

**What:** 将侧边栏从紫色霓虹灯风格升级为专业的深色仪表盘导航，同时将权限检查集成到 `root-layout.tsx` 中，使侧边栏菜单项根据用户角色/权限动态显示。

**侧边栏权限映射:**
| 菜单项 | 所需权限 | admin | configurator | auditor | user |
|--------|---------|:-----:|:------------:|:-------:|:----:|
| 监控 | `system:status` | ✅ | ✅ | ✅ | ✅ |
| 日志 | `waf:log:read` | ✅ | ✅ | ✅ | ✅ |
| 规则 | `config:read` | ✅ | ✅ | — | — |
| 告警 | `alert:channel:read` | ✅ | ✅ | ✅ | ✅ |
| 态势感知 | `waf:log:read` | ✅ | ✅ | ✅ | ✅ |
| Nuclei | `config:read` | ✅ | ✅ | — | — |
| AI 分析器 | `waf:log:read` | ✅ | ✅ | ✅ | ✅ |
| 设置 | `config:read` | ✅ | ✅ | — | — |
| 流量捕获 | `config:read` | ✅ | ✅ | — | — |

- [ ] **Step 1: 重写 `web/src/components/layout/sidebar.tsx`**

完整重写侧边栏组件，采用深色仪表盘导航风格（参考 uitest #28 Data-Dense Dashboard 侧边栏 + #31 Real-Time Monitoring 配色）：

```typescript
'use client'

import { Link, useLocation, useNavigate } from 'react-router'
import { cn } from '@/lib/utils'
import { Settings, Shield, BarChart2, FileText, LogOut, Bell, Brain, ScanSearch, Crosshair, Radio, ChevronLeft, ChevronRight } from 'lucide-react'
import { ROUTES } from '@/routes/constants'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/store/auth'
import { useState } from 'react'

interface SidebarItem {
  title: string
  icon: React.ComponentType<{ className?: string }>
  href: string
  requiredPermission: string
}

function createSidebarItems(t: (key: string) => string): SidebarItem[] {
  return [
    { title: t('sidebar.monitor'), icon: BarChart2, href: ROUTES.MONITOR, requiredPermission: 'system:status' },
    { title: t('sidebar.logs'), icon: FileText, href: ROUTES.LOGS, requiredPermission: 'waf:log:read' },
    { title: t('sidebar.rules'), icon: Shield, href: ROUTES.RULES, requiredPermission: 'config:read' },
    { title: t('sidebar.alerts'), icon: Bell, href: ROUTES.ALERTS, requiredPermission: 'alert:channel:read' },
    { title: t('sidebar.situation'), icon: ScanSearch, href: ROUTES.SITUATION, requiredPermission: 'waf:log:read' },
    { title: t('sidebar.nuclei'), icon: Crosshair, href: ROUTES.NUCLEI, requiredPermission: 'config:read' },
    { title: t('sidebar.capture'), icon: Radio, href: ROUTES.CAPTURE, requiredPermission: 'config:read' },
    { title: t('sidebar.aiAnalyzer'), icon: Brain, href: ROUTES.AI_ANALYZER, requiredPermission: 'waf:log:read' },
    { title: t('sidebar.settings'), icon: Settings, href: ROUTES.SETTINGS, requiredPermission: 'config:read' },
  ]
}

interface SidebarProps {
  allowedItems?: string[] // 允许显示的菜单项 href 列表
}

export function Sidebar({ allowedItems }: SidebarProps) {
  const location = useLocation()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { logout, user } = useAuthStore()
  const [collapsed, setCollapsed] = useState(false)

  const currentFirstLevelPath = '/' + location.pathname.split('/')[1]

  const items = createSidebarItems(t).filter((item) => {
    if (!allowedItems) return true
    return allowedItems.includes(item.href)
  })

  return (
    <aside
      className={cn(
        'h-full bg-slate-900 text-slate-300 flex flex-col border-r border-slate-800 transition-all duration-300 relative',
        collapsed ? 'w-16' : 'w-60',
      )}
    >
      {/* 折叠按钮 */}
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="absolute -right-3 top-16 w-6 h-6 bg-slate-800 border border-slate-700 rounded-full flex items-center justify-center hover:bg-slate-700 transition-colors z-10"
      >
        {collapsed ? <ChevronRight className="w-3 h-3" /> : <ChevronLeft className="w-3 h-3" />}
      </button>

      {/* Logo */}
      <div className={cn('flex items-center gap-3 py-5 border-b border-slate-800', collapsed ? 'px-3 justify-center' : 'px-5')}>
        <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center flex-shrink-0">
          <Shield className="w-5 h-5 text-white" />
        </div>
        {!collapsed && (
          <div>
            <p className="font-bold text-white text-sm tracking-tight">MRYa WAF</p>
            <p className="text-xs text-slate-500">Security Console</p>
          </div>
        )}
      </div>

      {/* 用户信息 */}
      {!collapsed && user && (
        <div className="px-5 py-3 border-b border-slate-800">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-full bg-slate-700 flex items-center justify-center text-xs font-medium text-cyan-400">
              {user.username.charAt(0).toUpperCase()}
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-sm text-white truncate">{user.username}</p>
              <p className="text-xs text-slate-500 truncate">{user.role}</p>
            </div>
          </div>
        </div>
      )}

      {/* 导航 */}
      <nav className="flex-1 py-3 space-y-0.5 overflow-y-auto">
        {items.map((item) => {
          const isActive = currentFirstLevelPath === item.href
          return (
            <Link
              key={item.href}
              to={item.href}
              title={collapsed ? item.title : undefined}
              className={cn(
                'flex items-center gap-3 font-medium transition-all duration-200 mx-2 rounded-lg',
                collapsed ? 'px-3 py-3 justify-center' : 'px-4 py-2.5',
                isActive
                  ? 'bg-cyan-500/10 text-cyan-400 border border-cyan-500/20'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800',
              )}
            >
              <item.icon className={cn('w-5 h-5 flex-shrink-0', isActive && 'text-cyan-400')} />
              {!collapsed && <span className="text-sm">{item.title}</span>}
              {isActive && !collapsed && (
                <div className="ml-auto w-1.5 h-1.5 rounded-full bg-cyan-400" />
              )}
            </Link>
          )
        })}
      </nav>

      {/* 底部操作 */}
      <div className={cn('border-t border-slate-800 py-3', collapsed ? 'px-3' : 'px-4')}>
        <button
          onClick={() => { logout(); navigate('/login') }}
          className={cn(
            'flex items-center gap-3 text-slate-400 hover:text-red-400 transition-colors duration-200 w-full rounded-lg hover:bg-slate-800',
            collapsed ? 'px-3 py-3 justify-center' : 'px-4 py-2.5',
          )}
          title={collapsed ? t('sidebar.logout') : undefined}
        >
          <LogOut className="w-5 h-5" />
          {!collapsed && <span className="text-sm">{t('sidebar.logout')}</span>}
        </button>
      </div>
    </aside>
  )
}
```

- [ ] **Step 2: 修改 `web/src/components/layout/root-layout.tsx`，根据用户权限动态计算侧边栏可见菜单**

```typescript
import { Outlet } from 'react-router'
import { Sidebar } from './sidebar'
import { Breadcrumb } from './breadcrumb'
import { useAuthStore } from '@/store/auth'
import { hasAnyPermission } from '@/lib/permissions'
import { ROUTES } from '@/routes/constants'
import { useMemo } from 'react'

export function RootLayout() {
  const user = useAuthStore((state) => state.user)

  // 根据用户权限计算可见的侧边栏菜单项
  const allowedSidebarItems = useMemo(() => {
    if (!user) return []

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
    ]

    return mapping
      .filter((m) => hasAnyPermission(user, m.permissions))
      .map((m) => m.route)
  }, [user])

  return (
    <div className="flex h-screen bg-slate-950">
      <Sidebar allowedItems={allowedSidebarItems} />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <Breadcrumb />
        <main className="flex-1 overflow-auto bg-slate-950 p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: 在 `ROUTES` 常量中添加 CAPTURE 路由（如果缺失）**

检查 `web/src/routes/constants.ts` 是否包含 `CAPTURE`。若无则添加：

```typescript
CAPTURE: '/capture',
```

- [ ] **Step 4: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

- [ ] **Step 5: Commit**

```bash
git add web/src/components/layout/sidebar.tsx web/src/components/layout/root-layout.tsx web/src/routes/constants.ts
git commit -m "feat(ui+auth): 重构侧边栏(深色仪表盘风格) + 基于用户权限动态过滤菜单

- 侧边栏从紫色霓虹灯改为深色 slate 仪表盘风格
- 新增折叠/展开功能，用户信息栏显示当前用户名和角色
- root-layout 根据用户角色/权限动态计算 allowedSidebarItems
- admin 可见全部菜单，user 仅可见监控/日志/告警/态势/AI

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Phase 2: 权限体系贯通（RBAC Frontend）

### Task 4: 增强路由守卫 — 路由级权限检查

**Files:**
- Create: `web/src/feature/auth/components/RoleBasedRoute.tsx`
- Modify: `web/src/feature/auth/components/ProtectedRoute.tsx`
- Create: `web/src/pages/forbidden/page.tsx`
- Modify: `web/src/routes/constants.ts`（添加 FORBIDDEN）
- Modify: `web/src/routes/config.tsx`（添加 403 路由）

**Interfaces:**
- Consumes: `hasPermission` from `@/lib/permissions`, `useAuthStore`
- Produces:
  - `RoleBasedRoute` — 接收 `requiredPermission` prop，无权限时渲染 `<ForbiddenPage />`
  - `ForbiddenPage` — 403 无权限页面
  - `ProtectedRoute` — 增强版，增加角色过期检查（用户被禁用/角色被移除）

- [ ] **Step 1: 创建 `web/src/pages/forbidden/page.tsx`**

```typescript
import { useNavigate } from 'react-router'
import { ShieldOff } from 'lucide-react'
import { Button } from '@/components/ui/button'

export default function ForbiddenPage() {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center">
      <div className="text-center animate-fade-in">
        <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-red-500/10 border border-red-500/20 flex items-center justify-center">
          <ShieldOff className="w-10 h-10 text-red-400" />
        </div>
        <h1 className="text-3xl font-bold text-white mb-3">403 — 无权限访问</h1>
        <p className="text-slate-400 mb-8 max-w-md">
          您当前的账户角色没有访问此页面的权限。如需提升权限，请联系系统管理员。
        </p>
        <div className="flex items-center gap-3 justify-center">
          <Button onClick={() => navigate(-1)} variant="outline" className="border-slate-700 text-slate-300 hover:bg-slate-800">
            返回上一页
          </Button>
          <Button onClick={() => navigate('/')} className="bg-cyan-600 hover:bg-cyan-500 text-white">
            前往首页
          </Button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 创建 `web/src/feature/auth/components/RoleBasedRoute.tsx`**

```typescript
import { Outlet } from 'react-router'
import { useAuthStore } from '@/store/auth'
import { hasPermission } from '@/lib/permissions'
import ForbiddenPage from '@/pages/forbidden/page'

interface RoleBasedRouteProps {
  requiredPermission?: string
}

/**
 * 基于角色/权限的路由守卫。
 * 当 `requiredPermission` 指定时，检查当前用户是否有该权限。
 * 若无权限，渲染 ForbiddenPage 而不是子路由。
 */
export function RoleBasedRoute({ requiredPermission }: RoleBasedRouteProps) {
  const user = useAuthStore((state) => state.user)

  if (requiredPermission && !hasPermission(user, requiredPermission)) {
    return <ForbiddenPage />
  }

  return <Outlet />
}
```

- [ ] **Step 3: 增强 `web/src/feature/auth/components/ProtectedRoute.tsx`**

```typescript
import { useLocation, Navigate, Outlet } from 'react-router'
import { useAuthStore } from '@/store/auth'

/**
 * 认证守卫组件。
 * 1. 未认证 → 重定向到 /login
 * 2. needPasswordReset → 重定向到 /reset-password
 * 3. 认证通过 → 渲染子路由
 */
export function ProtectedRoute() {
  const { isAuthenticated, needPasswordReset } = useAuthStore()
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (needPasswordReset && location.pathname !== '/reset-password') {
    return <Navigate to="/reset-password" replace />
  }

  return <Outlet />
}
```

（此文件改动较小，主要是确保与 `RoleBasedRoute` 协同工作。）

- [ ] **Step 4: 修改 `web/src/routes/constants.ts` 添加 FORBIDDEN 常量**

```typescript
export const ROUTES = {
  // ... 现有常量
  FORBIDDEN: '/forbidden',
} as const
```

- [ ] **Step 5: 修改 `web/src/routes/config.tsx` 添加 403 路由**

在 `authRoutes` 数组中添加：

```typescript
{ path: '/forbidden', element: lazyLoad(lazy(() => import('@/pages/forbidden/page'))) },
```

- [ ] **Step 6: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

- [ ] **Step 7: Commit**

```bash
git add web/src/feature/auth/components/RoleBasedRoute.tsx web/src/feature/auth/components/ProtectedRoute.tsx web/src/pages/forbidden/page.tsx web/src/routes/constants.ts web/src/routes/config.tsx
git commit -m "feat(auth): 增加路由级权限检查——RoleBasedRoute + 403页面

- 新增 RoleBasedRoute 组件，接收 requiredPermission prop 进行路由级权限过滤
- 无权限时展示专业的 403 页面，引导用户返回或联系管理员
- ProtectedRoute 保持不变，专注认证检查
- 路由配置中添加 /forbidden 路由

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: 面包屑导航重构 + 权限感知

**Files:**
- Modify: `web/src/components/layout/breadcrumb.tsx`
- Modify: `web/src/routes/config.tsx:79-170`（`createBreadcrumbConfig`）

**Interfaces:**
- Consumes: 现有的面包屑配置函数，Zustand auth store
- Produces: 权限感知的面包屑配置——某些面包屑项仅在用户有对应权限时可见

- [ ] **Step 1: 重构 `web/src/components/layout/breadcrumb.tsx`**

更新样式以匹配深色仪表盘主题：

```typescript
import { Link, useLocation } from 'react-router'
import { ChevronRight } from 'lucide-react'
import { type BreadcrumbConfig, type RoutePath, useBreadcrumbMap } from '@/routes/config'
import { cn } from '@/lib/utils'

export function Breadcrumb() {
  const location = useLocation()
  const breadcrumbMap = useBreadcrumbMap()

  const segments = location.pathname.split('/').filter(Boolean)
  if (segments.length === 0) return null

  const parentPath = ('/' + segments[0]) as RoutePath
  const config: BreadcrumbConfig | undefined = breadcrumbMap[parentPath]

  const currentItem = config?.items?.find((item) => item.path === segments[1])
  const parentLabel = getParentLabel(parentPath)

  return (
    <nav className="px-6 py-3 bg-slate-900/50 border-b border-slate-800 backdrop-blur-sm flex items-center gap-2 text-sm">
      <Link to={parentPath} className="text-slate-400 hover:text-slate-200 transition-colors">
        {parentLabel}
      </Link>
      {currentItem && (
        <>
          <ChevronRight className="w-3 h-3 text-slate-600" />
          <span className="text-cyan-400 font-medium">{currentItem.title}</span>
        </>
      )}
      {/* 实时时钟 */}
      <div className="ml-auto text-xs text-slate-600 font-mono">
        {new Date().toLocaleTimeString()}
      </div>
    </nav>
  )
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
  }
  return labels[path] ?? path
}
```

- [ ] **Step 2: 扩展 `createBreadcrumbConfig` 以接受更多权限参数**

修改 `web/src/routes/config.tsx` 中的 `createBreadcrumbConfig` 函数签名，增加更多权限参数：

```typescript
export function createBreadcrumbConfig(
  t: TFunction,
  canReadUsers: boolean,
  canManageConfig: boolean,
  canReadAlerts: boolean,
): Record<RoutePath, BreadcrumbConfig> {
  const settingsItems: BreadcrumbItem[] = [
    { title: t('breadcrumb.settings.settings'), path: "global", component: <GlobalSettingPage /> },
    { title: t('breadcrumb.settings.siteManager'), path: "site", component: <SiteManagerPage /> },
  ]

  if (canManageConfig) {
    settingsItems.push({ title: t('breadcrumb.settings.certManager'), path: "cert", component: <CertificatesPage /> })
  }
  if (canReadUsers) {
    settingsItems.push({ title: t('breadcrumb.settings.userManager'), path: 'user', component: <UserManagementPage /> })
  }

  // ... 其余配置不变
}

export function useBreadcrumbMap() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const canReadUsers = hasPermission(user, 'user:read')
  const canManageConfig = hasPermission(user, 'config:update')
  const canReadAlerts = hasPermission(user, 'alert:channel:read')
  return createBreadcrumbConfig(t, canReadUsers, canManageConfig, canReadAlerts)
}
```

- [ ] **Step 3: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/layout/breadcrumb.tsx web/src/routes/config.tsx
git commit -m "feat(ui+auth): 重构面包屑导航(深色风格) + 基于权限过滤面包屑项

- 面包屑样式匹配深色仪表盘主题
- 新增实时时钟显示
- createBreadcrumbConfig 接受更细粒度的权限参数
- 证书管理仅 configurator/admin 可见，用户管理仅 admin 可见

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: 用户管理页面增强

**Files:**
- Modify: `web/src/feature/user-management/components/UserManagementTable.tsx`
- Modify: `web/src/feature/user-management/components/UserDialog.tsx`
- Create: `web/src/pages/forbidden/page.tsx`（已在 Task 4 创建）

**Interfaces:**
- Consumes: `userManagementApi`, `useAuthStore`, `hasPermission`
- Produces: 增强的用户管理表格 + 创建/编辑对话框

**What:** 完善用户管理页面的权限控制（按钮级），重构样式匹配新设计系统，增加角色描述文案。

- [ ] **Step 1: 更新 `web/src/feature/user-management/components/UserManagementTable.tsx`**

主要改动：更新 UI 样式（卡片背景、表格行样式、角色徽章颜色改为匹配新主题色），确保权限控制逻辑完整：

关键样式修改（伪代码描述，实际实现为完整组件）：
- 卡片容器：`bg-slate-900 border border-slate-800 rounded-xl`（替换白色卡片）
- 表头：`bg-slate-800 text-slate-400`（替换默认表头）
- 行悬停：`hover:bg-slate-800/50`（替换默认悬停）
- 角色徽章颜色映射更新：
  ```typescript
  function roleBadgeClass(role: ManagedUser['role']) {
    switch (role) {
      case 'admin': return 'bg-red-500/10 text-red-400 border-red-500/20'
      case 'auditor': return 'bg-blue-500/10 text-blue-400 border-blue-500/20'
      case 'configurator': return 'bg-purple-500/10 text-purple-400 border-purple-500/20'
      default: return 'bg-slate-500/10 text-slate-400 border-slate-500/20'
    }
  }
  ```
- 操作按钮权限控制已实现（`canCreate/canUpdate/canDelete`），保持不变

- [ ] **Step 2: 更新 `web/src/feature/user-management/components/UserDialog.tsx`**

增加角色选择的描述文案，帮助管理员理解每个角色的含义：

```typescript
const roleDescriptions: Record<string, string> = {
  admin: '超级管理员 — 拥有所有权限，可管理用户、站点、配置、告警等全部功能',
  configurator: '配置管理员 — 可管理站点、证书、规则、告警配置，但不能管理用户',
  auditor: '审计员 — 只读访问所有日志、配置、用户信息，不能进行任何修改',
  user: '普通用户 — 仅可查看 WAF 日志和系统状态，权限最小',
}
```

在角色选择 `<Select>` 下方增加辅助描述文本：
```typescript
{selectedRole && (
  <p className="text-xs text-slate-500 mt-1">{roleDescriptions[selectedRole]}</p>
)}
```

样式更新：对话框背景 `bg-slate-900 border-slate-800`，输入框 `bg-slate-800 border-slate-700 text-white`。

- [ ] **Step 3: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add web/src/feature/user-management/components/
git commit -m "feat(ui+auth): 增强用户管理页面——深色仪表盘样式 + 角色描述 + 按钮级权限

- UserManagementTable/CreateDialog/DeleteDialog 样式匹配深色主题
- 角色选择器增加详细描述（管理员/配置管理员/审计员/普通用户）
- canCreate/canUpdate/canDelete 权限控制操作按钮可见性

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Phase 3: 核心页面样式升级（Bento Grid + 仪表盘风格）

### Task 7: 监控仪表盘页面重构

**Files:**
- Modify: `web/src/pages/monitor/pages/stats/page.tsx`
- Create: `web/src/feature/monitor/components/MetricCard.tsx`
- Create: `web/src/feature/monitor/components/StatusBar.tsx`

**Interfaces:**
- Consumes: 现有 stats API，新的动画组件（`NumberTicker`, `PageTransition`, `StaggerList`）
- Produces: Bento Grid 布局的监控仪表盘，参考 uitest #39 Bento Box + #31 Real-Time Monitoring

**What:** 将监控页面重构为 Bento Grid 非对称布局，核心指标使用 `NumberTicker` 数字滚动动画，实时数据获取使用 React Query 的 `refetchInterval`。

**设计布局：**
```
┌──────────────────────────────────────────────────────┐
│  [StatusBar] 系统状态条                                │
├──────────┬──────────┬──────────┬──────────────────────┤
│ Requests │ Blocked  │ Threats  │   QPS Sparkline      │
│  12.4K   │  1,847   │   342    │   (Recharts)         │
│          │          │          │                      │
├──────────┴──────────┴──────────┼──────────────────────┤
│   Attack Type Distribution     │   Geo Distribution   │
│   (Pie/Donut Chart)            │   (World Map/Sunburst)│
│                                │                      │
├────────────────────────────────┼──────────────────────┤
│   Top Attacker IPs (Table)     │   Recent Alerts      │
│                                │   (Timeline List)    │
└────────────────────────────────┴──────────────────────┘
```

- [ ] **Step 1: 创建 `web/src/feature/monitor/components/MetricCard.tsx`**

```typescript
import { cn } from '@/lib/utils'
import { NumberTicker } from '@/components/ui/animation/number-ticker'
import type { LucideIcon } from 'lucide-react'

interface MetricCardProps {
  label: string
  value: number
  suffix?: string
  trend?: { value: number; isPositive: boolean }
  icon: LucideIcon
  color?: 'cyan' | 'green' | 'yellow' | 'red' | 'blue' | 'purple'
  className?: string
}

const colorMap = {
  cyan: { text: 'text-cyan-400', bg: 'bg-cyan-500/10', border: 'border-cyan-500/20', dot: 'bg-cyan-400' },
  green: { text: 'text-green-400', bg: 'bg-green-500/10', border: 'border-green-500/20', dot: 'bg-green-400' },
  yellow: { text: 'text-yellow-400', bg: 'bg-yellow-500/10', border: 'border-yellow-500/20', dot: 'bg-yellow-400' },
  red: { text: 'text-red-400', bg: 'bg-red-500/10', border: 'border-red-500/20', dot: 'bg-red-400' },
  blue: { text: 'text-blue-400', bg: 'bg-blue-500/10', border: 'border-blue-500/20', dot: 'bg-blue-400' },
  purple: { text: 'text-purple-400', bg: 'bg-purple-500/10', border: 'border-purple-500/20', dot: 'bg-purple-400' },
}

export function MetricCard({ label, value, suffix, trend, icon: Icon, color = 'cyan', className }: MetricCardProps) {
  const c = colorMap[color]

  return (
    <div className={cn('bg-slate-900 border border-slate-800 rounded-xl p-5 hover:border-slate-700 transition-all duration-300', className)}>
      <div className="flex items-start justify-between mb-3">
        <span className="text-xs font-medium text-slate-500 uppercase tracking-wider">{label}</span>
        <div className={cn('w-8 h-8 rounded-lg flex items-center justify-center', c.bg, c.border)}>
          <Icon className={cn('w-4 h-4', c.text)} />
        </div>
      </div>
      <div className={cn('text-3xl font-bold font-mono tracking-tight', c.text)}>
        <NumberTicker value={value} />
        {suffix && <span className="text-lg ml-0.5">{suffix}</span>}
      </div>
      {trend && (
        <div className="flex items-center gap-1 mt-2">
          <span className={cn('text-xs font-medium', trend.isPositive ? 'text-green-400' : 'text-red-400')}>
            {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
          </span>
          <span className="text-xs text-slate-600">vs last period</span>
        </div>
      )}
      {/* 底部发光条 */}
      <div className={cn('h-0.5 rounded-full mt-3', c.bg)}>
        <div className={cn('h-full rounded-full', c.bg)} style={{ width: '60%' }} />
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 创建 `web/src/feature/monitor/components/StatusBar.tsx`**

```typescript
import { Wifi, WifiOff } from 'lucide-react'
import { cn } from '@/lib/utils'

interface StatusBarProps {
  online?: boolean
  lastSync?: string
}

export function StatusBar({ online = true, lastSync }: StatusBarProps) {
  return (
    <div className={cn(
      'flex items-center justify-between px-4 py-2 rounded-lg text-sm mb-4 border',
      online ? 'bg-emerald-500/5 border-emerald-500/20 text-emerald-400' : 'bg-red-500/5 border-red-500/20 text-red-400',
    )}>
      <div className="flex items-center gap-2">
        {online ? (
          <>
            <span className="relative flex h-2 w-2">
              <span className="animate-ping-slow absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
            </span>
            <span className="font-medium">All Systems Operational</span>
          </>
        ) : (
          <>
            <WifiOff className="w-4 h-4" />
            <span className="font-medium">System Degraded</span>
          </>
        )}
      </div>
      {lastSync && (
        <span className="text-xs font-mono text-slate-500">
          Last sync: {lastSync}
        </span>
      )}
    </div>
  )
}
```

- [ ] **Step 3: 重构 `web/src/pages/monitor/pages/stats/page.tsx` 为 Bento Grid 布局**

完整重写 StatsPage，使用 Bento Grid 布局整合 MetricCard、StatusBar 和现有图表：

```typescript
import { useQuery } from '@tanstack/react-query'
import { Activity, ShieldAlert, Globe, Server, Zap, AlertTriangle } from 'lucide-react'
import { statsApi } from '@/api/stats'
import { PageTransition } from '@/components/ui/animation/page-transition'
import { StaggerList } from '@/components/ui/animation/stagger-list'
import { MetricCard } from '@/feature/monitor/components/MetricCard'
import { StatusBar } from '@/feature/monitor/components/StatusBar'

export default function StatsPage() {
  const { data: overview } = useQuery({
    queryKey: ['stats', 'overview'],
    queryFn: () => statsApi.getOverview(),
    refetchInterval: 15000, // 每15秒刷新
  })

  return (
    <PageTransition className="space-y-6">
      <StatusBar lastSync="2s ago" />

      {/* Bento Grid 布局 */}
      <StaggerList className="grid grid-cols-4 grid-rows-2 gap-4 h-[480px]">
        {/* 主 KPI: Requests - 横跨2列 */}
        <div className="col-span-2 row-span-1">
          <MetricCard label="Total Requests" value={12847} suffix="" trend={{ value: 12.5, isPositive: true }} icon={Activity} color="cyan" />
        </div>

        {/* Blocked - 1列 */}
        <div>
          <MetricCard label="Blocked" value={1847} trend={{ value: 8.3, isPositive: true }} icon={ShieldAlert} color="red" />
        </div>

        {/* Threats - 1列 */}
        <div>
          <MetricCard label="Active Threats" value={342} trend={{ value: 3.1, isPositive: false }} icon={AlertTriangle} color="yellow" />
        </div>

        {/* 图表区域: QPS - 横跨3列 */}
        <div className="col-span-3 bg-slate-900 border border-slate-800 rounded-xl p-5">
          <p className="text-xs font-medium text-slate-500 uppercase tracking-wider mb-3">QPS Timeline</p>
          {/* 复用现有 Recharts QPS 图表组件 */}
        </div>

        {/* Sites - 1列 */}
        <div>
          <MetricCard label="Protected Sites" value={8} icon={Globe} color="blue" />
        </div>
      </StaggerList>

      {/* 下半部分: 表格 + 列表 */}
      <div className="grid grid-cols-2 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-5">
          <p className="text-xs font-medium text-slate-500 uppercase tracking-wider mb-3">Top Attacker IPs</p>
          {/* Top-N 攻击者 IP 表格 */}
        </div>
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-5">
          <p className="text-xs font-medium text-slate-500 uppercase tracking-wider mb-3">Recent Alerts</p>
          {/* 最近告警时间线列表 */}
        </div>
      </div>
    </PageTransition>
  )
}
```

（注：实际的 API 调用和图表组件需根据现有代码结构调整。核心改动是布局 + MetricCard + StatusBar。）

- [ ] **Step 4: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/monitor/pages/stats/page.tsx web/src/feature/monitor/
git commit -m "feat(ui): 监控仪表盘 Bento Grid 重构——MetricCard数字动画 + StatusBar + 实时刷新

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 8: 日志/规则/告警页面样式统一

**Files:**
- Modify: `web/src/pages/logs/pages/event/page.tsx`（攻击日志页面）
- Modify: `web/src/feature/log/components/AttackDetailDialog.tsx`（攻击详情弹窗）
- Modify: `web/src/pages/alert/pages/channel/page.tsx`（告警渠道页面头）
- Modify: `web/src/pages/rule/pages/micro-rule/page.tsx`（微规则页面头）

**Interfaces:**
- Consumes: 现有各页面的 API hooks，新的 CSS 主题
- Produces: 统一外观（深色卡片、表格、按钮风格一致性）

**What:** 将日志/规则/告警等核心操作页面的容器、卡片、表格统一样式为深色仪表盘主题。在关键交互节点（AttackDetailDialog）增加权限检查。

- [ ] **Step 1: 攻击日志页面容器样式统一**

在 `web/src/pages/logs/pages/event/page.tsx` 中：
- 页面容器添加 `PageTransition` 包装
- 表格容器改为 `bg-slate-900 border border-slate-800 rounded-xl`
- 筛选工具栏改为深色风格

- [ ] **Step 2: 攻击详情弹窗增加权限检查**

在 `web/src/feature/log/components/AttackDetailDialog.tsx` 中，快速处置按钮组增加权限检查：

```typescript
const canBlock = hasPermission(user, 'config:update')
// ...
{canBlock && (
  <Button onClick={handleQuickBlock} className="bg-red-600 hover:bg-red-500">
    立即封禁
  </Button>
)}
```

- [ ] **Step 3: 告警渠道/规则页面容器样式统一**

应用与日志页面相同的容器样式：`bg-slate-900 border border-slate-800 rounded-xl`。

- [ ] **Step 4: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/logs/ web/src/pages/alert/ web/src/pages/rule/ web/src/feature/log/
git commit -m "feat(ui+auth): 日志/规则/告警页面样式统一 + 攻击详情增加权限控制

- 所有操作页面容器统一为深色仪表盘风格
- AttackDetailDialog 快速处置按钮增加 config:update 权限检查
- 表格/筛选栏/卡片样式一致性

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Phase 4: 态势感知页面升级 + 收尾

### Task 9: 态势感知大屏样式升级

**Files:**
- Modify: `web/src/pages/situation/layout.tsx`
- Modify: `web/src/pages/situation/page.tsx`
- Modify: `web/src/feature/situation/components/SituationDashboard.tsx`
- Modify: `web/src/feature/situation/components/AttackChainTimeline.tsx`

**Interfaces:**
- Consumes: 现有态势感知 API + WebSocket hooks，新的 UI 组件
- Produces: Bento Grid 布局的态势大屏，专业数据可视化风格

**What:** 态势感知是 WAF 控制台的核心页面，需要最高级别的视觉打磨。应用 Bento Grid 布局 + MetricCard 组件 + AttackChainTimeline 深色风格重绘。

- [ ] **Step 1: 态势大屏主页面采用全屏 Bento 布局**

- 顶部：StatusBar（系统状态条）
- 左侧 60%：攻击链时间线（MITRE 阶段着色，深色背景）
- 右侧 40%：攻击者排行 + 实时攻击事件流
- 底部：态势指标卡片行

- [ ] **Step 2: 攻击链时间线深色风格重绘**

MITRE ATT&CK 阶段着色（深色背景适配）：
- Reconnaissance: 蓝灰
- Scanning: 黄
- Exploitation: 红
- Lateral Movement: 橙
- C2: 紫
- Exfiltration: 深红

- [ ] **Step 3: 验证构建**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add web/src/pages/situation/ web/src/feature/situation/
git commit -m "feat(ui): 态势感知大屏 Bento Grid 升级——MITRE攻击链深色重绘 + 实时事件流

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 10: 全局页面过渡动画 + 最终验证

**Files:**
- Modify: `web/src/components/layout/root-layout.tsx:15-25`（在 Outlet 外包装 AnimatePresence）
- Modify: `web/src/components/layout/animated-route.tsx`（更新或创建）
- Modify: `web/src/pages/auth/login.tsx`（微调动画）

**Interfaces:**
- Consumes: Motion (Framer Motion) `AnimatePresence`
- Produces: 全局路由切换过渡动画

**What:** 使用 Motion 的 `AnimatePresence` 在路由切换时实现平滑的淡入淡出过渡效果。

- [ ] **Step 1: 在 root-layout 中添加 AnimatePresence**

```typescript
import { AnimatePresence } from 'motion/react'
import { useLocation, Outlet } from 'react-router'
// ...

export function RootLayout() {
  const location = useLocation()
  // ...

  return (
    <div className="flex h-screen bg-slate-950">
      <Sidebar allowedItems={allowedSidebarItems} />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <Breadcrumb />
        <main className="flex-1 overflow-auto bg-slate-950 p-6">
          <AnimatePresence mode="wait">
            <Outlet key={location.pathname} />
          </AnimatePresence>
        </main>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 全量构建验证**

```bash
cd /Users/duheling/AI-Waf/web && npx tsc --noEmit && npm run build
```

Expected: 零 TypeScript 错误，Vite 构建成功。

- [ ] **Step 3: Commit**

```bash
git add web/src/components/layout/root-layout.tsx
git commit -m "feat(ui): 全局路由过渡动画——AnimatePresence 页面切换淡入淡出

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## 验证清单（所有 Task 完成后）

- [ ] `cd /Users/duheling/AI-Waf/web && npx tsc --noEmit` — 零 TypeScript 错误
- [ ] `cd /Users/duheling/AI-Waf/web && npm run build` — Vite 构建成功
- [ ] `grep -rn 'TODO\|FIXME\|HACK' web/src/feature/auth/ web/src/lib/permissions.ts web/src/components/layout/` — 零匹配
- [ ] 侧边栏根据用户角色正确过滤菜单项（admin 全可见, user 仅监控/日志/告警/态势/AI）
- [ ] 面包屑中证书管理仅 configurator/admin 可见
- [ ] 用户管理页面仅 admin 可见
- [ ] 403 页面在无权限访问时正确展示
- [ ] 攻击详情弹窗中快速处置按钮仅 config:update 权限可见
- [ ] 暗色/亮色主题切换正常，所有页面在新 CSS 变量下渲染正确
- [ ] 页面切换有平滑的 fade-out → fade-in 过渡动画
- [ ] 数字滚动动画（NumberTicker）在监控页正确执行
- [ ] StaggerList 交错入场动画在列表页正确执行
