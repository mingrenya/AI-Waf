import { useEffect } from 'react'
import { useNavigate, useLocation, Outlet } from 'react-router' // 修改导入路径
import useAuthStore from '@/store/auth'

export function ProtectedRoute() {
    const { isAuthenticated, needPasswordReset } = useAuthStore()
    const navigate = useNavigate()
    const location = useLocation()

    useEffect(() => {
        if (!isAuthenticated) {
            // Redirect to login, but save the current location
            navigate('/login', { state: { from: location } })
        } else if (needPasswordReset && location.pathname !== '/reset-password') {
            // If user needs to reset password, force them to the reset page
            navigate('/reset-password')
        }
    }, [isAuthenticated, needPasswordReset, navigate, location])

    // If authenticated and doesn't need reset (or is on reset page), render children
    return isAuthenticated ? <Outlet /> : null
}