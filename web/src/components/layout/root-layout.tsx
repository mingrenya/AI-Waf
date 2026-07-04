import { Outlet } from 'react-router';
import { Sidebar } from './sidebar';
import { HeaderBar } from './header-bar';
import { PageHeader } from './breadcrumb';
import { useAuthStore } from '@/store/auth';
import { hasAnyPermission } from '@/lib/permissions';
import { ROUTES } from '@/routes/constants';
import { useMemo } from 'react';

export function RootLayout() {
  const user = useAuthStore((state) => state.user);

  const allowedSidebarItems = useMemo(() => {
    if (!user) return [];

    const mapping: { route: string; permissions: string[] }[] = [
      { route: ROUTES.MONITOR, permissions: ['system:status'] },
      { route: ROUTES.LOGS, permissions: ['waf:log:read'] },
      { route: ROUTES.RULES, permissions: ['config:read'] },
      { route: ROUTES.ALERTS, permissions: ['alert:channel:read'] },
      { route: ROUTES.SITUATION, permissions: ['waf:log:read'] },
      { route: ROUTES.NUCLEI, permissions: ['config:read'] },
      { route: ROUTES.CAPTURE, permissions: ['config:read'] },
      { route: ROUTES.AI_ANALYZER, permissions: ['waf:log:read'] },
      { route: ROUTES.SETTINGS, permissions: ['config:read'] },
    ];

    return mapping
      .filter((m) => hasAnyPermission(user, m.permissions))
      .map((m) => m.route);
  }, [user]);

  return (
    <div
      className="flex h-screen"
      style={{
        background: 'var(--bg-page)',
        backgroundAttachment: 'fixed',
      }}
    >
      <Sidebar allowedItems={allowedSidebarItems} />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <HeaderBar />
        <PageHeader />
        <main
          className="flex-1 overflow-auto p-6"
          style={{ background: 'transparent' }}
        >
          <Outlet />
        </main>
      </div>
    </div>
  );
}
