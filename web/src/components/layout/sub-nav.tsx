import { Link, useLocation } from 'react-router'
import { cn } from '@/lib/utils'
import { ReactNode } from 'react'

export interface NavItem {
    title: string
    path: string
    icon?: ReactNode
}

interface SubNavProps {
    items: NavItem[]
}

export function SubNav({ items }: SubNavProps) {
    const location = useLocation()

    return (
        <nav className="flex space-x-6 border-b border-border bg-background px-6">
            {items.map((item) => {
                const isActive = location.pathname.endsWith(item.path)
                
                return (
                    <Link
                        key={item.path}
                        to={item.path}
                        className={cn(
                            'relative flex items-center gap-2 py-4 text-sm font-medium transition-colors hover:text-foreground',
                            isActive
                                ? 'text-foreground after:absolute after:bottom-0 after:left-0 after:right-0 after:h-0.5 after:bg-primary'
                                : 'text-muted-foreground'
                        )}
                    >
                        {item.icon}
                        {item.title}
                    </Link>
                )
            })}
        </nav>
    )
}