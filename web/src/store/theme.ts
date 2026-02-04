/**
 * Theme Store - 使用 Zustand 管理主题状态
 * 替代原有的 Context API 实现，提供更好的性能和开发体验
 */
import { create } from 'zustand'
import { persist, devtools } from 'zustand/middleware'

export type Theme = 'dark' | 'light' | 'system'
export type ResolvedTheme = 'dark' | 'light'

interface ThemeState {
    theme: Theme
    resolvedTheme: ResolvedTheme
    setTheme: (theme: Theme) => void
    initTheme: () => void
}

/**
 * 解析系统主题偏好
 */
const getSystemTheme = (): ResolvedTheme => {
    if (typeof window === 'undefined') return 'light'
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

/**
 * 应用主题到 DOM
 */
const applyTheme = (resolvedTheme: ResolvedTheme) => {
    const root = window.document.documentElement
    root.classList.remove('light', 'dark')
    root.classList.add(resolvedTheme)
}

/**
 * Theme Store
 * 使用 Zustand + persist 中间件实现主题状态管理
 */
export const useThemeStore = create<ThemeState>()(
    devtools(
        persist(
            (set, get) => ({
                theme: 'system',
                resolvedTheme: 'light',

                /**
                 * 设置主题
                 */
                setTheme: (theme: Theme) => {
                    let resolvedTheme: ResolvedTheme

                    if (theme === 'system') {
                        resolvedTheme = getSystemTheme()
                    } else {
                        resolvedTheme = theme
                    }

                    applyTheme(resolvedTheme)
                    set({ theme, resolvedTheme })
                },

                /**
                 * 初始化主题
                 * 在应用启动时调用，处理系统主题偏好监听
                 */
                initTheme: () => {
                    const { theme, setTheme } = get()
                    
                    // 应用当前主题
                    setTheme(theme)

                    // 监听系统主题变化
                    if (typeof window !== 'undefined') {
                        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
                        
                        const handleChange = () => {
                            const { theme } = get()
                            if (theme === 'system') {
                                const resolvedTheme = getSystemTheme()
                                applyTheme(resolvedTheme)
                                set({ resolvedTheme })
                            }
                        }

                        // 现代浏览器使用 addEventListener
                        if (mediaQuery.addEventListener) {
                            mediaQuery.addEventListener('change', handleChange)
                        } else {
                            // 旧版浏览器兼容
                            mediaQuery.addListener(handleChange)
                        }
                    }
                },
            }),
            {
                name: 'theme-storage',
                // 只持久化 theme，resolvedTheme 在初始化时计算
                partialize: (state) => ({ theme: state.theme }),
            }
        ),
        { name: 'ThemeStore' }
    )
)

// ==================== 选择器 Hooks ====================
// 提供细粒度的状态订阅，避免不必要的重渲染

/**
 * 获取当前主题设置
 */
export const useTheme = () => useThemeStore((state) => state.theme)

/**
 * 获取解析后的主题（实际应用的主题）
 */
export const useResolvedTheme = () => useThemeStore((state) => state.resolvedTheme)

/**
 * 获取设置主题的函数
 */
export const useSetTheme = () => useThemeStore((state) => state.setTheme)

/**
 * 获取完整的主题状态和操作
 */
export const useThemeActions = () => useThemeStore((state) => ({
    theme: state.theme,
    resolvedTheme: state.resolvedTheme,
    setTheme: state.setTheme,
}))
