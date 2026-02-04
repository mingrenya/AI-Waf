/**
 * Store 统一导出
 * 提供所有状态管理的单一入口
 */

// Auth Store
export {
    useAuthStore,
    useUser,
    useToken,
    useIsAuthenticated,
    useNeedPasswordReset,
    useAuthActions,
    useAuth,
} from './auth'

// Theme Store
export {
    useThemeStore,
    useTheme,
    useResolvedTheme,
    useSetTheme,
    useThemeActions,
    type Theme,
    type ResolvedTheme,
} from './theme'

// Toast Store
export {
    useToastStore,
    useToasts,
    useToastActions,
    toast,
    toastSuccess,
    toastError,
    toastInfo,
    dismissAllToasts,
    type Toast,
    type ToastVariant,
} from './toast'
