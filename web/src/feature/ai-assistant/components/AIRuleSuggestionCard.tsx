/**
 * AI规则建议卡片组件
 * 显示AI生成的规则建议，支持批准、拒绝和部署操作
 */
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  CheckCircle,
  XCircle,
  Rocket,
  Info,
  TrendingUp,
  Clock,
} from 'lucide-react'
import { aiAnalyzerApi } from '@/api/ai-analyzer'
import type { GeneratedRule } from '@/types/ai-analyzer'
import { toast } from '@/store'

const statusConfig = {
  pending: { label: '待审核', variant: 'secondary' as const },
  approved: { label: '已批准', variant: 'default' as const },
  rejected: { label: '已拒绝', variant: 'destructive' as const },
  deployed: { label: '已部署', variant: 'outline' as const },
}

export function AIRuleSuggestionCard() {
  const [statusFilter, setStatusFilter] = useState<string>('pending')
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['generated-rules', statusFilter],
    queryFn: () =>
      aiAnalyzerApi.listGeneratedRules({
        page: 1,
        size: 20,
        status: statusFilter,
      }),
  })

  const reviewMutation = useMutation({
    mutationFn: ({ ruleId, action, comment }: { ruleId: string; action: 'approve' | 'reject'; comment: string }) =>
      aiAnalyzerApi.reviewRule({ ruleId, action, comment }),
    onSuccess: (_, variables) => {
      toast({ 
        title: variables.action === 'approve' ? '规则已批准' : '规则已拒绝'
      })
      queryClient.invalidateQueries({ queryKey: ['generated-rules'] })
    },
    onError: () => {
      toast({ title: '操作失败', variant: 'destructive' })
    },
  })

  const deployMutation = useMutation({
    mutationFn: aiAnalyzerApi.deployRule,
    onSuccess: () => {
      toast({ title: '规则已部署' })
      queryClient.invalidateQueries({ queryKey: ['generated-rules'] })
    },
    onError: () => {
      toast({ title: '部署失败', variant: 'destructive' })
    },
  })

  const rules = data?.list || []

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              AI 规则建议
            </CardTitle>
            <CardDescription>
              基于攻击模式分析自动生成的防护规则建议
            </CardDescription>
          </div>
          <div className="flex gap-2">
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[120px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="pending">待审核</SelectItem>
                <SelectItem value="approved">已批准</SelectItem>
                <SelectItem value="rejected">已拒绝</SelectItem>
                <SelectItem value="deployed">已部署</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[500px] pr-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <p className="text-muted-foreground">加载中...</p>
            </div>
          ) : rules.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Info className="h-12 w-12 mb-2 opacity-50" />
              <p>暂无规则建议</p>
            </div>
          ) : (
            <div className="space-y-4">
              {rules.map((rule: GeneratedRule) => {
                return (
                  <Card key={rule.id}>
                    <CardHeader className="pb-3">
                      <div className="flex items-start justify-between">
                        <div className="space-y-1 flex-1">
                          <CardTitle className="text-base">{rule.rule_type}</CardTitle>
                          <CardDescription className="text-sm">
                            模式 ID: {rule.pattern_id || 'N/A'}
                          </CardDescription>
                        </div>
                        <div className="flex flex-col items-end gap-2">
                          <Badge variant={statusConfig[rule.status]?.variant || 'outline'}>
                            {statusConfig[rule.status]?.label || rule.status}
                          </Badge>
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="space-y-3 pb-3">
                      <div className="space-y-2">
                        <div className="flex items-center gap-4 text-xs text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {new Date(rule.created_at).toLocaleString('zh-CN')}
                          </span>
                          <span>置信度: {(rule.confidence * 100).toFixed(1)}%</span>
                          <Badge variant="outline" className="text-xs">
                            {rule.rule_type}
                          </Badge>
                        </div>
                      </div>
                      <Separator />
                      <div className="space-y-1">
                        <p className="text-xs font-medium">规则内容:</p>
                        <pre className="text-xs text-muted-foreground bg-muted p-2 rounded overflow-x-auto">
                          {rule.rule_content}
                        </pre>
                      </div>
                      {rule.review_comment && (
                        <>
                          <Separator />
                          <div className="space-y-1">
                            <p className="text-xs font-medium">审核意见:</p>
                            <p className="text-xs text-muted-foreground">{rule.review_comment}</p>
                          </div>
                        </>
                      )}
                    </CardContent>
                    <CardFooter className="flex justify-end gap-2">
                      {rule.status === 'pending' && (
                        <>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => reviewMutation.mutate({ ruleId: rule.id, action: 'reject', comment: '拒绝' })}
                            disabled={reviewMutation.isPending}
                          >
                            <XCircle className="h-4 w-4 mr-1" />
                            拒绝
                          </Button>
                          <Button
                            size="sm"
                            onClick={() => reviewMutation.mutate({ ruleId: rule.id, action: 'approve', comment: '批准' })}
                            disabled={reviewMutation.isPending}
                          >
                            <CheckCircle className="h-4 w-4 mr-1" />
                            批准
                          </Button>
                        </>
                      )}
                      {rule.status === 'approved' && (
                        <Button
                          size="sm"
                          onClick={() => deployMutation.mutate(rule.id)}
                          disabled={deployMutation.isPending}
                        >
                          <Rocket className="h-4 w-4 mr-1" />
                          部署
                        </Button>
                      )}
                    </CardFooter>
                  </Card>
                )
              })}
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  )
}
