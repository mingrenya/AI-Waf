import { useNavigate } from 'react-router';
import { Sun, Moon, Bot, Shield, LogOut, User } from 'lucide-react';
import { useThemeStore, useAuthStore } from '@/store';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useTranslation } from 'react-i18next';
import { ROUTES } from '@/routes/constants';

export function HeaderBar() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { resolvedTheme, setTheme } = useThemeStore();
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header
      className="h-12 flex items-center justify-between px-4 flex-shrink-0 surface-root border-b z-10"
      style={{ borderBottom: '1px solid var(--surface-root-border)' }}
    >
      {/* Left: Logo + Product Name */}
      <div className="flex items-center gap-2">
        <div
          className="w-7 h-7 rounded-md flex items-center justify-center"
          style={{ background: 'linear-gradient(135deg, var(--color-primary-5), var(--color-primary-9))' }}
        >
          <Shield className="w-4 h-4" style={{color:'#ffffff'}} />
        </div>
        <span className="font-bold text-sm" style={{ color: 'var(--text-primary)' }}>
          MRYa WAF
        </span>
      </div>

      {/* Center: System Status */}
      <div className="flex items-center gap-2">
        <span className="status-dot-online" />
        <span className="text-xs" style={{ color: 'var(--text-muted)' }}>
          {t('system.status', { defaultValue: 'System Normal' })}
        </span>
      </div>

      {/* Right: Action Area */}
      <div className="flex items-center gap-1">
        {/* Theme Toggle */}
        <button
          onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
          className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-black/5 dark:hover:bg-white/10 transition-colors"
          title={resolvedTheme === 'dark' ? 'Switch to light' : 'Switch to dark'}
          style={{ color: 'var(--text-secondary)' }}
        >
          {resolvedTheme === 'dark' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
        </button>

        {/* AI Assistant */}
        <button
          onClick={() => navigate(ROUTES.AI_ANALYZER + '/assistant')}
          className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-black/5 dark:hover:bg-white/10 transition-colors"
          title={t('sidebar.aiAssistant', { defaultValue: 'AI Assistant' })}
          style={{ color: 'var(--text-secondary)' }}
        >
          <Bot className="w-4 h-4" />
        </button>

        {/* User Avatar */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium text-white ml-1"
              style={{ background: 'linear-gradient(135deg, var(--color-primary-5), var(--color-primary-9))' }}
            >
              {user?.username?.charAt(0).toUpperCase() ?? 'U'}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48 surface-floating">
            <DropdownMenuLabel>
              <div className="flex flex-col gap-1">
                <span style={{ color: 'var(--text-primary)' }} className="text-sm">{user?.username}</span>
                <span style={{ color: 'var(--text-muted)' }} className="text-xs font-normal">{user?.role}</span>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => navigate('/settings/user')} className="cursor-pointer">
              <User className="w-4 h-4 mr-2" />
              {t('auth.profile', { defaultValue: 'Profile' })}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleLogout} className="cursor-pointer text-red-500">
              <LogOut className="w-4 h-4 mr-2" />
              {t('sidebar.logout')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
