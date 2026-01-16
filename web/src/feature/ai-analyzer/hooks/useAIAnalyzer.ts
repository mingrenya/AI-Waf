import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { aiAnalyzerApi } from "@/api/ai-analyzer"
import { toast } from "@/hooks/use-toast"
import type {
  AttackPatternListParams,
  GeneratedRuleListParams,
  MCPConversationListParams,
} from "@/types/ai-analyzer"

// ============================================
// 攻击模式相关hooks
// ============================================

export const useAttackPatterns = (params: AttackPatternListParams = { page: 1, size: 10 }) => {
  const queryParams = { page: 1, size: 10, ...params }
  return useQuery({
    queryKey: ["attack-patterns", queryParams],
    queryFn: () => aiAnalyzerApi.listAttackPatterns(queryParams),
  })
}

export const useAttackPattern = (id: string) => {
  return useQuery({
    queryKey: ["attack-pattern", id],
    queryFn: () => aiAnalyzerApi.getAttackPattern(id),
    enabled: !!id,
  })
}

export const useDeleteAttackPattern = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: aiAnalyzerApi.deleteAttackPattern,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["attack-patterns"] })
      toast({
        title: "删除成功",
        description: "攻击模式已删除",
      })
    },
    onError: (error: any) => {
      toast({
        title: "删除失败",
        description: error.message || "请稍后重试",
        variant: "destructive",
      })
    },
  })
}

// ============================================
// 生成规则相关hooks
// ============================================

export const useGeneratedRules = (params: GeneratedRuleListParams = { page: 1, size: 10 }) => {
  const queryParams = { page: 1, size: 10, ...params }
  return useQuery({
    queryKey: ["generated-rules", queryParams],
    queryFn: () => aiAnalyzerApi.listGeneratedRules(queryParams),
  })
}

export const useGeneratedRule = (id: string) => {
  return useQuery({
    queryKey: ["generated-rule", id],
    queryFn: () => aiAnalyzerApi.getGeneratedRule(id),
    enabled: !!id,
  })
}

export const usePendingRules = () => {
  return useQuery({
    queryKey: ["pending-rules"],
    queryFn: () => aiAnalyzerApi.getPendingRules({ page: 1, size: 10 }),
  })
}

export const useDeleteGeneratedRule = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: aiAnalyzerApi.deleteGeneratedRule,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["generated-rules"] })
      queryClient.invalidateQueries({ queryKey: ["pending-rules"] })
      toast({
        title: "删除成功",
        description: "规则已删除",
      })
    },
    onError: (error: any) => {
      toast({
        title: "删除失败",
        description: error.message || "请稍后重试",
        variant: "destructive",
      })
    },
  })
}

export const useReviewRule = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ ruleId, action, comment }: { ruleId: string; action: "approve" | "reject"; comment?: string }) =>
      aiAnalyzerApi.reviewRule({ ruleId, action, comment: comment || "" }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["generated-rules"] })
      queryClient.invalidateQueries({ queryKey: ["pending-rules"] })
      queryClient.invalidateQueries({ queryKey: ["generated-rule", variables.ruleId] })
      queryClient.invalidateQueries({ queryKey: ["analyzer-stats"] })
      toast({
        title: "审核成功",
        description: `规则已${variables.action === "approve" ? "批准" : "拒绝"}`,
      })
    },
    onError: (error: any) => {
      toast({
        title: "审核失败",
        description: error.message || "请稍后重试",
        variant: "destructive",
      })
    },
  })
}

export const useDeployRule = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: aiAnalyzerApi.deployRule,
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["generated-rules"] })
      queryClient.invalidateQueries({ queryKey: ["generated-rule", id] })
      queryClient.invalidateQueries({ queryKey: ["analyzer-stats"] })
      toast({
        title: "部署成功",
        description: "规则已部署到生产环境",
      })
    },
    onError: (error: any) => {
      toast({
        title: "部署失败",
        description: error.message || "请稍后重试",
        variant: "destructive",
      })
    },
  })
}

// ============================================
// 配置相关hooks
// ============================================

export const useAnalyzerConfig = () => {
  return useQuery({
    queryKey: ["ai-analyzer-config"],
    queryFn: aiAnalyzerApi.getAnalyzerConfig,
  })
}

export const useUpdateAnalyzerConfig = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: aiAnalyzerApi.updateAnalyzerConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ai-analyzer-config"] })
      toast({
        title: "配置更新成功",
        description: "AI分析器配置已保存",
      })
    },
    onError: (error: any) => {
      toast({
        title: "配置更新失败",
        description: error.message || "请稍后重试",
        variant: "destructive",
      })
    },
  })
}

// ============================================
// MCP对话相关hooks
// ============================================

export const useMCPConversations = (params: MCPConversationListParams = { page: 1, size: 10 }) => {
  const queryParams = { page: 1, size: 10, ...params }
  return useQuery({
    queryKey: ["mcp-conversations", queryParams],
    queryFn: () => aiAnalyzerApi.listMCPConversations(queryParams),
  })
}

export const useMCPConversation = (id: string) => {
  return useQuery({
    queryKey: ["mcp-conversation", id],
    queryFn: () => aiAnalyzerApi.getMCPConversation(id),
    enabled: !!id,
  })
}

export const useDeleteMCPConversation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: aiAnalyzerApi.deleteMCPConversation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["mcp-conversations"] })
      toast({
        title: "删除成功",
        description: "对话记录已删除",
      })
    },
    onError: (error: any) => {
      toast({
        title: "删除失败",
        description: error.message || "请稍后重试",
        variant: "destructive",
      })
    },
  })
}

// ============================================
// 统计和触发相关hooks
// ============================================

export const useAnalyzerStats = () => {
  return useQuery({
    queryKey: ["analyzer-stats"],
    queryFn: () => aiAnalyzerApi.getAnalyzerStats({}),
    refetchInterval: 30000, // 每30秒刷新一次
  })
}

export const useTriggerAnalysis = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: aiAnalyzerApi.triggerAnalysis,
    onSuccess: () => {
      toast({
        title: "AI分析已触发",
        description: "正在后台运行攻击模式检测，请稍后查看结果",
      })
      // 5秒后刷新相关数据
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ["attack-patterns"] })
        queryClient.invalidateQueries({ queryKey: ["generated-rules"] })
        queryClient.invalidateQueries({ queryKey: ["analyzer-stats"] })
      }, 5000)
    },
    onError: (error: any) => {
      toast({
        title: "触发失败",
        description: error.message || "请检查系统状态后重试",
        variant: "destructive",
      })
    },
  })
}
