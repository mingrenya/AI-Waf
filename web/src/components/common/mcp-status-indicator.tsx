/**
 * MCP连接状态指示器组件
 * 显示MCP服务器的实时连接状态
 */
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'
import { CheckCircle2, XCircle, Clock, Activity } from 'lucide-react'
import { getMCPStatus } from '@/api/mcp'
import { cn } from '@/lib/utils'

export function MCPStatusIndicator() {
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['mcp-status'],
    queryFn: getMCPStatus,
    refetchInterval: 10000, // 每10秒刷新一次
  })

  const status = data?.data
  const isConnected = status?.connected || false

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={cn(
            "h-8 px-2 gap-2",
            isConnected ? "text-green-600 hover:text-green-700" : "text-red-600 hover:text-red-700"
          )}
        >
          {isConnected ? (
            <CheckCircle2 className="h-4 w-4" />
          ) : (
            <XCircle className="h-4 w-4" />
          )}
          <span className="text-xs font-medium">
            MCP {isConnected ? '已连接' : '未连接'}
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="end">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-semibold">MCP 服务器状态</h4>
            <Badge variant={isConnected ? 'default' : 'destructive'}>
              {isConnected ? '在线' : '离线'}
            </Badge>
          </div>

          <Separator />

          {isLoading ? (
            <div className="flex items-center justify-center py-4">
              <Activity className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : status ? (
            <div className="space-y-3">
              {status.serverVersion && (
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">服务器版本</span>
                  <span className="font-mono">{status.serverVersion}</span>
                </div>
              )}

              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">可用工具</span>
                <span className="font-semibold">{status.totalTools} 个</span>
              </div>

              {status.lastConnectedAt && (
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">最后连接</span>
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {new Date(status.lastConnectedAt).toLocaleString('zh-CN')}
                  </span>
                </div>
              )}

              {status.error && (
                <div className="rounded-md bg-destructive/10 p-3">
                  <p className="text-xs text-destructive">{status.error}</p>
                </div>
              )}

              {status.availableTools && status.availableTools.length > 0 && (
                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground">
                    工具列表（前5个）:
                  </p>
                  <div className="flex flex-wrap gap-1">
                    {status.availableTools.slice(0, 5).map((tool) => (
                      <Badge key={tool} variant="outline" className="text-xs">
                        {tool}
                      </Badge>
                    ))}
                    {status.availableTools.length > 5 && (
                      <Badge variant="secondary" className="text-xs">
                        +{status.availableTools.length - 5} 更多
                      </Badge>
                    )}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">
              无法获取状态信息
            </p>
          )}

          <Separator />

          <Button
            variant="outline"
            size="sm"
            className="w-full"
            onClick={() => refetch()}
          >
            刷新状态
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
