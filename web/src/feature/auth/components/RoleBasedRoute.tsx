import { Outlet } from 'react-router'
import { useAuthStore } from '@/store/auth'
import { hasPermission } from '@/lib/permissions'
import ForbiddenPage from '@/pages/forbidden/page'

interface RoleBasedRouteProps {
  requiredPermission?: string
}

/**
 * Role/permission-based route guard.
 * When requiredPermission is set, checks if the current user has that permission.
 * If they don't, renders ForbiddenPage (403) instead of the child routes.
 * Wrap specific route groups that need permission checks.
 */
export function RoleBasedRoute({ requiredPermission }: RoleBasedRouteProps) {
  const user = useAuthStore((state) => state.user)

  if (requiredPermission && !hasPermission(user, requiredPermission)) {
    return <ForbiddenPage />
  }

  return <Outlet />
}
