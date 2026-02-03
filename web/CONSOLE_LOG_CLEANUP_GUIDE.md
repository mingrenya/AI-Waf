# 批量替换 console.log 为 logger - 修复指南

## 已创建文件
- ✅ `web/src/utils/logger.ts` - Logger 工具类

## 需要修改的文件 (20处)

### 1. Globe3DMap.tsx (16处)
文件路径: `web/src/feature/security-dashboard/component/globe3D-map/Globe3DMap.tsx`

**修改步骤**:
1. 在文件顶部添加导入:
```typescript
import logger from '@/utils/logger'
```

2. 替换所有 console.log/info/warn:
- `console.log(...)` → `logger.debug(...)`
- `console.info(...)` → `logger.info(...)`
- `console.warn(...)` → `logger.warn(...)`
- `console.error(...)` 保持不变（错误始终记录）

### 2. AttackEventFilter.tsx (1处)
文件路径: `web/src/feature/log/components/AttackEventFilter.tsx`

**第91行**:
```typescript
// 修改前
console.log('Form Values Changed:', value)

// 修改后
logger.debug('Form Values Changed:', value)
```

### 3. AdaptiveConfigForm.tsx (1处)
文件路径: `web/src/feature/adaptive-throttling/components/AdaptiveConfigForm.tsx`

**第99行**:
```typescript
// 修改前
console.log('No existing config, using defaults')

// 修改后
logger.info('No existing config, using defaults')
```

### 4. SecurityDashboardLayout.tsx (2处)
文件路径: `web/src/feature/security-dashboard/component/SecurityDashboardLayout.tsx`

**第104行**:
```typescript
// 修改前
console.log('统计数据发生变化，更新状态...')

// 修改后
logger.debug('统计数据发生变化，更新状态...')
```

**第124行**:
```typescript
// 修改前
console.log('攻击事件数据发生变化，更新3D地球和攻击IP列表,实时攻击列表...')

// 修改后
logger.debug('攻击事件数据发生变化，更新3D地球和攻击IP列表,实时攻击列表...')
```

## 自动化替换命令

如果想批量替换，可以使用以下命令:

```bash
# 进入web目录
cd web

# 查找所有需要替换的文件
grep -rl "console.log\|console.info\|console.warn" src/ --include="*.tsx" --include="*.ts" | grep -v node_modules

# 对每个文件执行替换（需要手动确认）
# 注意：console.error 不替换，因为错误应该始终记录
```

## 验证步骤

1. 确保所有文件顶部都导入了 logger:
```typescript
import logger from '@/utils/logger'
```

2. 在生产环境构建时验证没有输出:
```bash
npm run build
# 检查 dist/ 目录中的代码，确认日志已被条件编译
```

3. 在开发环境中验证日志正常输出:
```bash
npm run dev
# 打开浏览器控制台，确认日志仍然可见
```

## Logger 使用规范

```typescript
// ❌ 错误用法
console.log('Debug info')
console.info('Info message')

// ✅ 正确用法
logger.debug('Debug info')       // 调试信息 - 仅开发环境
logger.info('Info message')      // 一般信息 - 仅开发环境
logger.warn('Warning')           // 警告 - 仅开发环境
logger.error('Error occurred')   // 错误 - 开发和生产都记录
```

## 环境配置

Logger 根据 `import.meta.env.MODE` 自动判断环境:
- `development` - 所有日志输出
- `production` - 仅 error 输出

无需额外配置！
