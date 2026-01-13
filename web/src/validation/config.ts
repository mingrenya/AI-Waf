// src/validation/config.ts
import { z } from 'zod'

export const configSchema = z.object({
    isDebug: z.boolean().default(false),
    isResponseCheck: z.boolean().default(false),
    haproxy: z.object({
        thread: z.number()
            .min(0, "线程数不能小于0")
            .max(256, "线程数不能超过256")
            .default(0),
        configBaseDir: z.string().min(1, "配置根目录不能为空"),
        haproxyBin: z.string().min(1, "可执行文件路径不能为空"),
    }),
    engine: z.object({
        appConfig: z.array(z.object({
            name: z.string(),
            directives: z.string()
        }))
    })
})

export type ConfigFormValues = z.infer<typeof configSchema>