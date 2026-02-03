import { get, post } from "./index"
import type {
  RuleTemplateListResponse,
  RuleTemplate,
  RuleEffectivenessScoreListResponse,
  RuleEffectivenessScore,
  ProtectionProfileListResponse,
  ProtectionProfile,
  ApplyProfileResponse
} from "@/types/rule-enhanced"

export const ruleEnhancedApi = {
  // ============== 规则模板相关 ==============
  
  // 获取规则模板列表
  listTemplates: (params?: {
    category?: string
    severity?: string
  }) => get<RuleTemplateListResponse>("/rules/templates", { params }),

  // 获取模板详情
  getTemplate: (id: string) => get<RuleTemplate>(`/rules/templates/${id}`),

  // 从模板创建规则
  createRuleFromTemplate: (data: {
    template_id: string
    custom_name?: string
  }) => post<{ message: string; rule_id: string; name: string }>("/rules/templates/create-rule", data),

  // ============== 规则有效性评分相关 ==============
  
  // 获取规则评分列表
  listScores: (params?: {
    sortBy?: string
    order?: number
  }) => get<RuleEffectivenessScoreListResponse>("/rules/effectiveness", { params }),

  // 获取单个规则评分
  getScore: (id: string) => get<RuleEffectivenessScore>(`/rules/effectiveness/${id}`),

  // 计算规则评分
  calculateScore: (data: {
    rule_id: string
    period: "24h" | "7d" | "30d"
  }) => post<RuleEffectivenessScore>("/rules/effectiveness/calculate", data),

  // 批量计算评分
  batchCalculateScores: (data: {
    period: "24h" | "7d" | "30d"
  }) => post<{ message: string }>("/rules/effectiveness/batch-calculate", data),

  // ============== 保护配置文件相关 ==============
  
  // 获取保护配置文件列表
  listProfiles: () => get<ProtectionProfileListResponse>("/rules/profiles"),

  // 获取配置文件详情
  getProfile: (id: string) => get<ProtectionProfile>(`/rules/profiles/${id}`),

  // 应用保护配置文件
  applyProfile: (data: {
    profile_id: string
  }) => post<ApplyProfileResponse>("/rules/profiles/apply", data),
}
