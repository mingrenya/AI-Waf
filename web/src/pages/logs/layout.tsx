import { Card } from "@/components/ui/card"
import { Outlet } from "react-router"
import { AnimatedRoute } from "@/components/layout/animated-route"

export function LogsLayout() {
    return (
        <Card className="flex-1 border-none shadow-none p-0 overflow-hidden rounded-none">
            <AnimatedRoute>
                <Outlet />
            </AnimatedRoute>
        </Card>
    )
}