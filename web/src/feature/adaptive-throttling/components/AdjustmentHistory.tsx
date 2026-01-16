import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TrendingUp, TrendingDown, ArrowRight } from 'lucide-react'
import { adaptiveThrottlingApi } from '@/api/adaptive-throttling'
import type { ThrottleAdjustmentLog } from '@/types/adaptive-throttling'
import { DataTablePagination } from '@/components/table/pagination'
import { getCoreRowModel, getPaginationRowModel, useReactTable } from '@tanstack/react-table'

export function AdjustmentHistory() {
    const { t } = useTranslation()
    const [logs, setLogs] = useState<ThrottleAdjustmentLog[]>([])
    const [loading, setLoading] = useState(true)
    const [filterType, setFilterType] = useState<string>('all')
    const [page, setPage] = useState(1)
    const [totalPages, setTotalPages] = useState(0)

    useEffect(() => {
        fetchLogs()
    }, [filterType, page])

    const fetchLogs = async () => {
        try {
            setLoading(true)
            const response = await adaptiveThrottlingApi.getAdjustmentLogs({
                type: filterType === 'all' ? undefined : filterType as any,
                page,
                pageSize: 10
            })
            setLogs(response.results)
            setTotalPages(response.totalPages)
        } catch (error) {
            console.error('Failed to fetch adjustment logs:', error)
        } finally {
            setLoading(false)
        }
    }

    const getAdjustmentColor = (ratio: number) => {
        if (ratio > 1) return 'text-green-600'
        if (ratio < 1) return 'text-red-600'
        return 'text-gray-600'
    }

    const getTypeLabel = (type: string) => {
        const labels: Record<string, string> = {
            visit: t('adaptiveThrottling.history.visit', '访问'),
            attack: t('adaptiveThrottling.history.attack', '攻击'),
            error: t('adaptiveThrottling.history.error', '错误')
        }
        return labels[type] || type
    }

    const table = useReactTable({
        data: logs,
        columns: [],
        pageCount: totalPages,
        getCoreRowModel: getCoreRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        manualPagination: true,
        state: {
            pagination: {
                pageIndex: page - 1,
                pageSize: 10
            }
        },
        onPaginationChange: (updater) => {
            if (typeof updater === 'function') {
                const newState = updater({ pageIndex: page - 1, pageSize: 10 })
                setPage(newState.pageIndex + 1)
            }
        }
    })

    return (
        <div className="space-y-6">
            {/* 筛选器 */}
            <Card>
                <CardHeader>
                    <CardTitle>{t('adaptiveThrottling.history.filters', '筛选条件')}</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center gap-4">
                        <div className="flex-1">
                            <Select value={filterType} onValueChange={setFilterType}>
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">{t('adaptiveThrottling.history.all', '全部类型')}</SelectItem>
                                    <SelectItem value="visit">{t('adaptiveThrottling.history.visit', '访问')}</SelectItem>
                                    <SelectItem value="attack">{t('adaptiveThrottling.history.attack', '攻击')}</SelectItem>
                                    <SelectItem value="error">{t('adaptiveThrottling.history.error', '错误')}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 调整记录列表 */}
            <Card>
                <CardHeader>
                    <CardTitle>{t('adaptiveThrottling.history.title', '调整历史记录')}</CardTitle>
                    <CardDescription>
                        {t('adaptiveThrottling.history.description', '查看系统自动调整限流阈值的详细记录')}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="space-y-4">
                            {[1, 2, 3].map(i => (
                                <div key={i} className="animate-pulse">
                                    <div className="h-20 bg-gray-200 rounded" />
                                </div>
                            ))}
                        </div>
                    ) : logs.length === 0 ? (
                        <p className="text-center text-muted-foreground py-8">
                            {t('adaptiveThrottling.history.noData', '暂无调整记录')}
                        </p>
                    ) : (
                        <div className="space-y-4">
                            <div className="rounded-md border">
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>{t('adaptiveThrottling.history.time', '时间')}</TableHead>
                                            <TableHead>{t('adaptiveThrottling.history.type', '类型')}</TableHead>
                                            <TableHead>{t('adaptiveThrottling.history.adjustment', '调整')}</TableHead>
                                            <TableHead>{t('adaptiveThrottling.history.baseline', '基线')}</TableHead>
                                            <TableHead>{t('adaptiveThrottling.history.reason', '原因')}</TableHead>
                                            <TableHead>{t('adaptiveThrottling.history.triggeredBy', '触发方式')}</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {logs.map((log) => (
                                            <TableRow key={log.id}>
                                                <TableCell className="whitespace-nowrap">
                                                    {new Date(log.timestamp).toLocaleString()}
                                                </TableCell>
                                                <TableCell>
                                                    <Badge variant="outline">{getTypeLabel(log.type)}</Badge>
                                                </TableCell>
                                                <TableCell>
                                                    <div className="flex items-center gap-2">
                                                        <span className="font-mono">{log.oldThreshold}</span>
                                                        <ArrowRight className={`h-4 w-4 ${getAdjustmentColor(log.adjustmentRatio)}`} />
                                                        <span className="font-mono font-semibold">{log.newThreshold}</span>
                                                        {log.adjustmentRatio !== 1 && (
                                                            <>
                                                                {log.adjustmentRatio > 1 ? (
                                                                    <TrendingUp className="h-4 w-4 text-green-600" />
                                                                ) : (
                                                                    <TrendingDown className="h-4 w-4 text-red-600" />
                                                                )}
                                                                <span className={`text-sm ${getAdjustmentColor(log.adjustmentRatio)}`}>
                                                                    {(log.adjustmentRatio * 100 - 100).toFixed(1)}%
                                                                </span>
                                                            </>
                                                        )}
                                                    </div>
                                                </TableCell>
                                                <TableCell>
                                                    <div className="flex items-center gap-2">
                                                        <span className="font-mono text-sm">{log.oldBaseline.toFixed(2)}</span>
                                                        <ArrowRight className="h-3 w-3 text-gray-400" />
                                                        <span className="font-mono text-sm font-semibold">{log.newBaseline.toFixed(2)}</span>
                                                    </div>
                                                </TableCell>
                                                <TableCell className="max-w-xs">
                                                    <p className="text-sm truncate" title={log.reason}>
                                                        {log.reason}
                                                    </p>
                                                    <p className="text-xs text-muted-foreground">
                                                        {t('adaptiveThrottling.history.currentTraffic', '当前流量')}: {log.currentTraffic.toFixed(2)} | 
                                                        {t('adaptiveThrottling.history.anomalyScore', '异常分数')}: {log.anomalyScore.toFixed(2)}
                                                    </p>
                                                </TableCell>
                                                <TableCell>
                                                    <Badge variant={log.triggeredBy === 'auto' ? 'default' : 'secondary'}>
                                                        {log.triggeredBy === 'auto' 
                                                            ? t('adaptiveThrottling.history.auto', '自动')
                                                            : t('adaptiveThrottling.history.manual', '手动')
                                                        }
                                                    </Badge>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </div>
                            <DataTablePagination table={table} />
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}
