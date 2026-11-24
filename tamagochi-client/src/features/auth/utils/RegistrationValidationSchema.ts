import z from 'zod';

const loginRegex = /^[A-Za-z0-9_.-]+$/;
const phoneRegex = /^\+[0-9]{9,14}$/;

export const registrationSchema = z
  .object({
    login: z
      .string({ error: 'Введите логин' })
      .trim()
      .min(1, 'Введите логин')
      .min(3, 'Логин должен содержать минимум 3 символа')
      .max(24, 'Логин должен содержать максимум 24 символа')
      .regex(loginRegex, 'Используйте только латинские буквы, цифры и символы _ . -'),
    name: z
      .string({ error: 'Введите имя' })
      .trim()
      .min(1, 'Введите имя')
      .min(3, 'Имя должно содержать минимум 3 символа')
      .max(100, 'Имя должно содержать максимум 100 символов'),
    surname: z
      .string({ error: 'Введите фамилию' })
      .trim()
      .min(1, 'Введите фамилию')
      .min(3, 'Фамилия должна содержать минимум 3 символа')
      .max(100, 'Фамилия должна содержать максимум 100 символов'),
    email: z
      .string({ error: 'Введите адрес электронной почты' })
      .trim()
      .min(1, 'Введите адрес электронной почты')
      .min(3, 'Адрес электронной почты должен содержать минимум 3 символа')
      .max(100, 'Адрес электронной почты должен содержать максимум 100 символов')
      .email('Введите корректный адрес электронной почты'),
    phone_number: z
      .string({ error: 'Введите номер телефона' })
      .trim()
      .min(1, 'Введите номер телефона')
      .regex(phoneRegex, 'Используйте формат +79991234567'),
    password: z
      .string({ error: 'Введите пароль' })
      .min(1, 'Введите пароль')
      .min(8, 'Пароль должен содержать минимум 8 символов'),
    confirmPassword: z.string({ error: 'Повторите пароль' }).min(1, 'Повторите пароль'),
    policy: z.boolean({ error: 'Подтвердите согласие с политикой' }).refine((value) => value, {
      error: 'Подтвердите согласие с политикой',
    }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    error: 'Пароли не совпадают. Проверьте введённые значения',
    path: ['confirmPassword'],
  });

export type RegistrationFields = z.infer<typeof registrationSchema>;
