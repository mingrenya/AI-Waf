import * as z from 'zod'
import { WAFMode } from '@/types/site'

// 服务器验证模式
const serverSchema = z.object({
    host: z.string().min(1, "主机地址不能为空"),
    port: z.number().min(1, "端口必须大于0").max(65535, "端口不能超过65535"),
    isSSL: z.boolean().default(false)
})

// 后端配置验证模式
const backendSchema = z.object({
    servers: z.array(serverSchema).min(1, "至少需要一个服务器")
})

// 证书验证模式（可选）
const certificateSchema = z.object({
    certName: z.string().optional(),
    expireDate: z.string().optional(),
    fingerPrint: z.string().optional(),
    issuerName: z.string().optional(),
    privateKey: z.string().optional(),
    publicKey: z.string().optional()
}).optional()

// 站点表单验证模式
export const siteFormSchema = z.object({
    name: z.string().min(1, "站点名称不能为空"),
    domain: z.string().min(1, "域名不能为空"),
    listenPort: z.number().min(1, "监听端口必须大于0").max(65535, "监听端口不能超过65535"),
    enableHTTPS: z.boolean().default(false),
    activeStatus: z.boolean().default(true),
    wafEnabled: z.boolean().default(false),
    wafMode: z.enum([WAFMode.Protection, WAFMode.Observation]).default(WAFMode.Observation),
    backend: backendSchema,
    certificate: certificateSchema,
    certificateId: z.string().optional() // 用于选择已有证书
})

// 站点更新表单验证模式
export const siteUpdateFormSchema = siteFormSchema.partial() 