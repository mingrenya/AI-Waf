import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router'
import { LoginForm } from '@/feature/auth/components/LoginForm'
import useAuthStore from '@/store/auth'
import { useTranslation } from 'react-i18next'
import { Heart, Shield, Activity, Globe, Server } from 'lucide-react'

const features = [
  { icon: Shield, label: 'WAF Engine', value: 'Coraza 3.0', color: 'text-indigo-300' },
  { icon: Activity, label: 'Threats Blocked', value: '12.4K', color: 'text-emerald-300' },
  { icon: Globe, label: 'Protected Sites', value: '8', color: 'text-sky-300' },
  { icon: Server, label: 'Uptime', value: '99.9%', color: 'text-amber-300' },
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
        <div className="glass-bg-animated min-h-screen flex flex-col relative overflow-hidden">
            {/* Floating particles */}
            <div className="absolute inset-0 pointer-events-none overflow-hidden">
                <div className="glass-particle w-72 h-72 top-[10%] left-[5%]" style={{ animationDelay: '0s' }} />
                <div className="glass-particle w-96 h-96 top-[40%] right-[10%]" style={{ animationDelay: '5s' }} />
                <div className="glass-particle w-64 h-64 bottom-[15%] left-[30%]" style={{ animationDelay: '10s' }} />
                <div className="glass-particle w-48 h-48 top-[20%] right-[30%]" style={{ animationDelay: '15s' }} />
            </div>

            {/* Status bar */}
            <div className="glass-nav text-white text-center py-1.5 text-sm font-medium flex items-center justify-center gap-2">
                <span className="status-dot-online" />
                {t('sidebar.title')} — Secure Access
            </div>

            <div className="flex-1 flex items-center justify-center p-6 relative z-10">
                <div className="w-full max-w-5xl grid grid-cols-1 lg:grid-cols-2 gap-8">
                    {/* Left: Product showcase */}
                    <div className="hidden lg:flex flex-col justify-center animate-slide-up">
                        <div className="mb-10">
                            <h1 className="text-5xl font-bold text-white mb-3 font-mono tracking-tight stagger-1">
                                MRYa<span className="text-indigo-300">WAF</span>
                            </h1>
                            <p className="text-white/70 text-lg stagger-2">
                                智能 Web 应用防火墙管理平台
                            </p>
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            {features.map((f, i) => (
                                <div key={f.label} className={`glass-card-light p-5 animate-scale-in stagger-${i + 3}`}>
                                    <f.icon className={`w-5 h-5 ${f.color} mb-2`} />
                                    <p className="text-xs text-white/50 mb-1">{f.label}</p>
                                    <p className={`text-xl font-bold font-mono ${f.color}`}>{f.value}</p>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Right: Login form */}
                    <div className="flex items-center justify-center animate-slide-in-right">
                        <div className="w-full max-w-md glass-card-emphasis p-8">
                            <div className="text-center mb-6">
                                <h2 className="text-2xl font-bold text-white">{t('auth.login')}</h2>
                                <p className="text-white/50 text-sm mt-1">{t('auth.loginDescription')}</p>
                            </div>
                            <LoginForm />
                        </div>
                    </div>
                </div>
            </div>

            {/* Bottom status */}
            <div className="py-3 text-center text-xs text-white/40 border-t border-white/10 relative z-10 flex items-center justify-center gap-1">
                System Status: <span className="text-emerald-400">All Systems Operational</span>
                <span className="mx-2 opacity-30">|</span>
                <span>Made with</span>
                <Heart className="h-3 w-3 text-red-500 fill-red-500" />
                <span>by</span>
                <a href="https://github.com/mingrenya/AI-Waf" target="_blank" rel="noopener noreferrer" className="text-white/60 hover:text-white/90 transition-colors">MRYa WAF team</a>
            </div>
        </div>
    )
}
