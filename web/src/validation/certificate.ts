import * as z from 'zod';

export const certificateFormSchema = z.object({
  name: z.string().min(1, '证书名称不能为空'),
  description: z.string().optional(),
  publicKey: z.string().min(1, '公钥不能为空'),
  privateKey: z.string().min(1, '私钥不能为空'),
});

export const certificateUpdateSchema = z.object({
  name: z.string().optional(),
  description: z.string().optional(),
  publicKey: z.string().optional(),
  privateKey: z.string().optional(),
}); 