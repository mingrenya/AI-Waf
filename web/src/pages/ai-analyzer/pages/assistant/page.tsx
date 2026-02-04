/**
 * AI助手页面
 * 展示AI规则建议和AI助手交互界面
 */
import { useQuery } from '@tanstack/react-query'
import { AIRuleSuggestionCard } from '@/feature/ai-assistant'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Sparkles, TrendingUp, Activity, Target } from 'lucide-react'
import { aiAnalyzerApi } from '@/api/ai-analyzer'

export default function AIAssistantPage() {
  // 获取AI分析器统计数据
  const { data: stats, isLoading } = useQuery({
    queryKey: ['ai-analyzer-stats'],
    queryFn: () => aiAnalyzerApi.getAnalyzerStats({}),
    refetchInterval: 30000, // 每30秒刷新一次
  })

  return (
    <div className="p-6 space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center gap-3">
        <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
          <Sparkles className="h-6 w-6 text-primary" />
        </div>
        <div>
          <h1 className="text-3xl font-bold">AI 智能助手</h1>
          <p className="text-muted-foreground">
            通过 MCP 协议提供 AI 驱动的安全分析和规则生成
          </p>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待审核规则</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-16" />
            ) : (
              <>
                <div className="text-2xl font-bold">{Number(stats?.rules_pending) || 0}</div>
                <p className="text-xs text-muted-foreground">
                  需要人工审核
                </p>
              </>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">已部署规则</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-16" />
            ) : (
              <>
                <div className="text-2xl font-bold">{Number(stats?.rules_deployed) || 0}</div>
                <p className="text-xs text-muted-foreground">
                  正在生效中
                </p>
              </>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">已生成规则</CardTitle>
            <Sparkles className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-16" />
            ) : (
              <>
                <div className="text-2xl font-bold">{Number(stats?.rules_generated) || 0}</div>
                <p className="text-xs text-muted-foreground">
                  AI 自动生成
                </p>
              </>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">检测模式数</CardTitle>
            <Target className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-16" />
            ) : (
              <>
                <div className="text-2xl font-bold">{Number(stats?.patterns_detected) || 0}</div>
                <p className="text-xs text-muted-foreground">
                  攻击模式识别
                </p>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* AI规则建议 */}
      <AIRuleSuggestionCard />
    </div>
  )
}
