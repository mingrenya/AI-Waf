import { Card, CardContent, CardTitle } from "@/components/ui/card"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { ReactNode } from "react"
import { useNavigate } from "react-router"

interface StatsCardProps {
    title: string
    value: string | number
    icon?: ReactNode
    change?: string | number
    trend?: 'up' | 'down' | 'neutral'
    loading?: boolean
    isTraffic?: boolean
    link?: string
    iconTone?: 'default' | 'inbound' | 'outbound' | 'danger'
}

export function StatsCard({
    title,
    value,
    icon,
    change,
    trend,
    loading = false,
    isTraffic = false,
    link,
    iconTone = 'default',
}: StatsCardProps) {
    const navigate = useNavigate()

    // 用于格式化流量数据
    const formatTraffic = (bytes: number): string => {
        if (bytes < 1024) return `${bytes} B`
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
        if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
        return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
    }

    // 颜色处理
    const getTrendColor = () => {
        if (!trend) return 'text-muted-foreground'
        return trend === 'up'
            ? 'text-emerald-500 dark:text-emerald-400'
            : trend === 'down'
                ? 'text-red-500 dark:text-red-400'
                : 'text-muted-foreground'
    }

    // Format value for display
    const displayValue = isTraffic && typeof value === 'number'
        ? formatTraffic(value)
        : String(value)

    const handleCardClick = () => {
        if (!link) return
        
        // 检查是否为外部链接
        if (link.startsWith('http://') || link.startsWith('https://')) {
            window.open(link, '_blank')
        } else {
            // 内部路由跳转
            navigate(link)
        }
    }

    const iconToneClass = {
        default: 'bg-primary/10 text-primary border-primary/20',
        inbound: 'bg-sky-500/10 text-sky-500 border-sky-500/20 dark:text-sky-300',
        outbound: 'bg-violet-500/10 text-violet-500 border-violet-500/20 dark:text-violet-300',
        danger: 'bg-red-500/10 text-red-500 border-red-500/20 dark:text-red-300',
    }[iconTone]

    return (
        <Card 
            className={`border p-4 transition-colors duration-150 ${link ? 'cursor-pointer' : ''}`}
            style={{
                background: 'var(--surface-card-bg)',
                borderColor: 'var(--surface-card-border)',
            }}
            onClick={link ? handleCardClick : undefined}
        >
            <CardTitle className="text-sm font-medium text-muted-foreground mb-2 flex items-center gap-2">
                {icon && (
                    <span className={`inline-flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-md border ${iconToneClass}`}>
                        {icon}
                    </span>
                )}
                {title}
            </CardTitle>

            <CardContent className="p-0">
                {loading ? (
                    <div className="h-9 w-24 animate-pulse bg-blue-100 dark:bg-white/10 rounded"></div>
                ) : (
                    <div className="flex flex-col">
                        <TooltipProvider delayDuration={300}>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <div className="text-2xl font-bold truncate">
                                        {displayValue}
                                    </div>
                                </TooltipTrigger>
                                <TooltipContent
                                    className="max-w-[350px] break-all bg-white border border-slate-200 shadow-md py-2 px-3 text-sm text-slate-800 dark:bg-slate-800 dark:border-slate-700 dark:!text-slate-200"
                                    side="top"
                                >
                                    {displayValue}
                                </TooltipContent>
                            </Tooltip>
                        </TooltipProvider>

                        {change && (
                            <TooltipProvider delayDuration={300}>
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <div className={`text-xs ${getTrendColor()} flex items-center mt-1 truncate`}>
                                            {trend === 'up' && '↑ '}
                                            {trend === 'down' && '↓ '}
                                            {change}
                                        </div>
                                    </TooltipTrigger>
                                    <TooltipContent
                                        className="max-w-[350px] break-all bg-white border border-slate-200 shadow-md py-2 px-3 text-sm text-slate-800 dark:bg-slate-800 dark:border-slate-700 dark:!text-slate-200"
                                        side="top"
                                    >
                                        {trend === 'up' ? '↑ ' : trend === 'down' ? '↓ ' : ''}{change}
                                    </TooltipContent>
                                </Tooltip>
                            </TooltipProvider>
                        )}
                    </div>
                )}
            </CardContent>
        </Card>
    )
}
