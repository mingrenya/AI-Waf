export const BASE_PATH = '/' as const

export const ROUTES = {
    LOGS: "/logs",
    MONITOR: "/monitor",
    RULES: "/rules",
    SETTINGS: "/settings",
} as const

export type RouteKey = keyof typeof ROUTES
export type RoutePath = typeof ROUTES[RouteKey]

// 辅助函数，用于类型安全地获取路由
export const getRoute = (key: RouteKey): RoutePath => ROUTES[key] 