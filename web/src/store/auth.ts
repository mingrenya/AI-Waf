/**
 * Auth Store - 使用 Zustand 管理认证状态
 * 支持状态持久化、DevTools 调试
 */
import { useMemo } from 'react'
import { create } from 'zustand'
import { persist, devtools } from 'zustand/middleware'
import { AuthState, User } from '@/types/auth'

/**
 * Auth Store
 * 管理用户认证状态、token、登录状态等
 */
export const useAuthStore = create<AuthState>()(
    devtools(
        persist(
            (set) => ({
                user: null,
                token: null,
                isAuthenticated: false,
                needPasswordReset: false,

                /**
                 * 登录
                 * @param token - 认证 token
                 * @param user - 用户信息
                 */
                login: (token: string, user: User) => {
                    set({
                        user,
                        token,
                        isAuthenticated: true,
                        needPasswordReset: user.needReset || false,
                    })
                },

                /**
                 * 登出
                 * 清除所有认证状态
                 */
                logout: () => {
                    set({
                        user: null,
                        token: null,
                        isAuthenticated: false,
                        needPasswordReset: false,
                    })
                },

                /**
                 * 更新用户信息
                 * @param user - 新的用户信息
                 */
                setUser: (user: User) => {
                    set({
                        user,
                        needPasswordReset: user.needReset || false,
                    })
                },
            }),
            {
                name: 'auth-storage',
                // 只持久化必要字段
                partialize: (state) => ({
                    token: state.token,
                    user: state.user,
                    isAuthenticated: state.isAuthenticated,
                    needPasswordReset: state.needPasswordReset,
                }),
            }
        ),
        { name: 'AuthStore' }
    )
)

// ==================== 选择器 Hooks ====================
// 提供细粒度的状态订阅，避免不必要的重渲染

/**
 * 获取当前用户信息
 */
export const useUser = () => useAuthStore((state) => state.user)

/**
 * 获取认证 token
 */
export const useToken = () => useAuthStore((state) => state.token)

/**
 * 获取认证状态
 */
export const useIsAuthenticated = () => useAuthStore((state) => state.isAuthenticated)

/**
 * 获取是否需要重置密码
 */
export const useNeedPasswordReset = () => useAuthStore((state) => state.needPasswordReset)

/**
 * 获取认证操作函数
 */
export const useAuthActions = () => {
    const login = useAuthStore((state) => state.login)
    const logout = useAuthStore((state) => state.logout)
    const setUser = useAuthStore((state) => state.setUser)

    return useMemo(
        () => ({
            login,
            logout,
            setUser,
        }),
        [login, logout, setUser]
    )
}

/**
 * 获取完整的认证状态和操作
 */
export const useAuth = () => {
    const user = useAuthStore((state) => state.user)
    const token = useAuthStore((state) => state.token)
    const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
    const needPasswordReset = useAuthStore((state) => state.needPasswordReset)
    const login = useAuthStore((state) => state.login)
    const logout = useAuthStore((state) => state.logout)
    const setUser = useAuthStore((state) => state.setUser)

    return useMemo(
        () => ({
            user,
            token,
            isAuthenticated,
            needPasswordReset,
            login,
            logout,
            setUser,
        }),
        [user, token, isAuthenticated, needPasswordReset, login, logout, setUser]
    )
}

export default useAuthStore 