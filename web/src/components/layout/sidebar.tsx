'use client'

import { Link, useLocation, useNavigate } from 'react-router'
import { cn } from '@/lib/utils'
import { Settings, Shield, BarChart2, FileText, LogOut, Bell, Brain, ScanSearch, Crosshair, Radio, ChevronLeft, ChevronRight } from 'lucide-react'
import { ROUTES } from '@/routes/constants'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/store/auth'
import { useState } from 'react'

interface SidebarItem {
  title: string
  icon: React.ComponentType<{ className?: string }>
  href: string
}

function createSidebarItems(t: (key: string) => string): SidebarItem[] {
  return [
    { title: t('sidebar.monitor'), icon: BarChart2, href: ROUTES.MONITOR },
    { title: t('sidebar.logs'), icon: FileText, href: ROUTES.LOGS },
    { title: t('sidebar.rules'), icon: Shield, href: ROUTES.RULES },
    { title: t('sidebar.alerts'), icon: Bell, href: ROUTES.ALERTS },
    { title: t('sidebar.situation'), icon: ScanSearch, href: ROUTES.SITUATION },
    { title: t('sidebar.nuclei'), icon: Crosshair, href: ROUTES.NUCLEI },
    { title: t('sidebar.capture'), icon: Radio, href: ROUTES.CAPTURE },
    { title: t('sidebar.aiAnalyzer'), icon: Brain, href: ROUTES.AI_ANALYZER },
    { title: t('sidebar.settings'), icon: Settings, href: ROUTES.SETTINGS },
  ]
}

interface SidebarProps {
  allowedItems?: string[]
}

export function Sidebar({ allowedItems }: SidebarProps) {
  const location = useLocation()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { logout, user } = useAuthStore()
  const [collapsed, setCollapsed] = useState(false)

  const currentFirstLevelPath = '/' + location.pathname.split('/')[1]

  const items = createSidebarItems(t).filter((item) => {
    if (!allowedItems) return true
    return allowedItems.includes(item.href)
  })

  return (
    <aside
      className={cn(
        'h-full text-white flex flex-col transition-all duration-300 relative',
        collapsed ? 'w-16' : 'w-60',
      )}
      style={{
        background: 'rgba(255, 255, 255, 0.06)',
        backdropFilter: 'blur(30px)',
        WebkitBackdropFilter: 'blur(30px)',
        borderRight: '1px solid rgba(255, 255, 255, 0.1)',
      }}
    >
      {/* Collapse toggle */}
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="absolute -right-3 top-20 w-6 h-6 rounded-full flex items-center justify-center hover:scale-110 transition-transform z-10"
        style={{ background: 'rgba(255,255,255,0.15)', border: '1px solid rgba(255,255,255,0.2)' }}
      >
        {collapsed ? <ChevronRight className="w-3 h-3 text-white/70" /> : <ChevronLeft className="w-3 h-3 text-white/70" />}
      </button>

      {/* Logo */}
      <div className={cn('flex items-center gap-3 py-5 border-b border-white/10', collapsed ? 'px-3 justify-center' : 'px-5')}>
        <div className="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0"
             style={{ background: 'linear-gradient(135deg, #667eea, #764ba2)' }}>
          <Shield className="w-5 h-5 text-white" />
        </div>
        {!collapsed && (
          <div>
            <p className="font-bold text-white text-sm tracking-tight">MRYa WAF</p>
            <p className="text-xs text-white/40">Security Console</p>
          </div>
        )}
      </div>

      {/* User info */}
      {!collapsed && user && (
        <div className="px-5 py-3 border-b border-white/10">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-full flex items-center justify-center text-xs font-medium text-white flex-shrink-0"
                 style={{ background: 'linear-gradient(135deg, #667eea, #764ba2)' }}>
              {user.username.charAt(0).toUpperCase()}
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-sm text-white truncate">{user.username}</p>
              <p className="text-xs text-white/50 truncate">{user.role}</p>
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 py-3 space-y-0.5 overflow-y-auto scrollbar-glass">
        {items.map((item) => {
          const isActive = currentFirstLevelPath === item.href
          return (
            <Link
              key={item.href}
              to={item.href}
              title={collapsed ? item.title : undefined}
              className={cn(
                'flex items-center gap-3 font-medium transition-all duration-200 mx-2 rounded-xl',
                collapsed ? 'px-3 py-3 justify-center' : 'px-4 py-2.5',
                isActive
                  ? 'text-white'
                  : 'text-white/60 hover:text-white/90',
              )}
              style={isActive ? {
                background: 'rgba(102, 126, 234, 0.2)',
                border: '1px solid rgba(102, 126, 234, 0.3)',
              } : {}}
            >
              <item.icon className={cn('w-5 h-5 flex-shrink-0', isActive && 'text-indigo-300')} />
              {!collapsed && <span className="text-sm">{item.title}</span>}
              {isActive && !collapsed && (
                <div className="ml-auto w-1.5 h-1.5 rounded-full bg-indigo-300" />
              )}
            </Link>
          )
        })}
      </nav>

      {/* Bottom: Logout */}
      <div className={cn('border-t border-white/10 py-3', collapsed ? 'px-3' : 'px-4')}>
        <button
          onClick={() => { logout(); navigate('/login') }}
          className={cn(
            'flex items-center gap-3 text-white/50 hover:text-red-300 transition-colors duration-200 w-full rounded-xl',
            'hover:bg-white/5',
            collapsed ? 'px-3 py-3 justify-center' : 'px-4 py-2.5',
          )}
          title={collapsed ? t('sidebar.logout') : undefined}
        >
          <LogOut className="w-5 h-5" />
          {!collapsed && <span className="text-sm">{t('sidebar.logout')}</span>}
        </button>

        {!collapsed && (
          <div className="text-center text-xs text-white/30 mt-4 px-4">
            MRYa WAF v1.0
          </div>
        )}
      </div>
    </aside>
  )
}
