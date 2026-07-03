import { Link, useLocation } from 'react-router'
import { ChevronRight } from 'lucide-react'
import { type BreadcrumbConfig, type RoutePath, useBreadcrumbMap } from '@/routes/config'

export function Breadcrumb() {
  const location = useLocation()
  const breadcrumbMap = useBreadcrumbMap()

  const segments = location.pathname.split('/').filter(Boolean)
  if (segments.length === 0) return null

  const parentPath = ('/' + segments[0]) as RoutePath
  const config: BreadcrumbConfig | undefined = breadcrumbMap[parentPath]

  const currentItem = config?.items?.find((item) => item.path === segments[1])
  const parentLabel = getParentLabel(parentPath)

  return (
    <nav
      className="px-6 py-3 flex items-center gap-2 text-sm z-10"
      style={{
        background: 'rgba(255, 255, 255, 0.06)',
        backdropFilter: 'blur(20px)',
        WebkitBackdropFilter: 'blur(20px)',
        borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
      }}
    >
      <Link to={parentPath} className="text-white/60 hover:text-white/90 transition-colors">
        {parentLabel}
      </Link>
      {currentItem && (
        <>
          <ChevronRight className="w-3 h-3 text-white/30" />
          <span className="text-white font-medium">{currentItem.title}</span>
        </>
      )}
      <div className="ml-auto text-xs text-white/40 font-mono">
        {new Date().toLocaleTimeString()}
      </div>
    </nav>
  )
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
  }
  return labels[path] ?? path
}
