/**
 * use-toast 兼容层
 * 为了向后兼容，保留原有的 use-toast hook 接口
 * 内部使用新的 Zustand Toast Store
 */
import { useToasts, useToastActions, toast } from '@/store'

/**
 * 兼容旧版 useToast hook
 * @deprecated 建议直接使用 @/store 中的 useToasts 和 toast 函数
 */
export function useToast() {
    const toasts = useToasts()
    const { dismissToast } = useToastActions()

    return {
        toasts,
        toast,
        dismiss: dismissToast,
    }
}

// 重新导出 toast 函数供直接使用
export { toast }
