import { useEffect, useRef } from 'react'
import { useRealtimeQPS } from '@/feature/stats/hooks/useStats'
import * as echarts from 'echarts'

/**
 * 大屏专用实时QPS图表组件
 * 适配底部布局，使用更多数据点
 * 优化版本：增强发光效果和视觉体验
 */
export function DashboardQPSChart() {
    const chartRef = useRef<HTMLDivElement>(null)
    const chartInstanceRef = useRef<echarts.ECharts | null>(null)

    // 使用120个数据点，适合大屏展示
    const { localData, isInitialized } = useRealtimeQPS(120)

    // 当前QPS值 - 取最新的数据点值
    const currentQPS = localData.length > 0 ? localData[localData.length - 1].value : 0

    // 格式化时间戳 - 显示时:分:秒
    const formatTimeLabel = (timestamp: string) => {
        const date = new Date(timestamp)
        return date.toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        })
    }

    // 初始化和更新图表
    useEffect(() => {
        if (!chartRef.current) return

        // 如果已经有图表实例，则不需要重新创建
        if (!chartInstanceRef.current) {
            chartInstanceRef.current = echarts.init(chartRef.current)
        }

        // 如果没有数据，显示空图表
        if (localData.length === 0) {
            chartInstanceRef.current.setOption({
                grid: {
                    left: '2%',
                    right: '2%',
                    bottom: '20%',
                    top: '5%',
                    containLabel: false
                },
                xAxis: {
                    type: 'category',
                    data: [],
                    axisLabel: {
                        color: 'rgba(255, 255, 255, 0.8)',
                        fontSize: 10
                    },
                    axisLine: {
                        lineStyle: {
                            color: 'rgba(255, 255, 255, 0.3)'
                        }
                    }
                },
                yAxis: {
                    type: 'value',
                    axisLabel: {
                        show: false
                    },
                    splitLine: {
                        show: false
                    }
                },
                series: [{
                    type: 'bar',
                    data: []
                }]
            })
            return
        }

        // 大屏专用颜色配置 - 增强版
        const CHART_PRIMARY_COLOR = '#a071da' // 紫色主题
        const CHART_SECONDARY_COLOR = 'rgba(160, 113, 218, 0.3)' // 半透明紫色
        const CHART_GLOW_COLOR = 'rgba(160, 113, 218, 0.8)' // 发光颜色
        const CHART_BRIGHT_COLOR = '#c299f0' // 更亮的紫色用于渐变顶部

        // 处理数据，确保0值也显示小柱子
        const processedData = localData.map(item => {
            // 如果值为0，给一个最小值0.1以便显示发光效果
            return item.value === 0 ? 0.1 : item.value
        })

        // 设置图表选项
        const option: echarts.EChartsOption = {
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'shadow',
                    shadowStyle: {
                        color: 'rgba(160, 113, 218, 0.15)'
                    }
                },
                formatter: function (params) {
                    if (!params || !Array.isArray(params) || params.length === 0) {
                        return ''
                    }

                    const dataIndex = params[0].dataIndex
                    if (typeof dataIndex === 'number' && dataIndex >= 0 && dataIndex < localData.length) {
                        const item = localData[dataIndex]
                        // 显示实际值，而不是处理后的值
                        return `<div style="text-shadow: 0 0 8px rgba(160, 113, 218, 0.8);">${formatTimeLabel(item.timestamp)}: <span style="font-weight: bold; color: ${CHART_BRIGHT_COLOR}; text-shadow: 0 0 10px rgba(160, 113, 218, 1);">${item.value}</span> QPS</div>`
                    }
                    return ''
                },
                backgroundColor: 'transparent',
                borderColor: 'transparent',
                textStyle: {
                    color: '#fff',
                    fontSize: 12,
                    fontWeight: 'normal'
                },
                extraCssText: 'box-shadow: none; backdrop-filter: none;'
            },
            grid: {
                left: '2%',
                right: '2%',
                bottom: '20%',
                top: '5%',
                containLabel: false
            },
            xAxis: {
                type: 'category',
                boundaryGap: true,
                data: localData.map(item => formatTimeLabel(item.timestamp)),
                axisLabel: {
                    show: true,
                    color: 'rgba(255, 255, 255, 0.8)',
                    fontSize: 10,
                    interval: Math.floor(localData.length / 8) // 显示8个左右的标签
                },
                axisLine: {
                    show: true,
                    lineStyle: {
                        color: 'rgba(255, 255, 255, 0.3)'
                    }
                },
                axisTick: {
                    show: false
                },
                splitLine: {
                    show: false
                }
            },
            yAxis: {
                type: 'value',
                axisLabel: {
                    show: false
                },
                axisLine: {
                    show: false
                },
                axisTick: {
                    show: false
                },
                splitLine: {
                    show: false
                },
                max: function (value) {
                    // 让图表顶部有一些留白，最大值上浮20%
                    return Math.max(10, value.max * 1.2)
                },
                min: 0
            },
            series: [
                {
                    name: 'QPS',
                    type: 'bar',
                    barWidth: '60%',
                    itemStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            { offset: 0, color: CHART_BRIGHT_COLOR },
                            { offset: 0.3, color: CHART_PRIMARY_COLOR },
                            { offset: 1, color: CHART_SECONDARY_COLOR }
                        ]),
                        borderRadius: [4, 4, 0, 0], // 增加圆角
                        // 增强发光效果 - 多层阴影
                        shadowColor: CHART_GLOW_COLOR,
                        shadowBlur: 20, // 增加模糊半径
                        shadowOffsetY: 0,
                        shadowOffsetX: 0
                    },
                    // 添加强调状态的发光效果
                    emphasis: {
                        itemStyle: {
                            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                                { offset: 0, color: '#d4b3ff' },
                                { offset: 0.3, color: '#b388e6' },
                                { offset: 1, color: 'rgba(160, 113, 218, 0.5)' }
                            ]),
                            shadowColor: CHART_GLOW_COLOR,
                            shadowBlur: 30, // 悬停时更强的发光
                            shadowOffsetY: 0,
                            shadowOffsetX: 0,
                            borderColor: CHART_BRIGHT_COLOR,
                            borderWidth: 1
                        }
                    },
                    data: processedData,
                    animationDuration: 300,
                    animationEasing: 'cubicOut'
                }
            ],
            animation: true
        }

        // 应用选项
        chartInstanceRef.current.setOption(option)

        // 调整窗口大小时重绘图表
        const handleResize = () => {
            if (chartInstanceRef.current) {
                chartInstanceRef.current.resize()
            }
        }

        window.addEventListener('resize', handleResize)

        // 清理函数
        return () => {
            window.removeEventListener('resize', handleResize)
        }
    }, [localData, isInitialized])

    // 组件卸载时清理图表实例
    useEffect(() => {
        return () => {
            if (chartInstanceRef.current) {
                chartInstanceRef.current.dispose()
                chartInstanceRef.current = null
            }
        }
    }, [])

    return (
        <div className="h-full pb-12">
            <div className="flex items-center mb-2 ml-8">
                <div className="flex items-center gap-2">
                    <span className="text-xl font-bold text-white text-shadow-glow-white">{currentQPS}</span>
                    <span className="text-white text-xs text-shadow-glow-white">QPS</span>
                </div>
            </div>
            <div
                ref={chartRef}
                className='w-full h-full'
            />
        </div>
    )
} 