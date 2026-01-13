import { useOverviewStats } from '@/feature/stats/hooks/useStats'
import { useAttackEvents } from '@/feature/log/hook/useAttackEvents'
import { useMemo, useRef, useEffect, useState } from 'react'

const POLLING_INTERVAL = 5000 // 5秒轮询

/**
 * 安全大屏数据管理 Hook
 * 提供统计数据和实时攻击事件数据
 */
export const useSecurityDashboard = () => {
    const pollingTimerRef = useRef<number | null>(null)
    
    // 使用状态管理时间参数，确保每次轮询都能更新
    const [timeParams, setTimeParams] = useState(() => ({
        startTime: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
        endTime: new Date().toISOString()
    }))

    // 获取24小时统计概览数据
    const overviewStats = useOverviewStats('24h')

    // 使用状态中的时间参数构建查询参数
    const attackEventsParams = useMemo(() => ({
        pageSize: 50, // 获取更多数据用于展示
        startTime: timeParams.startTime,
        endTime: timeParams.endTime
    }), [timeParams.startTime, timeParams.endTime])

    // 获取24小时攻击事件数据
    const attackEvents = useAttackEvents(attackEventsParams)

    // 更新时间参数的函数
    const updateTimeParams = () => {
        setTimeParams({
            startTime: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
            endTime: new Date().toISOString()
        })
    }

    // 设置5秒轮询
    useEffect(() => {
        // 清除现有的轮询
        if (pollingTimerRef.current !== null) {
            clearInterval(pollingTimerRef.current)
            pollingTimerRef.current = null
        }

        // 设置新的轮询 - 5秒间隔
        pollingTimerRef.current = window.setInterval(() => {
            // 更新时间参数以获取最新的24小时数据
            // 时间参数更新会自动触发 useAttackEvents 重新请求，无需手动 refetch
            updateTimeParams()
            
            // 只需要手动刷新统计数据，因为它不依赖时间参数
            overviewStats.refetch()
        }, POLLING_INTERVAL)

        // 组件卸载时清除轮询
        return () => {
            if (pollingTimerRef.current !== null) {
                clearInterval(pollingTimerRef.current)
            }
        }
    }, [overviewStats.refetch])

    // 过滤有地理位置信息的攻击事件，用于实时攻击列表
    const realtimeAttacks = useMemo(() => {
        if (!attackEvents.data?.results) return []

        return attackEvents.data.results
            .filter(event => event.srcIpInfo?.city && event.srcIpInfo?.city?.nameZh && event.srcIpInfo?.city?.nameEn) // 只显示有城市信息的攻击
            .slice(0, 9) // 只显示最新的10条
    }, [attackEvents.data])

    // 攻击IP列表，用于左侧卡片4
    const attackIPs = useMemo(() => {
        if (!attackEvents.data?.results) return []

        return attackEvents.data.results
            .filter(event => event.srcIpInfo?.city) // 只显示有城市信息的攻击
            .slice(0, 5) // 只显示最新的5条
    }, [attackEvents.data])

    return {
        // 统计数据
        overviewStats: {
            data: overviewStats.data,
            isLoading: overviewStats.isLoading,
            error: overviewStats.error
        },
        // 攻击事件数据
        attackEvents: {
            data: attackEvents.data,
            isLoading: attackEvents.isLoading,
            error: attackEvents.error
        },
        // 处理后的数据
        realtimeAttacks,
        attackIPs,
        // 轮询间隔
        pollingInterval: POLLING_INTERVAL
    }
} 