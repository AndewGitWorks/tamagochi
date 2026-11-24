import z from 'zod';

export const loginSchema = z.object({
  login: z.string({ error: 'Это обязательное поле' }).trim().min(1),
  password: z.string({ error: 'Это обязательное поле' }).min(1),
  rememberMe: z.boolean().optional(),
});

export type LoginFields = z.infer<typeof loginSchema>;
