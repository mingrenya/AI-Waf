/**
 * 统一的 React Query 查询键工厂
 * 提供类型安全的查询键管理,避免字符串拼写错误
 * 遵循 TanStack Query 最佳实践
 */

export const queryKeys = {
    // ==================== 认证相关 ====================
    auth: {
        all: ['auth'] as const,
        currentUser: () => ['auth', 'currentUser'] as const,
    },

    // ==================== 统计数据相关 ====================
    stats: {
        all: ['stats'] as const,
        overview: (timeRange: string) => ['stats', 'overview', timeRange] as const,
        realtimeQPS: (limit: number) => ['stats', 'realtimeQPS', limit] as const,
        trafficTimeSeries: (timeRange: string) => ['stats', 'trafficTimeSeries', timeRange] as const,
        combinedTimeSeries: (timeRange: string) => ['stats', 'combinedTimeSeries', timeRange] as const,
    },

    // ==================== 告警相关 ====================
    alert: {
        all: ['alert'] as const,
        channels: {
            all: ['alert', 'channels'] as const,
            lists: () => ['alert', 'channels', 'list'] as const,
            list: (page: number, pageSize: number) => ['alert', 'channels', 'list', page, pageSize] as const,
            detail: (id: string) => ['alert', 'channels', 'detail', id] as const,
        },
        rules: {
            all: ['alert', 'rules'] as const,
            lists: () => ['alert', 'rules', 'list'] as const,
            list: (page: number, pageSize: number) => ['alert', 'rules', 'list', page, pageSize] as const,
            detail: (id: string) => ['alert', 'rules', 'detail', id] as const,
        },
        history: {
            all: ['alert', 'history'] as const,
            lists: () => ['alert', 'history', 'list'] as const,
            list: (filters: Record<string, unknown>) => ['alert', 'history', 'list', filters] as const,
            stats: () => ['alert', 'history', 'stats'] as const,
        },
    },

    // ==================== 站点管理相关 ====================
    site: {
        all: ['site'] as const,
        lists: () => ['site', 'list'] as const,
        list: (page: number, pageSize: number) => ['site', 'list', page, pageSize] as const,
        detail: (id: string) => ['site', 'detail', id] as const,
    },

    // ==================== 证书管理相关 ====================
    certificate: {
        all: ['certificate'] as const,
        lists: () => ['certificate', 'list'] as const,
        list: (page: number, pageSize: number) => ['certificate', 'list', page, pageSize] as const,
        detail: (id: string) => ['certificate', 'detail', id] as const,
    },

    // ==================== 全局配置相关 ====================
    config: {
        all: ['config'] as const,
        detail: () => ['config', 'detail'] as const,
    },

    // ==================== 运行器状态相关 ====================
    runner: {
        all: ['runner'] as const,
        status: () => ['runner', 'status'] as const,
    },

    // ==================== 日志相关 ====================
    logs: {
        all: ['logs'] as const,
        attack: {
            all: ['logs', 'attack'] as const,
            lists: () => ['logs', 'attack', 'list'] as const,
            list: (params: Record<string, unknown>) => ['logs', 'attack', 'list', params] as const,
        },
        protect: {
            all: ['logs', 'protect'] as const,
            lists: () => ['logs', 'protect', 'list'] as const,
            list: (params: Record<string, unknown>) => ['logs', 'protect', 'list', params] as const,
        },
    },

    // ==================== AI 分析器相关 ====================
    aiAnalyzer: {
        all: ['aiAnalyzer'] as const,
        patterns: {
            all: ['aiAnalyzer', 'patterns'] as const,
            lists: () => ['aiAnalyzer', 'patterns', 'list'] as const,
            list: (filters: Record<string, unknown>) => ['aiAnalyzer', 'patterns', 'list', filters] as const,
        },
        rules: {
            all: ['aiAnalyzer', 'rules'] as const,
            suggestions: () => ['aiAnalyzer', 'rules', 'suggestions'] as const,
        },
        config: {
            all: ['aiAnalyzer', 'config'] as const,
            detail: () => ['aiAnalyzer', 'config', 'detail'] as const,
        },
    },

    // ==================== MCP 状态相关 ====================
    mcp: {
        all: ['mcp'] as const,
        status: () => ['mcp', 'status'] as const,
    },

    // ==================== IP 组管理相关 ====================
    ipGroup: {
        all: ['ipGroup'] as const,
        lists: () => ['ipGroup', 'list'] as const,
        list: (page: number, pageSize: number) => ['ipGroup', 'list', page, pageSize] as const,
        detail: (id: string) => ['ipGroup', 'detail', id] as const,
    },

    // ==================== 流量控制相关 ====================
    flowControl: {
        all: ['flowControl'] as const,
        config: () => ['flowControl', 'config'] as const,
        stats: () => ['flowControl', 'stats'] as const,
        blockedIP: {
            all: ['flowControl', 'blockedIP'] as const,
            lists: () => ['flowControl', 'blockedIP', 'list'] as const,
            list: <T extends Record<string, unknown>>(params: T) => ['flowControl', 'blockedIP', 'list', params] as const,
        },
    },

    // ==================== 微规则相关 ====================
    microRule: {
        all: ['microRule'] as const,
        lists: () => ['microRule', 'list'] as const,
        list: (page: number, pageSize: number) => ['microRule', 'list', page, pageSize] as const,
        detail: (id: string) => ['microRule', 'detail', id] as const,
    },

    // ==================== 规则相关 ====================
    rules: {
        all: ['rules'] as const,
        system: {
            all: ['rules', 'system'] as const,
            lists: () => ['rules', 'system', 'list'] as const,
            list: (filters: Record<string, unknown>) => ['rules', 'system', 'list', filters] as const,
        },
        user: {
            all: ['rules', 'user'] as const,
            lists: () => ['rules', 'user', 'list'] as const,
            list: (filters: Record<string, unknown>) => ['rules', 'user', 'list', filters] as const,
        },
    },
} as const

/**
 * 类型辅助 - Query Key 类型
 */
export type QueryKeys = typeof queryKeys
