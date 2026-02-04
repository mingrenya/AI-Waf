# React 状态管理审计报告

**审计日期**: 2026年2月4日
**项目路径**: `web/src`

## 1. 总体架构评估

项目采用了 **Zustand + React Query** 的混合架构，这完全符合现代 React 应用的最佳实践（Pattern 5: Combining Client + Server State）。

- **客户端状态 (Client State)**: 使用 `Zustand` 管理 UI 交互状态（Auth, Theme, Toast）。
- **服务端状态 (Server State)**: 使用 `TanStack Query (React Query)` 管理 API 数据同步。

## 2. 详细分析

### 2.1 Zustand (客户端状态)

- **Store 设计**: Store 被拆分为独立的模块（Auth, Theme, Toast），职责单一。
- **性能优化**: 
  - 使用了细粒度的 Selector Hooks（如 `useUser`, `useToken`），避免了整个 Store 更新导致的组件不必要重渲染。
  - 使用了 `persist` 中间件对关键状态（Auth, Theme）进行持久化。
- **可维护性**: 通过 `web/src/store/index.ts` 统一导出，引用方便。

### 2.2 React Query (服务端状态)

- **Query Key 管理**: 
  - 采用了 **Query Key Factory** 模式 (`web/src/lib/query-keys.ts`)，统一管理所有 Query Key。
  - 这是一个非常优秀的实践，避免了分散的硬编码字符串，便于维护和重构。
- **Hook 封装**: 
  - 所有 API 请求都封装在 Custom Hooks 中（如 `useAttackPatterns`, `useCreateSite`）。
  -实现了自动的 `invalidateQueries` 策略，确保数据的一致性。
- **配置**: 在 `main.tsx` 中配置了全局的 `staleTime` 和 `retry` 策略，并集成了 DevTools。

## 3. 发现的问题与改进建议

尽管整体架构非常出色，仍发现以下优化空间：

1. **Mutation 状态冗余**:
   - 在 `web/src/feature/site/hooks/useSites.ts` (以及类似文件) 中，手动维护了 `error` 状态：
     ```typescript
     const [error, setError] = useState<string | null>(null)
     ```
   - **建议**: 直接使用 useMutation 返回的 `error` 和 `isError` 属性，减少样板代码。

2. **组件内的状态管理**:
   - 在 `web/src/feature/security-dashboard/component/SecurityDashboardLayout.tsx` 中，发现了使用 `useState` 存储大量数据：
     ```typescript
     const [globeData, setGlobeData] = useState<WAFAttackTrajectory[]>([])
     ```
   - **建议**: 如果这些数据来自 API，应完全交由 components 内的 `useQuery` 管理，利用其缓存和防抖特性，避免手动 useEffect fetch 数据。

## 4. 结论

该项目的前端状态管理处于**优秀**水平。代码结构清晰，技术选型合理，且遵循了严格的类型安全和最佳实践。无需进行架构层面的重构，仅需关注上述细节优化。
