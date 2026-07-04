import { cn } from '@/lib/utils'

interface StatusBarProps {
  online?: boolean
  lastSync?: string
}

export function StatusBar({ online = true, lastSync }: StatusBarProps) {
  return (
    <div className={cn(
      'flex items-center justify-between px-4 py-2.5 rounded-xl text-sm border',
      online
        ? 'bg-emerald-500/5 border-emerald-500/20 text-emerald-400'
        : 'bg-red-500/5 border-red-500/20 text-red-400',
    )}>
      <div className="flex items-center gap-2">
        {online ? (
          <>
            <span className="status-dot-online" />
            <span className="font-medium">All Systems Operational</span>
          </>
        ) : (
          <>
            <span className="status-dot-critical" />
            <span className="font-medium">System Degraded</span>
          </>
        )}
      </div>
      {lastSync && (
        <span className="text-xs font-mono" style={{color:'var(--text-primary)', opacity:0.4}}>
          Last sync: {lastSync}
        </span>
      )}
    </div>
  )
}
