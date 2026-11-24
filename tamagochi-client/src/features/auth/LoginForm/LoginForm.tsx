import { Button, Checkbox, PasswordField, TextField } from '@/shared/ui';

import styles from './LoginForm.module.css';
import { useForm, type SubmitHandler } from 'react-hook-form';
import axios from 'axios';
import { loginRequest } from '../services';
import { zodResolver } from '@hookform/resolvers/zod';
import { loginSchema, type LoginFields } from '../utils/LoginValidationSchema';
import { tokenStorage } from '@/shared/utils/tokenStorage';
import { useNavigate } from 'react-router';
import { useAppDispatch } from '@/store/hooks';
import { setAuthenticated } from '../store';

export const LoginForm = () => {
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFields>({ resolver: zodResolver(loginSchema) });
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const onFormSubmit: SubmitHandler<LoginFields> = async (data) => {
    const { login, password, rememberMe } = data;
    try {
      const authData = await loginRequest({ login, password });
      tokenStorage.setTokens(
        {
          accessToken: authData.access_token,
          refreshToken: authData.refresh_token,
        },
        rememberMe ? 'local' : 'session',
      );
      dispatch(setAuthenticated());
      navigate('/');
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 401) {
        setError('root.server', {
          type: 'server',
          message: 'Неверный логин или пароль',
        });

        return;
      }

      setError('root.server', {
        type: 'server',
        message: 'Произошла неизвестная ошибка',
      });
    }
  };

  return (
    <form className={styles.form} onSubmit={handleSubmit(onFormSubmit)}>
      <TextField
        placeholder="Введите логин"
        label="Логин"
        {...register('login')}
        error={errors.login?.message}
      />
      <PasswordField
        label="Пароль"
        placeholder="Введите пароль"
        {...register('password')}
        error={errors.password?.message}
      />
      <div className={styles.memory}>
        <Checkbox label="Запомнить меня" {...register('rememberMe')} />
      </div>
      <Button title="Войти" type="submit" disabled={isSubmitting} />

      {errors.root?.server?.message && (
        <span className={styles.errorMessage}>{errors.root?.server?.message}</span>
      )}
    </form>
  );
};
