import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { DataTable } from "@/components/table/data-table"
import {
    ColumnDef,
    useReactTable,
    getCoreRowModel,
} from "@tanstack/react-table"
import { MoreHorizontal, Eye, Trash2, Shield, RefreshCw } from "lucide-react"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAttackPatterns, useDeleteAttackPattern } from "../hooks"
import type { AttackPattern } from "@/types/ai-analyzer"
import { format } from "date-fns"
import { AttackPatternDetailDialog } from "./AttackPatternDetailDialog"
import { Card } from "@/components/ui/card"
import { useQueryClient } from "@tanstack/react-query"

export function AttackPatternTable() {
    const [page, _setPage] = useState(1)
    const [size, _setSize] = useState(20)
    const [detailDialogOpen, setDetailDialogOpen] = useState(false)
    const [selectedPattern, setSelectedPattern] = useState<AttackPattern | null>(null)
    
    const { data, isLoading, refetch } = useAttackPatterns({ page, size })
    const deleteMutation = useDeleteAttackPattern()
    const queryClient = useQueryClient()

    const handleView = (pattern: AttackPattern) => {
        setSelectedPattern(pattern)
        setDetailDialogOpen(true)
    }

    const handleDelete = (id: string) => {
        if (confirm("确定要删除这个攻击模式吗？")) {
            deleteMutation.mutate(id)
        }
    }

    const handleRefresh = () => {
        queryClient.invalidateQueries({ queryKey: ["attack-patterns"] })
        refetch()
    }

    const columns: ColumnDef<AttackPattern>[] = [
        {
            accessorKey: "attack_type",
            header: "攻击类型",
            cell: ({ row }) => (
                <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium">{row.original.attack_type}</span>
                </div>
            ),
        },
        {
            accessorKey: "severity",
            header: "严重程度",
            cell: ({ row }) => {
                const severity = row.original.severity
                const variants: Record<string, any> = {
                    critical: "destructive",
                    high: "destructive",
                    medium: "default",
                    low: "secondary",
                }
                return <Badge variant={variants[severity] || "outline"}>{severity}</Badge>
            },
        },
        {
            accessorKey: "sample_count",
            header: "样本数量",
            cell: ({ row }) => <span>{row.original.sample_count}</span>,
        },
        {
            accessorKey: "statistical_data",
            header: "Z-Score",
            cell: ({ row }) => (
                <span className="font-mono text-sm">
                    {row.original.statistical_data?.z_score?.toFixed(2) || "N/A"}
                </span>
            ),
        },
        {
            accessorKey: "detected_at",
            header: "检测时间",
            cell: ({ row }) => (
                <span className="text-sm text-muted-foreground">
                    {format(new Date(row.original.detected_at), "yyyy-MM-dd HH:mm")}
                </span>
            ),
        },
        {
            id: "actions",
            header: "操作",
            cell: ({ row }) => (
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => handleView(row.original)}>
                            <Eye className="h-4 w-4 mr-2" />
                            查看详情
                        </DropdownMenuItem>
                        <DropdownMenuItem
                            onClick={() => handleDelete(row.original.id)}
                            className="text-destructive"
                        >
                            <Trash2 className="h-4 w-4 mr-2" />
                            删除
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            ),
        },
    ]

    const table = useReactTable({
        data: data?.list || [],
        columns,
        getCoreRowModel: getCoreRowModel(),
    })

    return (
        <Card className="p-6 w-full min-h-full border-none shadow-none rounded-none">
            <div className="flex justify-between items-center mb-6 bg-zinc-50 dark:bg-muted/30 rounded-md p-4">
                <div>
                    <h2 className="text-xl font-semibold text-primary dark:text-white">攻击模式检测</h2>
                    <p className="text-sm text-muted-foreground mt-1">
                        基于机器学习算法自动检测的攻击模式
                    </p>
                </div>
                <Button variant="outline" size="sm" onClick={handleRefresh}>
                    <RefreshCw className="h-4 w-4 mr-2" />
                    刷新
                </Button>
            </div>

            <DataTable
                table={table}
                columns={columns}
                isLoading={isLoading}
                loadingStyle="skeleton"
            />

            <AttackPatternDetailDialog
                open={detailDialogOpen}
                onOpenChange={setDetailDialogOpen}
                pattern={selectedPattern}
            />
        </Card>
    )
}
