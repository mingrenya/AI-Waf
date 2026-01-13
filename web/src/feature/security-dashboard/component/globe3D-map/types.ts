// WAF攻击轨迹数据类型 - 用于安全仪表板可视化
export type WAFAttackTrajectory = {
    type: string
    order: number
    from: string
    to: string
    flightCode: string
    date: string
    status: boolean
    startLat: number
    startLng: number
    endLat: number
    endLng: number
    arcAlt: number
    colorIndex: number // 攻击类型颜色索引
}

// WAF攻击轨迹颜色配置 - 不同颜色代表不同类型的攻击
export const WAF_ATTACK_TRAJECTORY_COLORS = [
    "#8ed4ff", // 浅蓝色 - 一般攻击
    "#a071da", // 紫色 - 主题色攻击
    "#ff6b6b", // 红色 - 高危攻击
    "#4ecdc4", // 青色 - 中等攻击
    "#45b7d1", // 蓝色 - 探测攻击
    "#96ceb4", // 绿色 - 低级攻击
    "#ffeaa7", // 黄色 - 扫描类攻击
    "#fd79a8"  // 粉色 - 特殊攻击
]

// WAF防护中心坐标 (杭州)
export const WAF_DEFENSE_CENTER_COORDS = {
    lat: 30.274084,
    lng: 120.155070
} 