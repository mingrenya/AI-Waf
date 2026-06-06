import type { User } from '@/types/auth'

const ROLE_PERMISSIONS: Record<string, string[]> = {
    admin: [
        'user:create', 'user:read', 'user:update', 'user:delete',
        'site:create', 'site:read', 'site:update', 'site:delete',
        'config:read', 'config:update',
        'audit:read', 'system:restart', 'system:status', 'waf:log:read',
        'cert:create', 'cert:read', 'cert:update', 'cert:delete',
        'alert:channel:create', 'alert:channel:read', 'alert:channel:update', 'alert:channel:delete',
        'alert:rule:create', 'alert:rule:read', 'alert:rule:update', 'alert:rule:delete',
        'alert:history:read',
    ],
    auditor: [
        'user:read',
        'site:read',
        'config:read',
        'audit:read',
        'system:status',
        'waf:log:read',
        'cert:read',
        'alert:channel:read',
        'alert:rule:read',
        'alert:history:read',
    ],
    configurator: [
        'site:create', 'site:read', 'site:update', 'site:delete',
        'config:read', 'config:update',
        'system:status',
        'waf:log:read',
        'cert:read', 'cert:update', 'cert:delete',
        'alert:channel:create', 'alert:channel:read', 'alert:channel:update', 'alert:channel:delete',
        'alert:rule:create', 'alert:rule:read', 'alert:rule:update', 'alert:rule:delete',
        'alert:history:read',
    ],
    user: [
        'site:read',
        'system:status',
        'waf:log:read',
        'cert:read',
        'alert:channel:read',
        'alert:rule:read',
        'alert:history:read',
    ],
}

export function getEffectivePermissions(user: User | null | undefined): string[] {
    if (!user) {
        return []
    }

    const rolePerms = ROLE_PERMISSIONS[user.role] ?? []
    const extraPerms = user.permissions ?? []

    return Array.from(new Set([...rolePerms, ...extraPerms]))
}

export function hasPermission(user: User | null | undefined, permission: string): boolean {
    if (!user) {
        return false
    }
    if (user.role === 'admin') {
        return true
    }

    return getEffectivePermissions(user).includes(permission)
}

export function hasAnyPermission(user: User | null | undefined, permissions: string[]): boolean {
    return permissions.some((permission) => hasPermission(user, permission))
}
