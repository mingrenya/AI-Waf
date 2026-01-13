import React, { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Globe3DMap } from './globe3D-map'
import { WAFAttackTrajectory } from './globe3D-map/types'
import { StatCard } from './StatCard'
import { AttackIPList } from './AttackIPList'
import { RealtimeAttackList } from './RealtimeAttackList'
import { DashboardQPSChart } from './DashboardQPSChart'
import { useSecurityDashboard } from '../hooks/useSecurityDashboard'
import { Maximize2, Minimize2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { OverviewStats } from '@/types/stats'
import { AttackEventAggregateResult } from '@/types/log'
import { useDebounce } from '@/hooks/useDebounce'
import { geoDistance } from 'd3-geo';

/**
 * 安全大屏布局组件
 * 包含顶部标题栏、左侧统计卡片、右侧实时攻击列表、底部QPS图表和背景3D地球
 */
export const SecurityDashboardLayout: React.FC = () => {
    const { t, i18n } = useTranslation()
    const [currentTime, setCurrentTime] = useState(new Date())
    const [isFullscreen, setIsFullscreen] = useState(false)

    // 独立的状态管理 - 只有数据真正变化时才更新
    const [globeData, setGlobeData] = useState<WAFAttackTrajectory[]>([])
    const globeDataDebounce = useDebounce(globeData, 10000) // 10 秒防抖，防抖时间要大于轮训时间
    const [statsData, setStatsData] = useState<OverviewStats | null>(null)
    const [realtimeAttacksData, setRealtimeAttacksData] = useState<AttackEventAggregateResult[]>([])
    const [attackIPsData, setAttackIPsData] = useState<AttackEventAggregateResult[]>([])

    // 用于数据变化检测的ref - 避免无限渲染
    const prevDataRef = useRef<{
        overviewStats: OverviewStats | null
        attackEvents: typeof attackEvents.data | null
    }>({
        overviewStats: null,
        attackEvents: null
    })

    // 获取大屏数据
    const {
        overviewStats,
        attackEvents,
        realtimeAttacks,
        attackIPs
    } = useSecurityDashboard()

    // 时间更新器 - 每秒更新一次时间显示
    useEffect(() => {
        const timeUpdateRef = setInterval(() => {
            setCurrentTime(new Date())
        }, 1000)

        return () => {
            clearInterval(timeUpdateRef)
        }
    }, [])

    // 监听全屏状态变化
    useEffect(() => {
        const handleFullscreenChange = () => {
            setIsFullscreen(!!document.fullscreenElement)
        }

        document.addEventListener('fullscreenchange', handleFullscreenChange)
        return () => {
            document.removeEventListener('fullscreenchange', handleFullscreenChange)
        }
    }, [])

    // 格式化时间显示
    const formatCurrentTime = (date: Date) => {
        const locale = i18n.language === 'zh' ? 'zh-CN' : 'en-US'
        return date.toLocaleString(locale, {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        }).replace(/\//g, '-')
    }

    // 全屏切换
    const toggleFullscreen = () => {
        if (!document.fullscreenElement) {
            document.documentElement.requestFullscreen()
        } else {
            document.exitFullscreen()
        }
    }

    // 监听统计数据变化，只有真正变化时才更新状态
    useEffect(() => {
        const currentStats = overviewStats.data
        const prevStats = prevDataRef.current.overviewStats

        // 检查统计数据是否发生变化
        const statsChanged = JSON.stringify(currentStats) !== JSON.stringify(prevStats)

        if (statsChanged) {
            console.log('统计数据发生变化，更新状态...')
            prevDataRef.current.overviewStats = currentStats || null
            setStatsData(currentStats || null)
        }
    }, [overviewStats.data])


    // 监听攻击事件数据变化，更新3D地球数据和攻击IP列表
    useEffect(() => {
        const currentAttackEvents = attackEvents.data
        const prevAttackEvents = prevDataRef.current.attackEvents

        // 检查攻击事件数据是否发生变化（数据已在 hook 中标准化）
        const attackEventsChanged = JSON.stringify(currentAttackEvents) !== JSON.stringify(prevAttackEvents)


        if (!attackEventsChanged) {
            return // 数据没有变化，不更新
        }

        console.log('攻击事件数据发生变化，更新3D地球和攻击IP列表,实时攻击列表...')
        prevDataRef.current.attackEvents = currentAttackEvents

        if (!currentAttackEvents?.results || currentAttackEvents?.results?.length === 0) {
            setGlobeData([])
            setAttackIPsData([])
            setRealtimeAttacksData([])
            return
        }

        // 更新3D地球数据
        const newTrajectoryData = currentAttackEvents.results
            .filter(event => event.srcIpInfo?.location?.latitude) // 只显示有位置信息的攻击
            .map((event, index) => ({
                type: "waf_attack",
                order: index + 1,
                from: `${event.srcIp} (${event.srcIpInfo.location.latitude.toFixed(2)}, ${event.srcIpInfo.location.longitude.toFixed(2)})`,
                to: "WAF防护中心",
                flightCode: event.domain,
                date: event.firstAttackTime,
                status: event.isOngoing,
                startLat: event.srcIpInfo.location.latitude,
                startLng: event.srcIpInfo.location.longitude,
                endLat: 30.274084, // 杭州坐标
                endLng: 120.155070,
                // arcAlt: Math.min(0.3, Math.max(0.05, event.count / 500)),
                arcAlt: geoDistance([event.srcIpInfo.location.longitude, event.srcIpInfo.location.latitude], [120.155070, 30.274084]), // 计算起点和终点之间的地理距离(线条需要根据距离生成相应的贝塞尔曲线高度)
                colorIndex: Math.abs(event.srcIp.split('.').reduce((a, b) => a + parseInt(b), 0)) % 8 // 基于IP地址生成固定颜色索引
            }))

        setGlobeData(newTrajectoryData)

        // 更新攻击IP列表数据
        setAttackIPsData(attackIPs)
        setRealtimeAttacksData(realtimeAttacks)
    }, [attackEvents.data])

    return (
        <div
            className={`relative w-full bg-gradient-to-br from-[#0d0c27] via-[#1a1336] to-[#2d1b54] overflow-hidden ${isFullscreen
                ? 'fixed inset-0 z-[9999] h-screen w-screen'
                : 'h-screen'
                }`}
            style={isFullscreen ? {
                margin: 0,
                padding: 0,
                width: '100vw',
                height: '100vh'
            } : undefined}
        >
            {/* 3D地球背景 - 使用独立状态数据 */}
            <div className="absolute inset-0 z-0 w-full h-full">
                <Globe3DMap wafAttackTrajectoryData={globeDataDebounce} />
            </div>

            {/* 顶部标题栏 */}
            <div className="absolute top-0 left-0 right-0 z-10 h-24">
                <div className="flex items-center justify-between h-full px-6">
                    <div className="flex items-center ml-2">
                        <div className="font-bold text-2xl gap-2 flex">
                            <span className="text-[#E8DFFF] dark:text-[#F0EBFF] text-shadow-glow-purple transition-all duration-300">RuiQi WAF</span>
                            <span className="text-[#8ED4FF] dark:text-[#A5DEFF] text-shadow-glow-blue transition-all duration-300">{t('securityDashboard.title')}</span>
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className="text-white text-lg font-mono text-shadow-glow-white">
                            {formatCurrentTime(currentTime)}
                        </div>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={toggleFullscreen}
                            className="text-white transition-all duration-200"
                            title={isFullscreen ? "退出全屏模式" : "进入全屏模式"}
                        >
                            {isFullscreen ? (
                                <Minimize2 className="w-4 h-4" />
                            ) : (
                                <Maximize2 className="w-4 h-4" />
                            )}
                        </Button>
                    </div>
                </div>
            </div>

            {/* 左侧统计卡片区域 - 使用独立状态数据 */}
            <div className="absolute left-6 top-28 z-10 w-64 flex flex-col">
                {/* 前三个统计卡片 */}
                <div className="flex-none space-y-0.5">
                    <StatCard
                        title={t('securityDashboard.stats.blockCount24h')}
                        value={statsData?.blockCount || 0}
                    />
                    <StatCard
                        title={t('securityDashboard.stats.attackIPCount24h')}
                        value={statsData?.attackIPCount || 0}
                    />
                    <StatCard
                        title={t('securityDashboard.stats.totalRequests24h')}
                        value={statsData?.totalRequests || 0}
                    />
                </div>

                {/* 第四个卡片 - 攻击IP列表，使用独立状态数据 */}
                <div className="flex-1 min-h-0 mt-20">
                    <AttackIPList
                        attackIPs={attackIPsData}
                        isLoading={attackEvents.isLoading}
                    />
                </div>
            </div>

            {/* 右侧实时攻击列表 - 使用独立状态数据 */}
            <div className="absolute right-6 top-28 bottom-48 z-10 w-80">
                <RealtimeAttackList
                    realtimeAttacks={realtimeAttacksData}
                    isLoading={attackEvents.isLoading}
                />
            </div>

            {/* 底部QPS图表 */}
            <div className="absolute left-2 right-2 bottom-2 z-10 h-32">
                <DashboardQPSChart />
            </div>
        </div>
    )
} 