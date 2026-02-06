import { useQuery } from '@tanstack/react-query'
import { alertHistoryApi } from '@/api/alert'
import { queryKeys } from '@/lib/query-keys'
import { StatsCard } from '@/feature/stats/components/StatsCard'
import { AlertCircle, CheckCircle2, Clock3, XCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export function AlertStatsCards() {
    const { t } = useTranslation()

    const { data, isLoading } = useQuery({
        queryKey: queryKeys.alert.history.stats(),
        queryFn: () => alertHistoryApi.getStats(),
    })

    const total = data?.totalAlerts ?? 0
    const pending = data?.alertsByStatus?.pending ?? 0
    const sent = data?.alertsByStatus?.sent ?? 0
    const failed = data?.alertsByStatus?.failed ?? 0
    const acknowledged = data?.alertsByStatus?.acknowledged ?? 0

    return (
        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4 mb-4">
            <StatsCard
                title={t('alert.stats.total', { defaultValue: 'Total alerts' })}
                value={total}
                icon={<AlertCircle className="h-4 w-4" />}
                loading={isLoading}
            />
            <StatsCard
                title={t('alert.stats.pending', { defaultValue: 'Pending' })}
                value={pending}
                icon={<Clock3 className="h-4 w-4" />}
                loading={isLoading}
            />
            <StatsCard
                title={t('alert.stats.sent', { defaultValue: 'Sent' })}
                value={sent}
                icon={<CheckCircle2 className="h-4 w-4" />}
                loading={isLoading}
            />
            <StatsCard
                title={t('alert.stats.failed', { defaultValue: 'Failed' })}
                value={failed}
                icon={<XCircle className="h-4 w-4" />}
                loading={isLoading}
            />
            <StatsCard
                title={t('alert.stats.acknowledged', {
                    defaultValue: 'Acknowledged',
                })}
                value={acknowledged}
                icon={<CheckCircle2 className="h-4 w-4 text-emerald-500" />}
                loading={isLoading}
            />
        </div>
    )
}

