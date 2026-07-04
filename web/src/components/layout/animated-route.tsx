import { ReactNode } from "react"
import { useLocation } from "react-router"

interface AnimatedRouteProps {
    children: ReactNode
    transitionType?: "slide" | "fade" | "scale" | "flip"
}

/**
 * 轻量版路由过渡组件
 *
 * 修复要点:
 * 1. 去掉 AnimatePresence mode="wait" — 不再阻塞页面切换
 * 2. 去掉 blur filter 动画 — 避免 GPU 重排
 * 3. 去掉 spring 弹性动画 — 减少帧计算
 * 4. 只用 CSS transition + opacity/transform — GPU 硬件加速、零 JS 开销
 *
 * 过渡时间 <200ms, 几乎无感但视觉平滑
 */
export function AnimatedRoute({
    children,
}: AnimatedRouteProps) {
    const location = useLocation()

    return (
        <div
            key={location.pathname}
            className="animate-page-enter"
        >
            {children}
        </div>
    )
} 