// src/routes/AppRouter.tsx
import { createBrowserRouter, RouteObject, RouterProvider } from "react-router"
import { useRoutes } from "./config"
import { RouteErrorBoundary } from "@/handler/error-boundary"

export function AppRouter() {
    // 使用现有的路由配置
    const routes = useRoutes()

    // 遍历路由并添加errorElement
    const routesWithErrorHandling = addErrorElementToRoutes(routes)

    // 创建路由器
    const router = createBrowserRouter(routesWithErrorHandling)

    return <RouterProvider router={router} />
}

// 递归添加错误处理
function addErrorElementToRoutes(routes: RouteObject[]) {
    return routes.map(route => {
        // 为每个路由添加错误元素
        const updatedRoute = {
            ...route,
            errorElement: <RouteErrorBoundary />
        }

        // 递归处理子路由
        if (route.children) {
            updatedRoute.children = addErrorElementToRoutes(route.children)
        }

        return updatedRoute
    })
}