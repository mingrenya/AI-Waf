import { useMemo, useCallback } from 'react'
import {
    useReactTable,
    getCoreRowModel,
    ColumnDef,
} from '@tanstack/react-table'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { alertRuleApi } from '@/api/alert'
import { AlertRule, AlertSeverity } from '@/types/alert'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import {
    MoreHorizontal,
    Pencil,
    Trash2,
    AlertCircle,
    FileWarning,
} from 'lucide-react'
import { Loader2 } from 'lucide-react'
import { DataTable } from '@/components/table/motion-data-table'
import { useTranslation } from 'react-i18next'
import { toast } from '@/store'
import { queryKeys } from '@/lib/query-keys'

interface RuleTableProps {
    onEdit?: (rule: AlertRule) => void
    onDelete?: (id: string) => void
}

export function RuleTable({ onEdit, onDelete }: RuleTableProps) {
    const { t } = useTranslation()
    const queryClient = useQueryClient()

    // 获取规则列表（当前后端返回全量列表）
    const {
        data,
        isLoading,
        error,
    } = useQuery({
        queryKey: queryKeys.alert.rules.lists(),
        queryFn: () => alertRuleApi.getRules(),
    })

    const flatData = useMemo(
        () => (Array.isArray(data) ? data : []),
        [data],
    )

    // 启用状态切换（乐观更新）
    const toggleMutation = useMutation({
        mutationFn: (rule: AlertRule) =>
            alertRuleApi.updateRule(rule.id, { enabled: !rule.enabled }),
        onMutate: async (rule) => {
            await queryClient.cancelQueries({ queryKey: queryKeys.alert.rules.lists() })

            const previousData = queryClient.getQueryData<AlertRule[]>(queryKeys.alert.rules.lists())

            queryClient.setQueryData<AlertRule[]>(queryKeys.alert.rules.lists(), (old) =>
                old?.map((item) =>
                    item.id === rule.id ? { ...item, enabled: !item.enabled } : item,
                ) ?? [],
            )

            return { previousData }
        },
        onError: (_err, _rule, context) => {
            if (context?.previousData) {
                queryClient.setQueryData(queryKeys.alert.rules.lists(), context.previousData)
            }
            toast({
                title: t('common.error'),
                description: t('alert.updateFailed'),
                variant: 'destructive',
            })
        },
        onSuccess: () => {
            toast({
                title: t('common.success'),
                description: t('alert.updateSuccess'),
            })
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: queryKeys.alert.rules.lists() })
        },
    })

    const getSeverityLabel = useCallback((severity: AlertSeverity) => {
        return t(`alert.severity.${severity}`)
    }, [t])

    // 列定义
    const columns: ColumnDef<AlertRule>[] = useMemo(() => [
        {
            accessorKey: 'name',
            header: () => (
                <div className="font-medium py-3.5 whitespace-nowrap  dark:text-white">
                    {t('alert.ruleName')}
                </div>
            ),
            cell: ({ row }) => {
                if (!row.original) return null
                const isDisabled = !row.original.enabled
                return (
                    <div className={`font-medium  ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {row.original.name || '-'}
                    </div>
                )
            },
        },
        {
            accessorKey: 'severity',
            header: () => (
                <div className="font-medium py-3.5 whitespace-nowrap  dark:text-white">
                    {t('alert.severityLabel', { defaultValue: t('alert.severity.high') })}
                </div>
            ),
            cell: ({ row }) => {
                if (!row.original) return null
                const isDisabled = !row.original.enabled
                return (
                    <div className={` ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {getSeverityLabel(row.original.severity)}
                    </div>
                )
            },
        },
        {
            accessorKey: 'enabled',
            header: () => (
                <div className="font-medium py-3.5 whitespace-nowrap  dark:text-white">
                    {t('alert.status')}
                </div>
            ),
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
            accessorKey: 'cooldown',
            header: () => (
                <div className="font-medium py-3.5 whitespace-nowrap  dark:text-white">
                    {t('alert.cooldown', { defaultValue: 'Cooldown (min)' })}
                </div>
            ),
            cell: ({ row }) => {
                if (!row.original) return null
                const isDisabled = !row.original.enabled
                return (
                    <div className={` ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {row.original.cooldown ?? '-'}
                    </div>
                )
            },
        },
        {
            accessorKey: 'createdAt',
            header: () => (
                <div className="font-medium py-3.5 whitespace-nowrap  dark:text-white">
                    {t('createdAt')}
                </div>
            ),
            cell: ({ row }) => {
                if (!row.original) return null
                const isDisabled = !row.original.enabled
                return (
                    <div className={` ${isDisabled ? 'text-gray-400 dark:text-gray-500' : 'dark:text-slate-200'}`}>
                        {row.original.createdAt ? new Date(row.original.createdAt).toLocaleString() : '-'}
                    </div>
                )
            },
        },
        {
            id: 'actions',
            header: () => (
                <div className="font-medium py-3.5 whitespace-nowrap  dark:text-white">
                    {t('ipGroup.table.actions')}
                </div>
            ),
            cell: ({ row }) => {
                if (!row.original) return null
                return (
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 "
                                aria-label={t('ipGroup.table.actions')}
                            >
                                <MoreHorizontal className="h-4 w-4" />
                                <span className="sr-only">{t('ipGroup.table.actions')}</span>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="dark:bg-muted/95 dark:border-primary/20">
                            {onEdit && (
                                <DropdownMenuItem
                                    onClick={() => onEdit(row.original)}
                                    className=" dark:hover:bg-primary/20 cursor-pointer"
                                >
                                    <Pencil className="h-4 w-4 mr-2" />
                                    {t('certificate.edit')}
                                </DropdownMenuItem>
                            )}
                            {onDelete && (
                                <DropdownMenuItem
                                    onClick={() => onDelete(row.original.id)}
                                    className="text-red-600 dark:text-red-400 dark:hover:bg-red-500/20 cursor-pointer"
                                >
                                    <Trash2 className="h-4 w-4 mr-2" />
                                    {t('alert.deleteDialog.delete')}
                                </DropdownMenuItem>
                            )}
                        </DropdownMenuContent>
                    </DropdownMenu>
                )
            },
        },
    ], [t, getSeverityLabel, toggleMutation, onEdit, onDelete])

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
                <AlertCircle className="h-8 w-8 mb-2" />
                <p className="text-sm font-medium">{t('alert.loadFailed')}</p>
                <p className="text-xs text-muted-foreground mt-1">
                    {error instanceof Error ? error.message : String(error)}
                </p>
                <Button
                    onClick={() => queryClient.invalidateQueries({ queryKey: queryKeys.alert.rules.lists() })}
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
                <FileWarning className="h-8 w-8 mb-2" />
                <p className="text-sm">{t('alert.noRules', { defaultValue: 'No alert rules configured yet.' })}</p>
            </div>
        )
    }

    return (
        <div>
            <DataTable table={table} columns={columns} />
        </div>
    )
}

