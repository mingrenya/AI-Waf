import { Link, useLocation } from 'react-router';
import { cn } from '@/lib/utils';
import { type ComponentType } from 'react';
import {
  Settings, Shield, BarChart2, FileText, Bell, Brain,
  ScanSearch, Crosshair, Radio,
} from 'lucide-react';
import { ROUTES } from '@/routes/constants';
import { useTranslation } from 'react-i18next';

interface SidebarItem {
  title: string;
  icon: ComponentType<{ className?: string }>;
  href: string;
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
  ];
}

interface SidebarProps {
  allowedItems?: string[];
}

export function Sidebar({ allowedItems }: SidebarProps) {
  const location = useLocation();
  const { t } = useTranslation();

  const currentFirstLevelPath = '/' + location.pathname.split('/')[1];

  const items = createSidebarItems(t).filter((item) => {
    if (!allowedItems) return true;
    return allowedItems.includes(item.href);
  });

  return (
    <aside
      className="h-full flex flex-col surface-root"
      style={{
        width: '220px',
        minWidth: '220px',
        borderRight: '1px solid var(--surface-root-border)',
      }}
    >
      {/* 导航菜单 */}
      <nav className="flex-1 py-4 space-y-1 overflow-y-auto scrollbar-glass">
        {items.map((item) => {
          const isActive = currentFirstLevelPath === item.href;
          return (
            <Link
              key={item.href}
              to={item.href}
              className={cn(
                'flex items-center gap-3 px-4 py-2.5 mx-2 rounded-md text-sm font-medium transition-colors duration-150 relative',
              )}
              style={{
                color: isActive
                  ? 'var(--color-primary-5)'
                  : 'var(--text-secondary)',
                background: isActive ? 'var(--color-primary-1)' : 'transparent',
              }}
            >
              {/* Active 指示竖线 */}
              {isActive && (
                <div
                  className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full"
                  style={{ background: 'var(--color-primary-5)' }}
                />
              )}
              <item.icon className="w-5 h-5 flex-shrink-0" />
              <span>{item.title}</span>
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
