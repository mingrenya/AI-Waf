/**
 * MCP（Model Context Protocol）相关类型定义
 */

/**
 * MCP连接状态
 */
export interface MCPConnectionStatus {
  connected: boolean
  lastConnectedAt?: string
  serverVersion?: string
  totalTools: number
  availableTools: string[]
  error?: string
}

/**
 * MCP工具调用记录
 */
export interface MCPToolCall {
  id: string
  toolName: string
  timestamp: string
  duration: number
  success: boolean
  error?: string
}

/**
 * AI助手消息
 */
export interface AIAssistantMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
  toolCalls?: MCPToolCall[]
}

/**
 * AI助手会话
 */
export interface AIAssistantSession {
  id: string
  title: string
  messages: AIAssistantMessage[]
  createdAt: string
  updatedAt: string
}

/**
 * AI生成的规则建议
 */
export interface AIRuleSuggestion {
  id: string
  patternId?: string
  patternName: string
  ruleName: string
  ruleType: 'micro_rule' | 'modsecurity'
  confidence: number
  severity: 'low' | 'medium' | 'high' | 'critical'
  description: string
  recommendation: string
  ruleContent: any
  status: 'pending' | 'approved' | 'rejected' | 'deployed'
  createdAt: string
  reviewedAt?: string
  deployedAt?: string
}

/**
 * AI分析结果
 */
export interface AIAnalysisResult {
  totalPatterns: number
  highSeverityPatterns: number
  suggestedRules: number
  processingTime: number
  timestamp: string
}
