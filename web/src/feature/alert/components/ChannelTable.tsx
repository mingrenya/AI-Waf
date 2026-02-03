import { useRef, useEffect, useMemo, useCallback } from 'react'
import {
    useReactTable,
    getCoreRowModel,
    ColumnDef,
} from '@tanstack/react-table'
import { useInfiniteQuery } from '@tanstack/react-query'
import { alertChannelApi } from '@/api/alert'
import { AlertChannel, AlertChannelType } from '@/types/alert'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
    MoreHorizontal,
    Pencil,
    Trash2,
    Radio,
    TestTube2,
    Webhook,
    MessageSquare,
    AlertCircle
} from 'lucide-react'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Loader2 } from 'lucide-react'
import { DataTable } from '@/components/table/motion-data-table'
import { useTranslation } from 'react-i18next'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@/hooks/use-toast'

interface ChannelTableProps {
    onEdit: (channel: AlertChannel) => void
    onDelete: (id: string) => void
    onTest: (channel: AlertChannel) => void
}

// 分页数据类型定义
interface PaginatedResponse {
    items: AlertChannel[]
    total: number
    page: number
}

interface InfiniteQueryData {
    pages: PaginatedResponse[]
    pageParams: unknown[]
}

// 配置常量
const PAGINATION_CONFIG = {
    pageSize: 20,
    intersectionThreshold: 0.1,
    intersectionRootMargin: '100px 0px',
    retryAttempts: 3,
    retryDelayMax: 30000,
} as const

