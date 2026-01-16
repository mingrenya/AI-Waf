import { Outlet } from "react-router"
import { SidebarNav } from "@/components/common/sidebar-nav"
import { RoutePath, ROUTES } from "@/routes/constants"
import { useBreadcrumbMap } from "@/routes/config"

export function AIAnalyzerLayOut() {
    const breadcrumbMap = useBreadcrumbMap()
    const items = breadcrumbMap[ROUTES.AI_ANALYZER as RoutePath]?.items || []

    return (
        <div className="flex gap-4 p-4 h-full">
            <SidebarNav items={items} />
            <div className="flex-1 overflow-auto">
                <Outlet />
            </div>
        </div>
    )
}
