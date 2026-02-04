/**
 * Toast Store - 使用 Zustand 管理 Toast 通知状态
 * 替代原有的自定义 reducer 实现，提供更简洁和高性能的解决方案
 */
import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

export type ToastVariant = 'default' | 'destructive' | 'success'

export interface Toast {
    id: string
    title?: string
    description?: string
    variant?: ToastVariant
    duration?: number
    action?: React.ReactNode
}

interface ToastState {
    toasts: Toast[]
    addToast: (toast: Omit<Toast, 'id'>) => string
    removeToast: (id: string) => void
    dismissToast: (id: string) => void
    clearAll: () => void
}

const TOAST_LIMIT = 3
const DEFAULT_DURATION = 3000

/**
 * 生成唯一 ID
 */
const generateId = () => {
    return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
}

/**
 * Toast Store
 * 使用 Zustand 实现简洁的 Toast 状态管理
 */
export const useToastStore = create<ToastState>()(
    devtools(
        (set, get) => ({
            toasts: [],

            /**
             * 添加新的 Toast
             * @returns Toast ID，可用于后续移除
             */
            addToast: (toast) => {
                const id = generateId()
                const duration = toast.duration ?? DEFAULT_DURATION

                // 添加到队列，限制数量
                set((state) => ({
                    toasts: [
                        ...state.toasts,
                        { ...toast, id }
                    ].slice(-TOAST_LIMIT)
                }))

                // 自动移除（除非 duration 为 Infinity）
                if (duration !== Infinity) {
                    setTimeout(() => {
                        get().removeToast(id)
                    }, duration)
                }

                return id
            },

            /**
             * 移除指定 Toast
             */
            removeToast: (id) => {
                set((state) => ({
                    toasts: state.toasts.filter((toast) => toast.id !== id)
                }))
            },

            /**
             * 关闭 Toast（与 removeToast 相同，提供语义化接口）
             */
            dismissToast: (id) => {
                get().removeToast(id)
            },

            /**
             * 清除所有 Toast
             */
            clearAll: () => {
                set({ toasts: [] })
            },
        }),
        { name: 'ToastStore' }
    )
)

// ==================== 选择器 Hooks ====================

/**
 * 获取所有 Toast
 */
export const useToasts = () => useToastStore((state) => state.toasts)

/**
 * 获取 Toast 操作函数
 */
export const useToastActions = () => useToastStore((state) => ({
    addToast: state.addToast,
    removeToast: state.removeToast,
    dismissToast: state.dismissToast,
    clearAll: state.clearAll,
}))

// ==================== 便捷函数 ====================

/**
 * 便捷的 toast 函数，可直接调用
 * @example
 * toast({ title: '成功', description: '操作完成', variant: 'success' })
 */
export const toast = (options: Omit<Toast, 'id'>) => {
    return useToastStore.getState().addToast(options)
}

/**
 * 成功提示的便捷函数
 */
export const toastSuccess = (title: string, description?: string) => {
    return toast({
        title,
        description,
        variant: 'success',
    })
}

/**
 * 错误提示的便捷函数
 */
export const toastError = (title: string, description?: string) => {
    return toast({
        title,
        description,
        variant: 'destructive',
    })
}

/**
 * 默认提示的便捷函数
 */
export const toastInfo = (title: string, description?: string) => {
    return toast({
        title,
        description,
        variant: 'default',
    })
}

/**
 * 关闭所有 Toast
 */
export const dismissAllToasts = () => {
    useToastStore.getState().clearAll()
}
