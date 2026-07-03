import { useEffect, useMemo, useState } from 'react'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import type { CreateUserRequest, ManagedUser, UpdateUserRequest } from '@/types/user-management'

const roleDescriptions: Record<string, string> = {
    admin: '超级管理员 — 拥有所有权限，可管理用户、站点、配置、告警等全部功能',
    configurator: '配置管理员 — 可管理站点、证书、规则、告警配置，但不能管理用户',
    auditor: '审计员 — 只读访问所有日志、配置、用户信息，不能进行任何修改',
    user: '普通用户 — 仅可查看 WAF 日志和系统状态，权限最小',
}

interface UserDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    mode: 'create' | 'edit'
    user?: ManagedUser | null
    submitting?: boolean
    onSubmit: (payload: CreateUserRequest | UpdateUserRequest) => Promise<void>
}

export function UserDialog({ open, onOpenChange, mode, user, submitting = false, onSubmit }: UserDialogProps) {
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [role, setRole] = useState<'admin' | 'auditor' | 'configurator' | 'user'>('user')
    const [needReset, setNeedReset] = useState(true)

    useEffect(() => {
        if (!open) return

        if (mode === 'edit' && user) {
            setUsername(user.username)
            setPassword('')
            setRole(user.role)
            setNeedReset(user.needReset)
            return
        }

        setUsername('')
        setPassword('')
        setRole('user')
        setNeedReset(true)
    }, [mode, open, user])

    const title = useMemo(() => {
        return mode === 'create' ? '新增用户' : '编辑用户'
    }, [mode])

    const handleSubmit = async () => {
        if (!username.trim()) {
            return
        }

        if (mode === 'create') {
            if (password.length < 6) {
                return
            }

            const createPayload: CreateUserRequest = {
                username: username.trim(),
                password,
                role,
            }
            await onSubmit(createPayload)
            return
        }

        const updatePayload: UpdateUserRequest = {
            username: username.trim(),
            role,
            needReset,
        }
        if (password.trim()) {
            updatePayload.password = password
        }
        await onSubmit(updatePayload)
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="bg-white/10 backdrop-blur-xl border border-white/20 rounded-2xl">
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                </DialogHeader>

                <div className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="username">用户名</Label>
                        <Input
                            id="username"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="请输入用户名"
                        />
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="role">角色</Label>
                        <Select value={role} onValueChange={(value) => setRole(value as 'admin' | 'auditor' | 'configurator' | 'user')}>
                            <SelectTrigger id="role">
                                <SelectValue placeholder="选择角色" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="admin">管理员</SelectItem>
                                <SelectItem value="auditor">审计员</SelectItem>
                                <SelectItem value="configurator">配置管理员</SelectItem>
                                <SelectItem value="user">普通用户</SelectItem>
                            </SelectContent>
                        </Select>
                        {role && (
                            <p className="text-xs text-white/50 mt-1 italic">{roleDescriptions[role]}</p>
                        )}
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="password">{mode === 'create' ? '初始密码' : '新密码（留空表示不修改）'}</Label>
                        <Input
                            id="password"
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder={mode === 'create' ? '至少 6 位' : '留空不修改'}
                        />
                    </div>

                    {mode === 'edit' && (
                        <div className="flex items-center justify-between rounded-md border p-3">
                            <div>
                                <p className="text-sm font-medium">下次登录强制重置密码</p>
                                <p className="text-xs text-muted-foreground">开启后用户下次登录必须修改密码。</p>
                            </div>
                            <Switch checked={needReset} onCheckedChange={setNeedReset} />
                        </div>
                    )}
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)} disabled={submitting}>
                        取消
                    </Button>
                    <Button onClick={handleSubmit} disabled={submitting || !username.trim() || (mode === 'create' && password.length < 6)}>
                        {submitting ? '提交中...' : mode === 'create' ? '创建' : '保存'}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
