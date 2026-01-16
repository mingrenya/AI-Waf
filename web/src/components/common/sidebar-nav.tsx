import { cn } from "@/lib/utils"
import { Link, useLocation } from "react-router"

interface SidebarNavItem {
    title: string
    path: string
}

interface SidebarNavProps {
    items: SidebarNavItem[]
}

export function SidebarNav({ items }: SidebarNavProps) {
    const location = useLocation()
    
    return (
        <nav className="flex space-x-2 lg:flex-col lg:space-x-0 lg:space-y-1">
            {items.map((item) => {
                const isActive = location.pathname === item.path || location.pathname.startsWith(item.path + "/")
                return (
                    <Link
                        key={item.path}
                        to={item.path}
                        className={cn(
                            "inline-flex items-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
                            "justify-start px-4 py-2",
                            isActive
                                ? "bg-secondary text-secondary-foreground shadow-sm"
                                : "hover:bg-secondary/50 text-muted-foreground hover:text-foreground"
                        )}
                    >
                        {item.title}
                    </Link>
                )
            })}
        </nav>
    )
}
