import React from 'react'

interface StatCardProps {
    title: string
    value: number | string
    className?: string
}

/**
 * 统计卡片组件
 * 用于显示统计数据，左上角布局，无图标
 */
export const StatCard: React.FC<StatCardProps> = ({
    title,
    value,
    className = ''
}) => {
    // 格式化数字显示
    const formatValue = (val: number | string): string => {
        if (typeof val === 'string') return val

        if (val >= 1000000) {
            return `${(val / 1000000).toFixed(1)}M`
        } else if (val >= 1000) {
            return `${(val / 1000).toFixed(1)}K`
        }
        return val.toString()
    }

    return (
        <div className={`p-2 ${className}`}>
            <div className="text-xs text-white/70 mb-1 font-medium tracking-wide uppercase text-shadow-glow-purple">
                {title}
            </div>
            <div className="text-xl font-bold text-white">
                {formatValue(value)}
            </div>
        </div>
    )
} 