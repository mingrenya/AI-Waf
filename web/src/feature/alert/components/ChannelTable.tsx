import { useRef, useEffect, useMemo } from 'react'
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
    MessageSquare
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

export function ChannelTable({ onEdit, onDelete, onTest }: ChannelTableProps) {
    const { t } = useTranslation()
    const { toast } = useToast()
    const queryClient = useQueryClient()
    const sentinelRef = useRef<HTMLDivElement>(null)
    const PAGE_SIZE = 20

    // 获取通道列表
    const {
        data,
        isLoading,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useInfiniteQuery({
        queryKey: ['alertChannels'],
        queryFn: ({ pageParam }) => alertChannelApi.getChannels(pageParam as number, PAGE_SIZE),
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            const fetchedItemsCount = allPages.reduce((total, page) => total + page.items.length, 0)
            return fetchedItemsCount < lastPage.total ? allPages.length + 1 : undefined
        },
    })

    // 扁平化分页数据
    const flatData = useMemo(() =>
        data?.pages.flatMap(page => page.items) || [],
        [data]
    )

    // 切换启用状态
    const toggleMutation = useMutation({
        mutationFn: (channel: AlertChannel) => 
            alertChannelApi.updateChannel(channel.id, { enabled: !channel.enabled }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['alertChannels'] })
            toast({ title: 'Success', description: t('alert.updateSuccess') })
        },
        onError: () => {
            toast({ title: 'Error', description: t('alert.updateFailed'), variant: 'destructive' })
        }
    })

    // 无限滚动
    useEffect(() => {
        if (!hasNextPage) return

        const options = {
            threshold: 0.1,
            rootMargin: '100px 0px'
        }

        const handleObserver = (entries: IntersectionObserverEntry[]) => {
            const [entry] = entries
            if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
                fetchNextPage()
            }
        }

        const observer = new IntersectionObserver(handleObserver, options)
        const sentinel = sentinelRef.current
        if (sentinel) {
            observer.observe(sentinel)
        }

        return () => {
            if (sentinel) {
                observer.unobserve(sentinel)
            }
            observer.disconnect()
        }
    }, [hasNextPage, isFetchingNextPage, fetchNextPage])

    // 获取通道类型图标
    const getChannelIcon = (type: AlertChannelType) => {
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
    }

    // 获取通道类型标签
    const getChannelTypeLabel = (type: AlertChannelType) => {
        return t(`alert.channelType.${type}`)
    }

    // 表格列定义
    const columns: ColumnDef<AlertChannel>[] = [
        {
            accessorKey: 'name',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('alert.channelName')}</div>,
            cell: ({ row }) => {
                const isDisabled = !row.original.enabled
                return (
                    <div className={`font-medium dark:text-shadow-glow-white ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {row.original.name}
                    </div>
                )
            },
        },
        {
            accessorKey: 'type',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('alert.channels')}</div>,
            cell: ({ row }) => {
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
            cell: ({ row }) => (
                <Switch
                    checked={row.original.enabled}
                    onCheckedChange={() => toggleMutation.mutate(row.original)}
                    className="dark:data-[state=checked]:bg-primary"
                />
            ),
        },
        {
            accessorKey: 'createdAt',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('createdAt')}</div>,
            cell: ({ row }) => {
                const isDisabled = !row.original.enabled
                return (
                    <div className={`dark:text-shadow-glow-white ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {new Date(row.original.createdAt).toLocaleString()}
                    </div>
                )
            },
        },
        {
            id: 'actions',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('ipGroup.table.actions')}</div>,
            cell: ({ row }) => (
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 dark:text-shadow-glow-white"
                        >
                            <MoreHorizontal className="h-4 w-4" />
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
            ),
        },
    ]

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
