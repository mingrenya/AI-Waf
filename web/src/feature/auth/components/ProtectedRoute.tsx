import { useLocation, Navigate, Outlet } from 'react-router'
import useAuthStore from '@/store/auth'

export function ProtectedRoute() {
    const { isAuthenticated, needPasswordReset } = useAuthStore()
    const location = useLocation()

    if (!isAuthenticated) {
        return <Navigate to="/login" state={{ from: location }} replace />
    }

    if (needPasswordReset && location.pathname !== '/reset-password') {
        return <Navigate to="/reset-password" replace />
    }

    return <Outlet />
}