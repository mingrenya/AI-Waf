import React from "react"

interface TabsAnimationProviderProps {
    children: React.ReactNode
    currentView: string
    animationVariant?: "slide" | "fade" | "scale"
}

/**
 * 轻量版 Tab 过渡 — 去掉 AnimatePresence mode="wait"
 * 避免 tab 切换时阻塞渲染
 */
export function TabsAnimationProvider({
    children,
}: TabsAnimationProviderProps) {
    return (
        <div className="animate-page-enter">
            {children}
        </div>
    )
}