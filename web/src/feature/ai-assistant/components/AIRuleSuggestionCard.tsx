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
  AlertTriangle,
  Info,
  TrendingUp,
  Clock,
} from 'lucide-react'
import { getAIRuleSuggestions, approveAIRuleSuggestion, rejectAIRuleSuggestion, deployAIRuleSuggestion } from '@/api/mcp'
import type { AIRuleSuggestion } from '@/types/mcp'
import { useToast } from '@/hooks/use-toast'

const severityConfig = {
  critical: { label: '严重', variant: 'destructive' as const, icon: AlertTriangle, color: 'text-red-600' },
  high: { label: '高', variant: 'destructive' as const, icon: AlertTriangle, color: 'text-orange-600' },
  medium: { label: '中', variant: 'default' as const, icon: Info, color: 'text-yellow-600' },
  low: { label: '低', variant: 'secondary' as const, icon: Info, color: 'text-blue-600' },
}

const statusConfig = {
  pending: { label: '待审核', variant: 'secondary' as const },
  approved: { label: '已批准', variant: 'default' as const },
  rejected: { label: '已拒绝', variant: 'destructive' as const },
  deployed: { label: '已部署', variant: 'outline' as const },
}

export function AIRuleSuggestionCard() {
  const [statusFilter, setStatusFilter] = useState<string>('pending')
  const [severityFilter, setSeverityFilter] = useState<string>('all')
  const { toast } = useToast()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['ai-rule-suggestions', statusFilter, severityFilter],
    queryFn: () =>
      getAIRuleSuggestions({
        status: statusFilter,
        severity: severityFilter === 'all' ? undefined : severityFilter,
        limit: 20,
      }),
  })

  const approveMutation = useMutation({
    mutationFn: approveAIRuleSuggestion,
    onSuccess: () => {
      toast({ title: '规则已批准' })
      queryClient.invalidateQueries({ queryKey: ['ai-rule-suggestions'] })
    },
    onError: () => {
      toast({ title: '批准失败', variant: 'destructive' })
    },
  })

  const rejectMutation = useMutation({
    mutationFn: (suggestionId: string) => rejectAIRuleSuggestion(suggestionId),
    onSuccess: () => {
      toast({ title: '规则已拒绝' })
      queryClient.invalidateQueries({ queryKey: ['ai-rule-suggestions'] })
    },
    onError: () => {
      toast({ title: '拒绝失败', variant: 'destructive' })
    },
  })

  const deployMutation = useMutation({
    mutationFn: deployAIRuleSuggestion,
    onSuccess: () => {
      toast({ title: '规则已部署' })
      queryClient.invalidateQueries({ queryKey: ['ai-rule-suggestions'] })
    },
    onError: () => {
      toast({ title: '部署失败', variant: 'destructive' })
    },
  })

  const suggestions = data?.data?.data || []

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
            <Select value={severityFilter} onValueChange={setSeverityFilter}>
              <SelectTrigger className="w-[120px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部严重程度</SelectItem>
                <SelectItem value="critical">严重</SelectItem>
                <SelectItem value="high">高</SelectItem>
                <SelectItem value="medium">中</SelectItem>
                <SelectItem value="low">低</SelectItem>
              </SelectContent>
            </Select>
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
          ) : suggestions.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Info className="h-12 w-12 mb-2 opacity-50" />
              <p>暂无规则建议</p>
            </div>
          ) : (
            <div className="space-y-4">
              {suggestions.map((suggestion: AIRuleSuggestion) => {
                const SeverityIcon = severityConfig[suggestion.severity].icon
                return (
                  <Card key={suggestion.id}>
                    <CardHeader className="pb-3">
                      <div className="flex items-start justify-between">
                        <div className="space-y-1 flex-1">
                          <CardTitle className="text-base">{suggestion.ruleName}</CardTitle>
                          <CardDescription className="text-sm">
                            来源模式: {suggestion.patternName}
                          </CardDescription>
                        </div>
                        <div className="flex flex-col items-end gap-2">
                          <Badge variant={severityConfig[suggestion.severity].variant}>
                            <SeverityIcon className="h-3 w-3 mr-1" />
                            {severityConfig[suggestion.severity].label}
                          </Badge>
                          <Badge variant={statusConfig[suggestion.status].variant}>
                            {statusConfig[suggestion.status].label}
                          </Badge>
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="space-y-3 pb-3">
                      <div className="space-y-2">
                        <p className="text-sm">{suggestion.description}</p>
                        <div className="flex items-center gap-4 text-xs text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {new Date(suggestion.createdAt).toLocaleString('zh-CN')}
                          </span>
                          <span>置信度: {(suggestion.confidence * 100).toFixed(1)}%</span>
                          <Badge variant="outline" className="text-xs">
                            {suggestion.ruleType}
                          </Badge>
                        </div>
                      </div>
                      <Separator />
                      <div className="space-y-1">
                        <p className="text-xs font-medium">建议:</p>
                        <p className="text-xs text-muted-foreground">{suggestion.recommendation}</p>
                      </div>
                    </CardContent>
                    <CardFooter className="flex justify-end gap-2">
                      {suggestion.status === 'pending' && (
                        <>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => rejectMutation.mutate(suggestion.id)}
                            disabled={rejectMutation.isPending}
                          >
                            <XCircle className="h-4 w-4 mr-1" />
                            拒绝
                          </Button>
                          <Button
                            size="sm"
                            onClick={() => approveMutation.mutate(suggestion.id)}
                            disabled={approveMutation.isPending}
                          >
                            <CheckCircle className="h-4 w-4 mr-1" />
                            批准
                          </Button>
                        </>
                      )}
                      {suggestion.status === 'approved' && (
                        <Button
                          size="sm"
                          onClick={() => deployMutation.mutate(suggestion.id)}
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
