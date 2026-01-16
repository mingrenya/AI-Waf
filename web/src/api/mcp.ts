/**
 * MCP相关API服务
 */
import api from './index'
import type { MCPConnectionStatus, MCPToolCall, AIRuleSuggestion, AIAnalysisResult } from '@/types/mcp'

/**
 * 获取MCP连接状态
 */
export const getMCPStatus = () => {
  return api.get<MCPConnectionStatus>('/mcp/status')
}

/**
 * 获取MCP工具列表
 */
export const getMCPTools = () => {
  return api.get<{ tools: string[] }>('/mcp/tools')
}

/**
 * 获取MCP工具调用历史
 */
export const getMCPToolCallHistory = (params?: { limit?: number; offset?: number }) => {
  return api.get<{ data: MCPToolCall[]; total: number }>('/mcp/tool-calls', { params })
}

/**
 * 获取AI规则建议列表
 */
export const getAIRuleSuggestions = (params?: {
  status?: string
  severity?: string
  limit?: number
  offset?: number
}) => {
  return api.get<{ data: AIRuleSuggestion[]; total: number }>('/ai-analyzer/suggestions', { params })
}

/**
 * 批准AI规则建议
 */
export const approveAIRuleSuggestion = (suggestionId: string) => {
  return api.post(`/ai-analyzer/suggestions/${suggestionId}/approve`)
}

/**
 * 拒绝AI规则建议
 */
export const rejectAIRuleSuggestion = (suggestionId: string, reason?: string) => {
  return api.post(`/ai-analyzer/suggestions/${suggestionId}/reject`, { reason })
}

/**
 * 部署AI规则建议
 */
export const deployAIRuleSuggestion = (suggestionId: string) => {
  return api.post(`/ai-analyzer/suggestions/${suggestionId}/deploy`)
}

/**
 * 获取AI分析结果
 */
export const getAIAnalysisResult = (timeRange?: string) => {
  return api.get<AIAnalysisResult>('/ai-analyzer/analysis/result', {
    params: { timeRange }
  })
}

/**
 * 触发AI分析
 */
export const triggerAIAnalysis = (params?: {
  timeRange?: string
  minSamples?: number
  anomalyThreshold?: number
  clusteringMethod?: string
}) => {
  return api.post('/ai-analyzer/analyze/patterns', params)
}
