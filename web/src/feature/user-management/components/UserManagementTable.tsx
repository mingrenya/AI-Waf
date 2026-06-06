import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { MoreHorizontal, Pencil, Plus, RefreshCcw, Trash2, UserCog } from 'lucide-react'
import { useAuthStore } from '@/store/auth'
import { userManagementApi } from '@/api/user-management'
import type { CreateUserRequest, ManagedUser, UpdateUserRequest } from '@/types/user-management'
import { hasPermission } from '@/lib/permissions'
import { useToast } from '@/hooks/use-toast'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import { UserDialog } from './UserDialog'
import { DeleteUserDialog } from './DeleteUserDialog'

const QUERY_KEY = ['users']

function roleLabel(role: ManagedUser['role']) {
    switch (role) {
        case 'admin':
            return '管理员'
        case 'auditor':
            return '审计员'
        case 'configurator':
            return '配置管理员'
        default:
            return '普通用户'
    }
}

function roleBadgeClass(role: ManagedUser['role']) {
    switch (role) {
        case 'admin':
            return 'bg-red-100 text-red-700 border-red-200 dark:bg-red-900/30 dark:text-red-300'
        case 'auditor':
            return 'bg-blue-100 text-blue-700 border-blue-200 dark:bg-blue-900/30 dark:text-blue-300'
        case 'configurator':
            return 'bg-purple-100 text-purple-700 border-purple-200 dark:bg-purple-900/30 dark:text-purple-300'
        default:
            return 'bg-slate-100 text-slate-700 border-slate-200 dark:bg-slate-900/30 dark:text-slate-300'
    }
}

