import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router'
import { LoginForm } from '@/feature/auth/components/LoginForm'
import useAuthStore from '@/store/auth'
import { useTranslation } from 'react-i18next'
import { Heart, Shield, Activity, Globe, Server } from 'lucide-react'

const features = [
  { icon: Shield, label: 'WAF Engine', value: 'Coraza 3.0', color: 'text-primary-400' },
  { icon: Activity, label: 'Threats Blocked', value: '12.4K', color: 'text-primary-300' },
  { icon: Globe, label: 'Protected Sites', value: '8', color: 'text-primary-300' },
  { icon: Server, label: 'Uptime', value: '99.9%', color: 'text-primary-300' },
]

export default function LoginPage() {
    const { isAuthenticated, needPasswordReset } = useAuthStore()
    const navigate = useNavigate()
    const location = useLocation()
    const { t } = useTranslation()

    // Get the redirect path from location state, or default to '/'
    const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/'

    useEffect(() => {
        // If already authenticated
        if (isAuthenticated) {
            // If needs password reset, redirect to reset page
            if (needPasswordReset) {
                navigate('/reset-password')
            } else {
                // Otherwise, redirect to the page they tried to access or home
                navigate(from)
            }
        }
    }, [isAuthenticated, needPasswordReset, navigate, from])

    return (
        <div className="dark glass-bg-animated min-h-screen flex flex-col relative overflow-hidden">
            {/* Floating particles */}
            <div className="absolute inset-0 pointer-events-none overflow-hidden">
                <div className="glass-particle w-72 h-72 top-[10%] left-[5%]" style={{ animationDelay: '0s' }} />
                <div className="glass-particle w-96 h-96 top-[40%] right-[10%]" style={{ animationDelay: '5s' }} />
                <div className="glass-particle w-64 h-64 bottom-[15%] left-[30%]" style={{ animationDelay: '10s' }} />
                <div className="glass-particle w-48 h-48 top-[20%] right-[30%]" style={{ animationDelay: '15s' }} />
            </div>

            {/* Status bar */}
            <div className="surface-card text-center py-1.5 text-sm font-medium flex items-center justify-center gap-2" style={{ color: 'var(--text-primary)' }}>
                <span className="status-dot-online" />
                {t('sidebar.title')} — Secure Access
            </div>

            <div className="flex-1 flex items-center justify-center p-6 relative z-10">
                <div className="w-full max-w-5xl grid grid-cols-1 lg:grid-cols-2 gap-8">
                    {/* Left: Product showcase */}
                    <div className="hidden lg:flex flex-col justify-center animate-slide-up">
                        <div className="mb-10">
                            <h1 className="text-5xl font-bold mb-3 font-mono tracking-tight stagger-1" style={{ color: 'var(--text-primary)' }}>
                                MRYa<span className="text-primary-300">WAF</span>
                            </h1>
                            <p className="text-lg stagger-2" style={{ color: 'var(--text-secondary)' }}>
                                智能 Web 应用防火墙管理平台
                            </p>
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            {features.map((f, i) => (
                                <div key={f.label} className={`surface-card p-5 animate-scale-in stagger-${i + 3}`}>
                                    <f.icon className={`w-5 h-5 ${f.color} mb-2`} />
                                    <p className="text-xs mb-1" style={{ color: 'var(--text-muted)' }}>{f.label}</p>
                                    <p className={`text-xl font-bold font-mono ${f.color}`}>{f.value}</p>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Right: Login form */}
                    <div className="flex items-center justify-center animate-slide-in-right">
                        <div className="w-full max-w-md surface-card p-8" style={{ borderRadius: 'var(--radius-lg)' }}>
                            <div className="text-center mb-6">
                                <h2 className="text-2xl font-bold" style={{ color: 'var(--text-primary)' }}>{t('auth.login')}</h2>
                                <p className="text-sm mt-1" style={{ color: 'var(--text-muted)' }}>{t('auth.loginDescription')}</p>
                            </div>
                            <LoginForm />
                        </div>
                    </div>
                </div>
            </div>

            {/* Bottom status */}
            <div className="py-3 text-center text-xs border-t relative z-10 flex items-center justify-center gap-1" style={{ color: 'var(--text-dim)', borderColor: 'rgba(255,255,255,0.1)' }}>
                System Status: <span className="text-emerald-400">All Systems Operational</span>
                <span className="mx-2 opacity-30">|</span>
                <span>Made with</span>
                <Heart className="h-3 w-3 text-red-500 fill-red-500" />
                <span>by</span>
                <a href="https://github.com/mingrenya/AI-Waf" target="_blank" rel="noopener noreferrer" className="transition-colors" style={{ color: 'var(--text-secondary)' }}>MRYa WAF team</a>
            </div>
        </div>
    )
}
