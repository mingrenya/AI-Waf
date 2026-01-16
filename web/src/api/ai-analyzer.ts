import { get, post, put, del } from "./index"

export const aiAnalyzerApi = {
    // 攻击模式
    listAttackPatterns: (params: {
        page: number
        size: number
        severity?: string
        attackType?: string
        startTime?: string
        endTime?: string
    }) => get<{ list: any[], total: number }>("/ai-analyzer/patterns", { params }),

    getAttackPattern: (id: string) => get(`/ai-analyzer/patterns/${id}`),

    deleteAttackPattern: (id: string) => del(`/ai-analyzer/patterns/${id}`),

    // 生成规则
    listGeneratedRules: (params: {
        page: number
        size: number
        status?: string
        ruleType?: string
        patternId?: string
    }) => get<{ list: any[], total: number }>("/ai-analyzer/rules", { params }),

    getGeneratedRule: (id: string) => get(`/ai-analyzer/rules/${id}`),

    deleteGeneratedRule: (id: string) => del(`/ai-analyzer/rules/${id}`),

    reviewRule: (params: { ruleId: string; action: "approve" | "reject"; comment: string }) =>
        post("/ai-analyzer/rules/review", params),

    getPendingRules: (params: { page: number; size: number }) =>
        get<{ list: any[], total: number }>("/ai-analyzer/rules/pending", { params }),

    deployRule: (id: string) => post(`/ai-analyzer/rules/${id}/deploy`),

    // AI分析器配置
    getAnalyzerConfig: () => get<any>("/ai-analyzer/config"),

    updateAnalyzerConfig: (config: any) => put("/ai-analyzer/config", config),

    // MCP对话
    listMCPConversations: (params: { patternId?: string; page: number; size: number }) =>
        get<{ list: any[], total: number }>("/ai-analyzer/conversations", { params }),

    getMCPConversation: (id: string) => get(`/ai-analyzer/conversations/${id}`),

    deleteMCPConversation: (id: string) => del(`/ai-analyzer/conversations/${id}`),

    // 统计分析
    getAnalyzerStats: (params: { startTime?: string; endTime?: string }) =>
        get("/ai-analyzer/stats", { params }),
    
    // 手动触发AI分析
    triggerAnalysis: () => post("/ai-analyzer/trigger"),
}
