import React from 'react'
import { useTranslation } from 'react-i18next'
import { AttackEventAggregateResult } from '@/types/log'
import { MapPin } from 'lucide-react'

interface AttackIPListProps {
    attackIPs: AttackEventAggregateResult[]
    isLoading?: boolean
}

/**
 * 攻击IP列表组件
 * 显示24小时内的攻击IP信息
 */
export const AttackIPList: React.FC<AttackIPListProps> = ({
    attackIPs,
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

    return (
        <div className="h-full p-3">
            <div className="mb-3">
                <div className="text-xs font-semibold text-white flex items-center gap-2">
                    <MapPin className="w-3 h-3 text-[#a071da]" />
                    <span className="tracking-wide text-shadow-glow-white">{t('securityDashboard.attackIPList.title')}</span>
                </div>
            </div>
            <div className="space-y-2">
                {isLoading ? (
                    <div className="space-y-2">
                        {[...Array(3)].map((_, i) => (
                            <div key={i} className="animate-pulse">
                                <div className="h-2 bg-white/20 rounded w-3/4 mb-1"></div>
                                <div className="h-2 bg-white/10 rounded w-1/2"></div>
                            </div>
                        ))}
                    </div>
                ) : attackIPs.length === 0 ? (
                    <div className="text-left text-white/60 py-2 text-xs">
                        {t('securityDashboard.attackIPList.noData')}
                    </div>
                ) : (
                    <div className="space-y-2">
                        {attackIPs.map((attack, index) => (
                            <div key={`${attack.srcIp}-${index}`} className="border-b border-white/10 last:border-b-0 pb-2 last:pb-0">
                                <div className="flex items-center justify-between">
                                    <div className="flex-1 min-w-0">
                                        <div className="font-mono text-xs font-medium text-white truncate">
                                            {attack.srcIp}
                                        </div>
                                        <div className="text-xs text-white/70 mt-0.5">
                                            {formatLocation(attack.srcIpInfo)}
                                        </div>
                                    </div>
                                    <div className="ml-2 text-right">
                                        <div className="text-xs font-bold text-white">
                                            {attack.count}
                                        </div>
                                        <div className="text-xs text-white/60">
                                            {t('securityDashboard.attackIPList.attackCount')}
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