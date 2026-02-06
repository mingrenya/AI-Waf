import { useMemo, useRef, useEffect, useState } from 'react'
import {
    useReactTable,
    getCoreRowModel,
    ColumnDef,
} from '@tanstack/react-table'
import { useInfiniteQuery } from '@tanstack/react-query'
import { alertHistoryApi } from '@/api/alert'
import { AlertHistory, AlertHistoryStatus, AlertSeverity } from '@/types/alert'
import { Button } from '@/components/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Loader2, MoreHorizontal, ListFilter } from 'lucide-react'
import { DataTable } from '@/components/table/motion-data-table'
import { useTranslation } from 'react-i18next'
import { queryKeys } from '@/lib/query-keys'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'

interface HistoryTableProps {
    onViewDetail?: (history: AlertHistory) => void
}

interface PaginatedResponse {
    items: AlertHistory[]
    total: number
    page: number
    pageSize: number
}

const PAGE_SIZE = 20

export function HistoryTable({ onViewDetail }: HistoryTableProps) {
    const { t } = useTranslation()
    const sentinelRef = useRef<HTMLDivElement | null>(null)

    const [severityFilter, setSeverityFilter] = useState<string>('')
    const [statusFilter, setStatusFilter] = useState<string>('')

    const filters = useMemo(
        () => ({
            severity: severityFilter,
            status: statusFilter,
        }),
        [severityFilter, statusFilter],
    )

    const {
        data,
        isLoading,
        error,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        refetch,
    } = useInfiniteQuery({
        queryKey: queryKeys.alert.history.list(filters),
        queryFn: ({ pageParam = 1 }) => {
            const page = typeof pageParam === 'number' ? pageParam : 1
            const severity = severityFilter || undefined
            const status = statusFilter || undefined
            return alertHistoryApi.getHistory(page, PAGE_SIZE, undefined, severity, status)
        },
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            if (!lastPage || typeof lastPage.total === 'undefined') return undefined
            if (!allPages || allPages.length === 0) return undefined
            const loaded = allPages.reduce(
                (sum, page) => sum + (page?.items?.length || 0),
                0,
            )
            return loaded < lastPage.total ? allPages.length + 1 : undefined
        },
    })

    const flatData = useMemo(() => {
        if (!data?.pages) return []
        return data.pages
            .filter(
                (page): page is PaginatedResponse =>
                    !!page && Array.isArray(page.items),
            )
            .flatMap((page) => page.items)
    }, [data])

    useEffect(() => {
        if (!hasNextPage || !sentinelRef.current) return
        const observer = new IntersectionObserver(
            (entries) => {
                const [entry] = entries
                if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
                    fetchNextPage()
                }
            },
            {
                threshold: 0.1,
                rootMargin: '100px 0px',
            },
        )

        const sentinel = sentinelRef.current
        if (sentinel) {
            observer.observe(sentinel)
        }

        return () => observer.disconnect()
    }, [hasNextPage, isFetchingNextPage, fetchNextPage])

    const columns: ColumnDef<AlertHistory>[] = useMemo(
        () => [
            {
                accessorKey: 'ruleName',
                header: () => (
                    <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">
                        {t('alert.ruleName')}
                    </div>
                ),
                cell: ({ row }) => (
                    <div className="font-medium dark:text-shadow-glow-white">
                        {row.original.ruleName || '-'}
                    </div>
                ),
            },
            {
                accessorKey: 'severity',
                header: () => (
                    <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">
                        {t('alert.severityLabel', { defaultValue: 'Severity' })}
                    </div>
                ),
                cell: ({ row }) => (
                    <div className="dark:text-shadow-glow-white">
                        {t(`alert.severity.${row.original.severity}`)}
                    </div>
                ),
            },
            {
                accessorKey: 'status',
                header: () => (
                    <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">
                        {t('alert.status')}
                    </div>
                ),
                cell: ({ row }) => (
                    <div className="dark:text-shadow-glow-white capitalize">
                        {row.original.status}
                    </div>
                ),
            },
            {
                accessorKey: 'message',
                header: () => (
                    <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">
                        {t('message', { defaultValue: 'Message' })}
                    </div>
                ),
                cell: ({ row }) => (
                    <div className="max-w-xs truncate dark:text-shadow-glow-white">
                        {row.original.message}
                    </div>
                ),
            },
            {
                accessorKey: 'triggeredAt',
                header: () => (
                    <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">
                        {t('alert.triggeredAt', {
                            defaultValue: 'Triggered at',
                        })}
                    </div>
                ),
                cell: ({ row }) => (
                    <div className="dark:text-shadow-glow-white">
                        {row.original.triggeredAt
                            ? new Date(
                                  row.original.triggeredAt,
                              ).toLocaleString()
                            : '-'}
                    </div>
                ),
            },
            {
                id: 'actions',
                header: () => (
                    <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">
                        {t('ipGroup.table.actions')}
                    </div>
                ),
                cell: ({ row }) => (
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 dark:text-shadow-glow-white"
                                aria-label={t('ipGroup.table.actions')}
                            >
                                <MoreHorizontal className="h-4 w-4" />
                                <span className="sr-only">
                                    {t('ipGroup.table.actions')}
                                </span>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent
                            align="end"
                            className="dark:bg-muted/95 dark:border-primary/20"
                        >
                            {onViewDetail && (
                                <DropdownMenuItem
                                    onClick={() => onViewDetail(row.original)}
                                    className="dark:text-shadow-glow-white dark:hover:bg-primary/20 cursor-pointer"
                                >
                                    {t('detail')}
                                </DropdownMenuItem>
                            )}
                        </DropdownMenuContent>
                    </DropdownMenu>
                ),
            },
        ],
        [t, onViewDetail],
    )

    const table = useReactTable({
        data: flatData,
        columns,
        getCoreRowModel: getCoreRowModel(),
    })

    if (isLoading) {
        return (
            <div className="flex items-center justify-center p-8">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
        )
    }

    if (error) {
        return (
            <div className="flex flex-col items-center justify-center p-8 text-destructive">
                <p className="text-sm font-medium">
                    {t('error.loading', { defaultValue: 'Failed to load' })}
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                    {error instanceof Error ? error.message : String(error)}
                </p>
                <Button
                    onClick={() => refetch()}
                    variant="outline"
                    size="sm"
                    className="mt-4"
                >
                    {t('common.retry')}
                </Button>
            </div>
        )
    }

    if (!isLoading && flatData.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center p-8 text-muted-foreground">
                <p className="text-sm">
                    {t('alert.noHistory', {
                        defaultValue: 'No alert history records.',
                    })}
                </p>
            </div>
        )
    }

    return (
        <div className="space-y-4">
            {/* 筛选器 */}
            <div className="flex flex-wrap items-center gap-3 mb-2">
                <div className="flex items-center gap-2 text-sm text-muted-foreground dark:text-shadow-glow-white">
                    <ListFilter className="h-4 w-4" />
                    {t('filter')}
                </div>
                <Select
                    value={severityFilter}
                    onValueChange={(val) => setSeverityFilter(val === 'all' ? '' : val)}
                >
                    <SelectTrigger className="w-40 dark:text-shadow-glow-white">
                        <SelectValue
                            placeholder={t('alert.severityFilter', {
                                defaultValue: 'Severity',
                            })}
                        />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">
                            {t('alert.allSeverities', { defaultValue: 'All' })}
                        </SelectItem>
                        <SelectItem value={AlertSeverity.Low}>
                            {t('alert.severity.low')}
                        </SelectItem>
                        <SelectItem value={AlertSeverity.Medium}>
                            {t('alert.severity.medium')}
                        </SelectItem>
                        <SelectItem value={AlertSeverity.High}>
                            {t('alert.severity.high')}
                        </SelectItem>
                        <SelectItem value={AlertSeverity.Critical}>
                            {t('alert.severity.critical')}
                        </SelectItem>
                    </SelectContent>
                </Select>

                <Select
                    value={statusFilter}
                    onValueChange={(val) => setStatusFilter(val === 'all' ? '' : val)}
                >
                    <SelectTrigger className="w-40 dark:text-shadow-glow-white">
                        <SelectValue
                            placeholder={t('alert.status', {
                                defaultValue: 'Status',
                            })}
                        />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">
                            {t('alert.allStatuses', { defaultValue: 'All' })}
                        </SelectItem>
                        <SelectItem value={AlertHistoryStatus.Pending}>
                            {AlertHistoryStatus.Pending}
                        </SelectItem>
                        <SelectItem value={AlertHistoryStatus.Sent}>
                            {AlertHistoryStatus.Sent}
                        </SelectItem>
                        <SelectItem value={AlertHistoryStatus.Failed}>
                            {AlertHistoryStatus.Failed}
                        </SelectItem>
                        <SelectItem value={AlertHistoryStatus.Acknowledged}>
                            {AlertHistoryStatus.Acknowledged}
                        </SelectItem>
                    </SelectContent>
                </Select>

                <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                        setSeverityFilter('')
                        setStatusFilter('')
                    }}
                >
                    {t('reset')}
                </Button>
            </div>

            <div>
                <DataTable table={table} columns={columns} />
                <div ref={sentinelRef} className="h-4" />
                {isFetchingNextPage && (
                    <div className="flex items-center justify-center p-4">
                        <Loader2 className="h-6 w-6 animate-spin text-primary" />
                    </div>
                )}
            </div>
        </div>
    )
}

