import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { DataTable } from "@/components/table/data-table"
import {
    ColumnDef,
    useReactTable,
    getCoreRowModel,
} from "@tanstack/react-table"
import { MoreHorizontal, Eye, Trash2, CheckCircle2, Rocket, RefreshCw } from "lucide-react"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useGeneratedRules, useDeleteGeneratedRule, useDeployRule } from "../hooks"
import type { GeneratedRule } from "@/types/ai-analyzer"
import { format } from "date-fns"
import { RuleReviewDialog } from "./RuleReviewDialog"
import { RuleDetailDialog } from "./RuleDetailDialog"
import { Card } from "@/components/ui/card"
import { useQueryClient } from "@tanstack/react-query"

export function GeneratedRuleTable() {
    const [page, _setPage] = useState(1)
    const [size, _setSize] = useState(20)
    const [reviewDialogOpen, setReviewDialogOpen] = useState(false)
    const [detailDialogOpen, setDetailDialogOpen] = useState(false)
    const [selectedRule, setSelectedRule] = useState<GeneratedRule | null>(null)
    
    const { data, isLoading, refetch } = useGeneratedRules({ page, size })
    const deleteMutation = useDeleteGeneratedRule()
    const deployMutation = useDeployRule()
    const queryClient = useQueryClient()

    const handleView = (rule: GeneratedRule) => {
        setSelectedRule(rule)
        setDetailDialogOpen(true)
    }

    const handleReview = (rule: GeneratedRule) => {
        setSelectedRule(rule)
        setReviewDialogOpen(true)
    }

    const handleDeploy = (id: string) => {
        if (confirm("确定要部署这条规则吗？")) {
            deployMutation.mutate(id)
        }
    }

    const handleDelete = (id: string) => {
        if (confirm("确定要删除这条规则吗？")) {
            deleteMutation.mutate(id)
        }
    }

    const handleRefresh = () => {
        queryClient.invalidateQueries({ queryKey: ["generated-rules"] })
        refetch()
    }

    const getStatusBadge = (status: string) => {
        switch (status) {
            case "pending":
                return <Badge variant="secondary">待审核</Badge>
            case "approved":
                return <Badge variant="default">已批准</Badge>
            case "rejected":
                return <Badge variant="destructive">已拒绝</Badge>
            case "deployed":
                return <Badge variant="outline" className="border-green-500 text-green-600">已部署</Badge>
            default:
                return <Badge>{status}</Badge>
        }
    }

    const columns: ColumnDef<GeneratedRule>[] = [
        {
            accessorKey: "rule_type",
            header: "规则类型",
            cell: ({ row }) => (
                <span className="font-medium">{row.original.rule_type}</span>
            ),
        },
        {
            accessorKey: "confidence",
            header: "置信度",
            cell: ({ row }) => (
                <span className="font-mono text-sm">
                    {(row.original.confidence * 100).toFixed(1)}%
                </span>
            ),
        },
        {
            accessorKey: "status",
            header: "状态",
            cell: ({ row }) => getStatusBadge(row.original.status),
        },
        {
            accessorKey: "created_at",
            header: "创建时间",
            cell: ({ row }) => (
                <span className="text-sm text-muted-foreground">
                    {format(new Date(row.original.created_at), "yyyy-MM-dd HH:mm")}
                </span>
            ),
        },
        {
            id: "actions",
            header: "操作",
            cell: ({ row }) => {
                const rule = row.original
                return (
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                                <MoreHorizontal className="h-4 w-4" />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleView(rule)}>
                                <Eye className="h-4 w-4 mr-2" />
                                查看详情
                            </DropdownMenuItem>
                            {rule.status === "pending" && (
                                <DropdownMenuItem onClick={() => handleReview(rule)}>
                                    <CheckCircle2 className="h-4 w-4 mr-2" />
                                    审核
                                </DropdownMenuItem>
                            )}
                            {rule.status === "approved" && (
                                <DropdownMenuItem onClick={() => handleDeploy(rule.id)}>
                                    <Rocket className="h-4 w-4 mr-2" />
                                    部署
                                </DropdownMenuItem>
                            )}
                            <DropdownMenuItem
                                onClick={() => handleDelete(rule.id)}
                                className="text-destructive"
                            >
                                <Trash2 className="h-4 w-4 mr-2" />
                                删除
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                )
            },
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
                    <h2 className="text-xl font-semibold text-primary dark:text-white">AI生成规则</h2>
                    <p className="text-sm text-muted-foreground mt-1">
                        自动生成的ModSecurity和MicroRule防护规则
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

            <RuleReviewDialog
                open={reviewDialogOpen}
                onOpenChange={setReviewDialogOpen}
                rule={selectedRule}
            />

            <RuleDetailDialog
                open={detailDialogOpen}
                onOpenChange={setDetailDialogOpen}
                rule={selectedRule}
            />
        </Card>
    )
}
