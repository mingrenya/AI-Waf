import { Outlet } from "react-router"
import { SidebarNav } from "@/components/common/sidebar-nav"
import { RoutePath, ROUTES } from "@/routes/constants"
import { useBreadcrumbMap } from "@/routes/config"
import { Terminal, FileCode } from "lucide-react"

export function NucleiLayOut() {
    const breadcrumbMap = useBreadcrumbMap()
    const items = breadcrumbMap[ROUTES.NUCLEI as RoutePath]?.items || []

    // 注入图标到 sidebar items
    const itemsWithIcons = items.map((item, i) => ({
        ...item,
        icon: i === 0 ? Terminal : FileCode,
    }))

    return (
        <div className="flex gap-4 p-4 h-full">
            <SidebarNav items={itemsWithIcons} />
            <div className="flex-1 overflow-auto">
                <Outlet />
            </div>
        </div>
    )
}
