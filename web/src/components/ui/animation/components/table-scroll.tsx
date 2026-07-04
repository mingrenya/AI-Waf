import { ReactNode, useRef, useEffect, useState, useCallback } from "react"

interface TableScrollContainerProps {
    children: ReactNode
    className?: string
    showShadows?: boolean
    shadowOpacity?: number
}

/**
 * 轻量版滚动阴影容器 — 去除 motion/react 依赖
 * 使用原生滚动事件替代 useScroll + useTransform motion hooks
 */
export function TableScrollContainer({
    children,
    className = "",
    showShadows = true,
    shadowOpacity = 0.15
}: TableScrollContainerProps) {
    const containerRef = useRef<HTMLDivElement>(null)
    const [canScroll, setCanScroll] = useState(false)
    const [topOpacity, setTopOpacity] = useState(0)
    const [bottomOpacity, setBottomOpacity] = useState(0)

    const checkScrollability = useCallback(() => {
        if (containerRef.current) {
            const { scrollHeight, clientHeight } = containerRef.current
            setCanScroll(scrollHeight > clientHeight)
        }
    }, [])

    const handleScroll = useCallback(() => {
        if (!containerRef.current) return
        const { scrollTop, scrollHeight, clientHeight } = containerRef.current
        const maxScroll = scrollHeight - clientHeight
        if (maxScroll <= 0) {
            setTopOpacity(0)
            setBottomOpacity(0)
            return
        }
        const progress = scrollTop / maxScroll
        setTopOpacity(progress < 0.1 ? shadowOpacity : 0)
        setBottomOpacity(progress > 0.9 ? shadowOpacity : 0)
    }, [shadowOpacity])

    useEffect(() => {
        checkScrollability()
        window.addEventListener('resize', checkScrollability)
        return () => window.removeEventListener('resize', checkScrollability)
    }, [checkScrollability])

    useEffect(() => {
        const el = containerRef.current
        if (!el) return
        el.addEventListener('scroll', handleScroll, { passive: true })
        return () => el.removeEventListener('scroll', handleScroll)
    }, [handleScroll])

    return (
        <div className={`relative w-full h-full ${className}`}>
            {/* 顶部滚动阴影 */}
            {showShadows && canScroll && (
                <div
                    className="absolute top-0 left-0 right-0 h-4 pointer-events-none z-10 transition-opacity duration-300"
                    style={{
                        opacity: topOpacity,
                        background: `linear-gradient(to bottom, rgba(0,0,0,${shadowOpacity}), transparent)`
                    }}
                />
            )}

            {/* 滚动容器 */}
            <div
                ref={containerRef}
                className="overflow-auto h-full w-full scroll-smooth"
                style={{
                    WebkitOverflowScrolling: 'touch'
                }}
            >
                {children}
            </div>

            {/* 底部滚动阴影 */}
            {showShadows && canScroll && (
                <div
                    className="absolute bottom-0 left-0 right-0 h-4 pointer-events-none z-10 transition-opacity duration-300"
                    style={{
                        opacity: bottomOpacity,
                        background: `linear-gradient(to top, rgba(0,0,0,${shadowOpacity}), transparent)`
                    }}
                />
            )}
        </div>
    )
}