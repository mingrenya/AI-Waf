import { Link, useLocation } from 'react-router';
import { ChevronRight } from 'lucide-react';
import { type RoutePath, useBreadcrumbMap, type BreadcrumbConfig } from '@/routes/config';
import type { ReactNode } from 'react';

interface PageHeaderProps {
  actions?: ReactNode;
}

export function PageHeader({ actions }: PageHeaderProps) {
  const location = useLocation();
  const breadcrumbMap = useBreadcrumbMap();

  const segments = location.pathname.split('/').filter(Boolean);
  if (segments.length === 0) return null;

  const parentPath = ('/' + segments[0]) as RoutePath;
  const config: BreadcrumbConfig | undefined = breadcrumbMap[parentPath];
  const currentItem = config?.items?.find((item) => item.path === segments[1]);
  const parentLabel = getParentLabel(parentPath);

  return (
    <nav
      className="h-10 px-6 flex items-center gap-2 text-sm flex-shrink-0"
      style={{
        borderBottom: '1px solid var(--surface-root-border)',
        background: 'var(--surface-root-bg)',
      }}
    >
      <Link
        to={parentPath}
        style={{ color: 'var(--text-muted)' }}
        className="hover:underline transition-colors"
      >
        {parentLabel}
      </Link>
      {currentItem && (
        <>
          <ChevronRight className="w-3 h-3" style={{ color: 'var(--text-dim)' }} />
          <span className="font-medium" style={{ color: 'var(--text-primary)' }}>
            {currentItem.title}
          </span>
        </>
      )}
      {actions && (
        <div className="ml-auto flex items-center gap-2">
          {actions}
        </div>
      )}
    </nav>
  );
}

function getParentLabel(path: string): string {
  const labels: Record<string, string> = {
    '/monitor': '监控',
    '/logs': '日志',
    '/rules': '规则',
    '/alerts': '告警',
    '/settings': '设置',
    '/ai-analyzer': 'AI 分析器',
    '/nuclei': 'Nuclei',
    '/capture': '流量捕获',
    '/situation': '态势感知',
    '/ftw': 'FTW 测试',
  };
  return labels[path] ?? path;
}
