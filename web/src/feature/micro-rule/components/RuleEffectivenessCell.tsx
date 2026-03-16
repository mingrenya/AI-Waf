"use client"

import { useState } from "react"
import { Badge } from "@/components/ui/badge"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { TrendingUp, TrendingDown, AlertCircle, CheckCircle2, Loader2 } from "lucide-react"
import type { RuleEffectivenessScore } from "@/types/rule-enhanced"
import { ruleEnhancedApi } from "@/api/rule-enhanced"
import { useToast } from "@/hooks/use-toast"
import { motion } from "motion/react"

interface RuleEffectivenessCellProps {
  ruleId: string
  ruleName: string
}

export function RuleEffectivenessCell({ ruleId, ruleName }: RuleEffectivenessCellProps) {
  const [score, setScore] = useState<RuleEffectivenessScore | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasLoaded, setHasLoaded] = useState(false)
  const { toast } = useToast()

  const loadScore = async () => {
    if (hasLoaded) return

    setLoading(true)
    setError(null)
    try {
      const response = await ruleEnhancedApi.calculateScore({
        rule_id: ruleId,
        period: "7d"
      })
      setScore(response)
      setHasLoaded(true)
    } catch (error) {
      const err = error as Error
      setError(err.message || "加载失败")
      toast({
        variant: "destructive",
        title: "评分加载失败",
        description: err.message || "无法计算规则有效性评分"
      })
    } finally {
      setLoading(false)
    }
  }

  const getScoreColor = (score: number) => {
    if (score >= 80) return "text-emerald-600 dark:text-emerald-400"
    if (score >= 60) return "text-blue-600 dark:text-blue-400"
    if (score >= 40) return "text-yellow-600 dark:text-yellow-400"
    return "text-red-600 dark:text-red-400"
  }

  const getScoreBgColor = (score: number) => {
    if (score >= 80) return "bg-emerald-100 dark:bg-emerald-900/30"
    if (score >= 60) return "bg-blue-100 dark:bg-blue-900/30"
    if (score >= 40) return "bg-yellow-100 dark:bg-yellow-900/30"
    return "bg-red-100 dark:bg-red-900/30"
  }

  const getPerformanceColor = (impact: string) => {
    switch (impact) {
      case "low":
        return "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400"
      case "medium":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400"
      case "high":
        return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400"
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400"
    }
  }

  const getPerformanceLabel = (impact: string) => {
    switch (impact) {
      case "low": return "性能影响低"
      case "medium": return "性能影响中"
      case "high": return "性能影响高"
      default: return "未知"
    }
  }

  if (!hasLoaded && !loading) {
    return (
      <button
        onClick={loadScore}
        className="text-sm text-muted-foreground hover:text-primary transition-colors underline-offset-4 hover:underline"
      >
        点击查看评分
      </button>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        <span className="text-sm text-muted-foreground">计算中...</span>
      </div>
    )
  }

  if (error || !score) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="flex items-center gap-2 cursor-pointer" onClick={loadScore}>
              <AlertCircle className="h-4 w-4 text-destructive" />
              <span className="text-sm text-muted-foreground">加载失败</span>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            <p className="text-xs">点击重试</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex items-center gap-3 cursor-help"
          >
            <div className={`flex items-center gap-2 px-3 py-1.5 rounded-full ${getScoreBgColor(score.score)}`}>
              <span className={`text-xl font-bold tabular-nums ${getScoreColor(score.score)}`}>
                {score.score.toFixed(0)}
              </span>
              <div className="flex flex-col items-start -space-y-1">
                <span className={`text-[10px] font-medium uppercase tracking-wide ${getScoreColor(score.score)}`}>
                  分
                </span>
              </div>
            </div>

            <div className="flex flex-col gap-1">
              <Badge variant="outline" className={`text-xs ${getPerformanceColor(score.performance_impact)}`}>
                {getPerformanceLabel(score.performance_impact)}
              </Badge>
            </div>
          </motion.div>
        </TooltipTrigger>
        <TooltipContent side="right" className="max-w-sm p-4 bg-slate-950/95 text-slate-100 border border-white/20">
          <div className="space-y-3">
            <div>
              <h4 className="font-semibold mb-1 text-sm text-white">规则有效性详情</h4>
              <p className="text-xs text-slate-300">{ruleName}</p>
            </div>

            <div className="space-y-2 text-xs">
              <div className="flex items-center justify-between gap-4">
                <span className="text-slate-300">真阳性率</span>
                <div className="flex items-center gap-1.5">
                  <span className="font-semibold tabular-nums text-white">{(score.true_positive_rate * 100).toFixed(1)}%</span>
                  {score.true_positive_rate >= 0.8 ? (
                    <TrendingUp className="h-3 w-3 text-emerald-500" />
                  ) : (
                    <TrendingDown className="h-3 w-3 text-red-500" />
                  )}
                </div>
              </div>

              <div className="flex items-center justify-between gap-4">
                <span className="text-slate-300">假阳性率</span>
                <div className="flex items-center gap-1.5">
                  <span className="font-semibold tabular-nums text-white">{(score.false_positive_rate * 100).toFixed(1)}%</span>
                  {score.false_positive_rate <= 0.1 ? (
                    <CheckCircle2 className="h-3 w-3 text-emerald-500" />
                  ) : (
                    <AlertCircle className="h-3 w-3 text-yellow-500" />
                  )}
                </div>
              </div>

              <div className="flex items-center justify-between gap-4">
                <span className="text-slate-300">拦截率</span>
                <span className="font-semibold tabular-nums text-white">{(score.block_rate * 100).toFixed(1)}%</span>
              </div>

              <div className="flex items-center justify-between gap-4">
                <span className="text-slate-300">平均匹配时间</span>
                <span className="font-semibold tabular-nums text-white">{score.avg_match_time.toFixed(2)}ms</span>
              </div>
            </div>

            {score.recommendation && (
              <div className="pt-2 border-t border-white/20">
                <p className="text-xs text-slate-200 leading-relaxed">
                  <span className="font-medium text-white">建议：</span> {score.recommendation}
                </p>
              </div>
            )}

            {score.last_evaluated && (
              <div className="pt-1 text-[10px] text-slate-300">
                最后计算时间：{new Date(score.last_evaluated).toLocaleString('zh-CN')}
              </div>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