export function UserManagementTable() {
    const { toast } = useToast()
    const queryClient = useQueryClient()
    const currentUser = useAuthStore((state) => state.user)

    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)
    const [dialogOpen, setDialogOpen] = useState(false)
    const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
    const [selectedUser, setSelectedUser] = useState<ManagedUser | null>(null)
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

    const canRead = hasPermission(currentUser, 'user:read')
    const canCreate = hasPermission(currentUser, 'user:create')
    const canUpdate = hasPermission(currentUser, 'user:update')
    const canDelete = hasPermission(currentUser, 'user:delete')

    const { data: users = [], isLoading, refetch } = useQuery({
        queryKey: QUERY_KEY,
        queryFn: userManagementApi.getUsers,
        enabled: canRead,
    })

    const createMutation = useMutation({
        mutationFn: (payload: CreateUserRequest) => userManagementApi.createUser(payload),
        onSuccess: () => {
            toast({ title: '创建成功', description: '用户已创建。' })
            setDialogOpen(false)
            queryClient.invalidateQueries({ queryKey: QUERY_KEY })
        },
        onError: (error) => {
            const message = error instanceof Error ? error.message : '创建用户失败'
            toast({ variant: 'destructive', title: '创建失败', description: message })
        },
    })

    const updateMutation = useMutation({
        mutationFn: ({ id, payload }: { id: string; payload: UpdateUserRequest }) => userManagementApi.updateUser(id, payload),
        onSuccess: () => {
            toast({ title: '更新成功', description: '用户信息已更新。' })
            setDialogOpen(false)
            queryClient.invalidateQueries({ queryKey: QUERY_KEY })
        },
        onError: (error) => {
            const message = error instanceof Error ? error.message : '更新用户失败'
            toast({ variant: 'destructive', title: '更新失败', description: message })
        },
    })

    const deleteMutation = useMutation({
        mutationFn: (id: string) => userManagementApi.deleteUser(id),
        onSuccess: () => {
            toast({ title: '删除成功', description: '用户已删除。' })
            setDeleteDialogOpen(false)
            setSelectedUser(null)
            queryClient.invalidateQueries({ queryKey: QUERY_KEY })
        },
        onError: (error) => {
            const message = error instanceof Error ? error.message : '删除用户失败'
            toast({ variant: 'destructive', title: '删除失败', description: message })
        },
    })

    const sortedUsers = useMemo(() => {
        return [...users].sort((a, b) => {
            if (a.role === 'admin') return -1
            if (b.role === 'admin') return 1
            return a.username.localeCompare(b.username)
        })
    }, [users])

    const refreshUsers = () => {
        setIsRefreshAnimating(true)
        void refetch().finally(() => {
            setTimeout(() => setIsRefreshAnimating(false), 600)
        })
    }

    const openCreateDialog = () => {
        setDialogMode('create')
        setSelectedUser(null)
        setDialogOpen(true)
    }

    const openEditDialog = (user: ManagedUser) => {
        setDialogMode('edit')
        setSelectedUser(user)
        setDialogOpen(true)
    }

    const openDeleteDialog = (user: ManagedUser) => {
        setSelectedUser(user)
        setDeleteDialogOpen(true)
    }

    const handleSubmit = async (payload: CreateUserRequest | UpdateUserRequest) => {
        if (dialogMode === 'create') {
            await createMutation.mutateAsync(payload as CreateUserRequest)
            return
        }

        if (!selectedUser) {
            return
        }
        await updateMutation.mutateAsync({ id: selectedUser.id, payload: payload as UpdateUserRequest })
    }

    const handleDelete = async () => {
        if (!selectedUser) {
            return
        }
        await deleteMutation.mutateAsync(selectedUser.id)
    }

    if (!canRead) {
        return (
            <Card className="p-6 border-none shadow-none rounded-none">
                <div className="rounded-lg border border-dashed p-8 text-center">
                    <UserCog className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
                    <p className="text-sm text-muted-foreground">当前账号没有用户管理查看权限（需要 user:read）。</p>
                </div>
            </Card>
        )
    }

    return (
        <>
            <Card className="p-6 border-none shadow-none rounded-none flex flex-col h-full">
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-semibold text-primary dark:text-white">用户管理</h2>
                    <div className="flex gap-2">
                        <Button variant="outline" size="sm" onClick={refreshUsers} className="flex items-center gap-2">
                            <RefreshCcw className={isRefreshAnimating ? 'h-4 w-4 animate-spin' : 'h-4 w-4'} />
                            刷新
                        </Button>
                        {canCreate && (
                            <Button size="sm" onClick={openCreateDialog} className="flex items-center gap-1">
                                <Plus className="h-3.5 w-3.5" />
                                新增用户
                            </Button>
                        )}
                    </div>
                </div>

                <div className="overflow-auto rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>用户名</TableHead>
                                <TableHead>角色</TableHead>
                                <TableHead>状态</TableHead>
                                <TableHead>最后登录</TableHead>
                                <TableHead>创建时间</TableHead>
                                <TableHead className="w-[80px]">操作</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {isLoading ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center text-muted-foreground py-8">加载中...</TableCell>
                                </TableRow>
                            ) : sortedUsers.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center text-muted-foreground py-8">暂无用户数据</TableCell>
                                </TableRow>
                            ) : (
                                sortedUsers.map((user) => (
                                    <TableRow key={user.id}>
                                        <TableCell className="font-medium">{user.username}</TableCell>
                                        <TableCell>
                                            <Badge variant="outline" className={roleBadgeClass(user.role)}>
                                                {roleLabel(user.role)}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant={user.needReset ? 'outline' : 'secondary'}>
                                                {user.needReset ? '需重置密码' : '正常'}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>{user.lastLogin ? new Date(user.lastLogin).toLocaleString('zh-CN') : '-'}</TableCell>
                                        <TableCell>{user.createdAt ? new Date(user.createdAt).toLocaleString('zh-CN') : '-'}</TableCell>
                                        <TableCell>
                                            {(canUpdate || canDelete) ? (
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <Button variant="ghost" size="icon">
                                                            <MoreHorizontal className="h-4 w-4" />
                                                        </Button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent align="end">
                                                        {canUpdate && (
                                                            <DropdownMenuItem onClick={() => openEditDialog(user)}>
                                                                <Pencil className="mr-2 h-4 w-4" />
                                                                编辑
                                                            </DropdownMenuItem>
                                                        )}
                                                        {canDelete && currentUser?.id !== user.id && (
                                                            <DropdownMenuItem onClick={() => openDeleteDialog(user)} className="text-red-600 dark:text-red-400">
                                                                <Trash2 className="mr-2 h-4 w-4" />
                                                                删除
                                                            </DropdownMenuItem>
                                                        )}
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            ) : '-'}
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </div>
            </Card>

            <UserDialog
                open={dialogOpen}
                onOpenChange={setDialogOpen}
                mode={dialogMode}
                user={selectedUser}
                submitting={createMutation.isPending || updateMutation.isPending}
                onSubmit={handleSubmit}
            />

            <DeleteUserDialog
                open={deleteDialogOpen}
                onOpenChange={setDeleteDialogOpen}
                username={selectedUser?.username}
                submitting={deleteMutation.isPending}
                onConfirm={handleDelete}
            />
        </>
    )
}
