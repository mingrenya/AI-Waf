import { z } from 'zod';

// Login form validation schema
export const loginSchema = z.object({
  username: z
    .string()
    .min(3, { message: '用户名至少需要3个字符' })
    .max(20, { message: '用户名最多20个字符' }),
  password: z
    .string()
    .min(6, { message: '密码至少需要6个字符' }),
});

// Password reset validation schema
export const passwordResetSchema = z
  .object({
    oldPassword: z.string().min(1, { message: '请输入当前密码' }),
    newPassword: z.string().min(6, { message: '新密码至少需要6个字符' }),
    confirmPassword: z.string().min(6, { message: '请确认新密码' }),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword'],
  });

// Types based on the schemas
export type LoginFormValues = z.infer<typeof loginSchema>;
export type PasswordResetFormValues = z.infer<typeof passwordResetSchema>; 