export function ChannelTable({ onEdit, onDelete, onTest }: ChannelTableProps) {
    const { t } = useTranslation()
    const { toast } = useToast()
    const queryClient = useQueryClient()
    const sentinelRef = useRef<HTMLDivElement>(null)

    // 获取通道列表
    const {
        data,
        isLoading,
        error,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useInfiniteQuery({
        queryKey: ['alertChannels'],
        queryFn: ({ pageParam = 1 }) => {
            const page = typeof pageParam === 'number' ? pageParam : 1
            return alertChannelApi.getChannels(page, PAGINATION_CONFIG.pageSize)
        },
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            if (!lastPage || typeof lastPage.total === 'undefined') {
                return undefined
            }

            if (!allPages || allPages.length === 0) {
                return undefined
            }

            // 安全访问page.items防止undefined错误
            const fetchedItemsCount = allPages.reduce(
                (total, page) => total + (page?.items?.length || 0), 
                0
            )
            return fetchedItemsCount < lastPage.total ? allPages.length + 1 : undefined
        },
        retry: PAGINATION_CONFIG.retryAttempts,
        retryDelay: (attemptIndex) => 
            Math.min(1000 * 2 ** attemptIndex, PAGINATION_CONFIG.retryDelayMax),
    })

    // 扁平化分页数据 - 添加严格的数据验证
    const flatData = useMemo(() => {
        if (!data?.pages) return []
        
        return data.pages
            .filter((page): page is NonNullable<typeof page> => page != null && page.items != null)
            .flatMap(page => page.items)
            .filter((item): item is AlertChannel => {
                // 确保必需字段存在
                return item != null && 
                       typeof item.id === 'string' && 
                       typeof item.enabled === 'boolean'
            })
    }, [data])

    // 切换启用状态 - 使用乐观更新防止race condition
    const toggleMutation = useMutation({
        mutationFn: (channel: AlertChannel) => 
            alertChannelApi.updateChannel(channel.id, { enabled: !channel.enabled }),
        onMutate: async (channel) => {
            // 取消进行中的查询避免覆盖乐观更新
            await queryClient.cancelQueries({ queryKey: ['alertChannels'] })
            
            // 保存当前数据用于回滚
            const previousData = queryClient.getQueryData<InfiniteQueryData>(['alertChannels'])
            
            // 乐观更新UI
            queryClient.setQueryData<InfiniteQueryData>(['alertChannels'], (old) => {
                if (!old?.pages) return old
                return {
                    ...old,
                    pages: old.pages.map((page) => ({
                        ...page,
                        items: page.items?.map((item: AlertChannel) =>
                            item.id === channel.id 
                                ? { ...item, enabled: !item.enabled } 
                                : item
                        ) || []
                    }))
                }
            })
            
            return { previousData }
        },
        onError: (_err, _channel, context) => {
            // 出错时回滚到之前的状态
            if (context?.previousData) {
                queryClient.setQueryData<InfiniteQueryData>(['alertChannels'], context.previousData)
            }
            toast({ 
                title: t('common.error'), 
                description: t('alert.updateFailed'), 
                variant: 'destructive' 
            })
        },
        onSuccess: () => {
            toast({ 
                title: t('common.success'), 
                description: t('alert.updateSuccess') 
            })
        },
        onSettled: () => {
            // 无论成功失败都重新获取确保数据同步
            queryClient.invalidateQueries({ queryKey: ['alertChannels'] })
        }
    })

    // 无限滚动 - 优化清理逻辑
    useEffect(() => {
        if (!hasNextPage || !sentinelRef.current) return

        const options = {
            threshold: PAGINATION_CONFIG.intersectionThreshold,
            rootMargin: PAGINATION_CONFIG.intersectionRootMargin
        }

        const handleObserver = (entries: IntersectionObserverEntry[]) => {
            const [entry] = entries
            if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
                fetchNextPage()
            }
        }

        const observer = new IntersectionObserver(handleObserver, options)
        const sentinel = sentinelRef.current
        
        observer.observe(sentinel)

        return () => {
            // disconnect会自动清理所有观察目标
            observer.disconnect()
        }
    }, [hasNextPage, isFetchingNextPage, fetchNextPage])

    // 获取通道类型图标 - 使用useCallback优化
    const getChannelIcon = useCallback((type: AlertChannelType) => {
        switch (type) {
            case AlertChannelType.Webhook:
                return <Webhook className="h-3.5 w-3.5" />
            case AlertChannelType.Slack:
            case AlertChannelType.Discord:
            case AlertChannelType.DingTalk:
            case AlertChannelType.WeCom:
                return <MessageSquare className="h-3.5 w-3.5" />
            default:
                return <Radio className="h-3.5 w-3.5" />
        }
    }, [])

    // 获取通道类型标签 - 使用useCallback优化
    const getChannelTypeLabel = useCallback((type: AlertChannelType) => {
        return t(`alert.channelType.${type}`)
    }, [t])

    // 表格列定义 - 使用useMemo优化避免重复创建
    const columns: ColumnDef<AlertChannel>[] = useMemo(() => [
        {
            accessorKey: 'name',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('alert.channelName')}</div>,
            cell: ({ row }) => {
                if (!row.original) return null
                const isDisabled = !row.original.enabled
                return (
                    <div className={`font-medium dark:text-shadow-glow-white ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {row.original.name || '-'}
                    </div>
                )
            },
        },
        {
            accessorKey: 'type',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('alert.channels')}</div>,
            cell: ({ row }) => {
                if (!row.original) return null
                const isDisabled = !row.original.enabled
                return (
                    <div className={`flex items-center gap-1 dark:text-shadow-glow-white ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {getChannelIcon(row.original.type)}
                        <span>{getChannelTypeLabel(row.original.type)}</span>
                    </div>
                )
            },
        },
        {
            accessorKey: 'enabled',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('alert.status')}</div>,
            cell: ({ row }) => {
                if (!row.original) return null
                return (
                    <Switch
                        checked={row.original.enabled}
                        onCheckedChange={() => toggleMutation.mutate(row.original)}
                        disabled={toggleMutation.isPending}
                        className="dark:data-[state=checked]:bg-primary"
                        aria-label={t('alert.toggleStatus')}
                    />
                )
            },
        },
        {
            accessorKey: 'createdAt',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('createdAt')}</div>,
            cell: ({ row }) => {
                if (!row.original) return null
                const isDisabled = !row.original.enabled
                return (
                    <div className={`dark:text-shadow-glow-white ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {row.original.createdAt ? new Date(row.original.createdAt).toLocaleString() : '-'}
                    </div>
                )
            },
        },
        {
            id: 'actions',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('ipGroup.table.actions')}</div>,
            cell: ({ row }) => {
                if (!row.original) return null
                return (
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 dark:text-shadow-glow-white"
                                aria-label={t('ipGroup.table.actions')}
                            >
                                <MoreHorizontal className="h-4 w-4" />
                                <span className="sr-only">{t('ipGroup.table.actions')}</span>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="dark:bg-muted/95 dark:border-primary/20">
                            <DropdownMenuItem
                                onClick={() => onTest(row.original)}
                                className="dark:text-shadow-glow-white dark:hover:bg-primary/20 cursor-pointer"
                            >
                                <TestTube2 className="h-4 w-4 mr-2" />
                                {t('alert.testChannel')}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                                onClick={() => onEdit(row.original)}
                                className="dark:text-shadow-glow-white dark:hover:bg-primary/20 cursor-pointer"
                            >
                                <Pencil className="h-4 w-4 mr-2" />
                                {t('certificate.edit')}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                                onClick={() => onDelete(row.original.id)}
                                className="text-red-600 dark:text-red-400 dark:hover:bg-red-500/20 cursor-pointer"
                            >
                                <Trash2 className="h-4 w-4 mr-2" />
                                {t('alert.deleteDialog.delete')}
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                )
            },
        },
    ], [t, toggleMutation, getChannelIcon, getChannelTypeLabel, onTest, onEdit, onDelete])

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

    // 错误状态处理
    if (error) {
        return (
            <div className="flex flex-col items-center justify-center p-8 text-destructive">
                <AlertCircle className="h-8 w-8 mb-2" />
                <p className="text-sm font-medium">{t('alert.loadFailed')}</p>
                <p className="text-xs text-muted-foreground mt-1">
                    {error instanceof Error ? error.message : String(error)}
                </p>
                <Button 
                    onClick={() => queryClient.invalidateQueries({ queryKey: ['alertChannels'] })}
                    variant="outline"
                    size="sm"
                    className="mt-4"
                >
                    {t('common.retry')}
                </Button>
            </div>
        )
    }

    // 空状态处理
    if (!isLoading && flatData.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center p-8 text-muted-foreground">
                <Radio className="h-8 w-8 mb-2" />
                <p className="text-sm">{t('alert.noChannels')}</p>
            </div>
        )
    }

    return (
        <div>
            <DataTable table={table} columns={columns} />
            <div ref={sentinelRef} className="h-4" />
            {isFetchingNextPage && (
                <div className="flex items-center justify-center p-4">
                    <Loader2 className="h-6 w-6 animate-spin text-primary" />
                </div>
            )}
        </div>
    )
}
