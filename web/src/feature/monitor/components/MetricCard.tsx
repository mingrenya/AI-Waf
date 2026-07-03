import { cn } from '@/lib/utils'
import { NumberTicker } from '@/components/ui/animation/number-ticker'
import { type LucideIcon } from 'lucide-react'

interface MetricCardProps {
  label: string
  value: number
  trend?: {
    value: number
    isPositive: boolean
  }
  icon?: LucideIcon
  color?: 'cyan' | 'red' | 'yellow' | 'blue' | 'emerald' | 'purple' | 'orange'
}

const colorClasses: Record<NonNullable<MetricCardProps['color']>, { border: string; bg: string; text: string; icon: string }> = {
  cyan:     { border: 'border-cyan-500/20', bg: 'bg-cyan-500/5', text: 'text-cyan-400', icon: 'text-cyan-400' },
  red:      { border: 'border-red-500/20', bg: 'bg-red-500/5', text: 'text-red-400', icon: 'text-red-400' },
  yellow:   { border: 'border-yellow-500/20', bg: 'bg-yellow-500/5', text: 'text-yellow-400', icon: 'text-yellow-400' },
  blue:     { border: 'border-blue-500/20', bg: 'bg-blue-500/5', text: 'text-blue-400', icon: 'text-blue-400' },
  emerald:  { border: 'border-emerald-500/20', bg: 'bg-emerald-500/5', text: 'text-emerald-400', icon: 'text-emerald-400' },
  purple:   { border: 'border-purple-500/20', bg: 'bg-purple-500/5', text: 'text-purple-400', icon: 'text-purple-400' },
  orange:   { border: 'border-orange-500/20', bg: 'bg-orange-500/5', text: 'text-orange-400', icon: 'text-orange-400' },
}

export function MetricCard({ label, value, trend, icon: Icon, color = 'cyan' }: MetricCardProps) {
  const c = colorClasses[color]

  return (
    <div className={cn(
      'glass-card p-5 rounded-xl border backdrop-blur-sm transition-all duration-300 hover:shadow-lg',
      c.border, c.bg,
    )}>
      <div className="flex items-start justify-between mb-3">
        <p className="text-xs font-medium text-white/50 uppercase tracking-wider">{label}</p>
        {Icon && <Icon className={cn('h-5 w-5', c.icon)} />}
      </div>
      <div className="flex items-baseline gap-2">
        <NumberTicker value={value} className={cn('text-3xl font-bold font-mono', c.text)} />
        {trend && (
          <span className={cn(
            'text-xs font-medium',
            trend.isPositive ? 'text-emerald-400' : 'text-red-400',
          )}>
            {trend.isPositive ? '+' : '-'}{trend.value}%
          </span>
        )}
      </div>
    </div>
  )
}
