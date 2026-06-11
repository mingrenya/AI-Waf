import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Card } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { EChartWrapper } from '@/feature/stats/components/charts/EChartWrapper'
import { get } from '@/api'
import type { EChartsOption } from 'echarts'

interface StatsData {
  logVolume: { timestamp: number; count: number }[]
  byLevel: Record<string, number>
  byComponent: Record<string, number>
  totalHits: number
}

export function LokiStatsPanel() {
  const { t } = useTranslation()
  const [duration, setDuration] = useState('1h')

  const { data, isPending } = useQuery({
    queryKey: ['loki-stats', duration],
    queryFn: () => get<StatsData>(`/log/loki-stats?duration=${duration}`),
    refetchInterval: 30000,
  })

  if (isPending || !data) {
    return (
      <Card className="p-6">
        <div className="h-64 flex items-center justify-center text-muted-foreground">
          Loading stats...
        </div>
      </Card>
    )
  }

  const logVolume = data.logVolume || []
  const byLevel = data.byLevel || {}
  const byComponent = data.byComponent || {}

  // Time series
  const lineOption: EChartsOption = {
    tooltip: { trigger: 'axis' },
    grid: { left: 50, right: 20, top: 30, bottom: 30 },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{
      name: 'Logs/min',
      type: 'line',
      data: logVolume.map((p) => [p.timestamp * 1000, p.count]),
      smooth: true,
      showSymbol: false,
      areaStyle: { opacity: 0.15 },
      lineStyle: { width: 2 },
    }],
  }

  // Level pie
  const levelNames: Record<string, string> = {
    error: t('lokiSearch.error'),
    warn: t('lokiSearch.warn'),
    info: t('lokiSearch.info'),
    debug: t('lokiSearch.debug'),
  }
  const levelColors: Record<string, string> = { error: '#ef4444', warn: '#f59e0b', info: '#3b82f6', debug: '#6b7280' }
  const pieOption: EChartsOption = {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { fontSize: 10 } },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '45%'],
      data: Object.entries(byLevel).map(([k, v]) => ({
        name: levelNames[k] || k,
        value: v,
        itemStyle: { color: levelColors[k] },
      })),
      label: { show: false },
    }],
  }

  // Component bar
  const barOption: EChartsOption = {
    tooltip: { trigger: 'axis' },
    grid: { left: 120, right: 20, top: 10, bottom: 20 },
    xAxis: { type: 'value' },
    yAxis: {
      type: 'category',
      data: Object.keys(byComponent).map((k) => k.replace('traffic-analyzer', 'Traffic').replace('flow-controller', 'FlowCtrl').replace('pattern-detector', 'Pattern').replace('baseline', 'Baseline')),
      axisLabel: { fontSize: 11 },
    },
    series: [{
      name: 'Count',
      type: 'bar',
      data: Object.values(byComponent),
      itemStyle: { borderRadius: [0, 4, 4, 0], color: '#6366f1' },
    }],
  }

  return (
    <Card className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium text-sm text-muted-foreground">{t('lokiSearch.timeRange')}</h3>
        <Select value={duration} onValueChange={setDuration}>
          <SelectTrigger className="w-[110px] h-7 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1h">{t('lokiSearch.last1h')}</SelectItem>
            <SelectItem value="6h">{t('lokiSearch.last6h')}</SelectItem>
            <SelectItem value="24h">{t('lokiSearch.last24h')}</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="text-center text-2xl font-bold">
        {data.totalHits.toLocaleString()}
        <span className="text-sm text-muted-foreground ml-2">{t('lokiSearch.totalHits')}</span>
      </div>

      {logVolume.length > 0 && (
        <EChartWrapper options={lineOption} height={200} />
      )}

      {Object.keys(byLevel).length > 0 && (
        <div className="grid grid-cols-2 gap-4">
          <EChartWrapper options={pieOption} height={220} />
          <EChartWrapper options={barOption} height={220} />
        </div>
      )}

      {logVolume.length === 0 && Object.keys(byLevel).length === 0 && (
        <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">
          No stats available yet. Stats accumulate as traffic flows through the WAF.
        </div>
      )}
    </Card>
  )
}
