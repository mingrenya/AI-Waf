import React from 'react'
import { useTranslation } from 'react-i18next'
import { AttackEventAggregateResult } from '@/types/log'
import { Clock } from 'lucide-react'

interface RealtimeAttackListProps {
    realtimeAttacks: AttackEventAggregateResult[]
    isLoading?: boolean
}

/**
 * 实时攻击列表组件
 * 显示实时Web攻击信息
 */
export const RealtimeAttackList: React.FC<RealtimeAttackListProps> = ({
    realtimeAttacks,
    isLoading = false
}) => {
    const { t, i18n } = useTranslation()

    // 格式化地理位置信息
    const formatLocation = (srcIpInfo: AttackEventAggregateResult['srcIpInfo']) => {
        if (!srcIpInfo) return t('attackDetail.noLocationInfo')

        const parts = []

        if (srcIpInfo.country) {
            parts.push(i18n.language === 'zh' ? srcIpInfo.country.nameZh : srcIpInfo.country.nameEn)
        }

        if (srcIpInfo.subdivision) {
            parts.push(i18n.language === 'zh' ? srcIpInfo.subdivision.nameZh : srcIpInfo.subdivision.nameEn)
        }

        if (srcIpInfo.city) {
            parts.push(i18n.language === 'zh' ? srcIpInfo.city.nameZh : srcIpInfo.city.nameEn)
        }

        return parts.join(' - ')
    }

    // 格式化时间显示
    const formatTime = (timeString: string) => {
        const date = new Date(timeString)
        return date.toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        })
    }

    return (
        <div className="h-full p-3">
            <div className="mb-3">
                <div className="text-xs font-semibold flex items-center gap-2" style={{color:'var(--text-primary)'}}>
                    <span className="tracking-wide">{t('securityDashboard.realtimeAttacks.title')}</span>
                </div>
            </div>
            <div className="space-y-2">
                {isLoading ? (
                    <div className="space-y-2">
                        {[...Array(5)].map((_, i) => (
                            <div key={i} className="animate-pulse">
                                <div className="flex items-center gap-2">
                                    <div className="w-2 h-2 rounded-full" style={{background:'var(--text-primary)', opacity:0.2}}></div>
                                    <div className="flex-1">
                                        <div className="h-2 rounded w-3/4 mb-1" style={{background:'var(--text-primary)', opacity:0.2}}></div>
                                        <div className="h-2 rounded w-1/2" style={{background:'var(--text-primary)', opacity:0.1}}></div>
                                    </div>
                                    <div className="h-2 rounded w-6" style={{background:'var(--text-primary)', opacity:0.2}}></div>
                                </div>
                            </div>
                        ))}
                    </div>
                ) : realtimeAttacks.length === 0 ? (
                    <div className="text-left py-2 text-xs" style={{color:'var(--text-secondary)'}}>
                        {t('noResult')}
                    </div>
                ) : (
                    <div className="space-y-2">
                        {realtimeAttacks.map((attack, index) => (
                            <div key={`${attack.srcIp}-${index}`} className="flex items-start gap-2 py-1">
                                {/* 状态指示器 */}
                                <div className="flex-shrink-0 mt-1">
                                    <div
                                        className={`w-2 h-2 rounded-full ${attack.isOngoing ? 'bg-red-500 animate-pulse' : ''
                                            }`}
                                            style={attack.isOngoing ? undefined : {background:'var(--text-primary)', opacity:0.4}}
                                    ></div>
                                </div>

                                {/* 攻击信息 */}
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center justify-between mb-0.5">
                                        <div className="font-mono text-xs font-medium truncate" style={{color:'var(--text-primary)'}}>
                                            {attack.srcIp}
                                        </div>
                                        <div className="text-xs font-bold ml-2" style={{color:'var(--text-primary)'}}>
                                            {attack.count}
                                        </div>
                                    </div>

                                    <div className="text-xs mb-0.5" style={{color:'var(--text-muted)'}}>
                                        {formatLocation(attack.srcIpInfo)}
                                    </div>

                                    <div className="flex items-center gap-2 text-xs" style={{color:'var(--text-secondary)'}}>
                                        <div className="flex items-center gap-1">
                                            <Clock className="w-2 h-2" />
                                            {formatTime(attack.lastAttackTime)}
                                        </div>
                                        <div className="truncate">
                                            {attack.domain}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    )
} 