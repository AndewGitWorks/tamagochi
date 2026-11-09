import { Button, Checkbox, PasswordField, TextField } from '@/shared/ui';
import styles from './RegistrationForm.module.css';
import { useForm, type SubmitHandler } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { registrationSchema, type RegistrationFields } from '../utils/RegistrationValidationSchema';
import { registrationRequest } from '../services';
import axios from 'axios';
import { useNavigate } from 'react-router';
import { tokenStorage } from '@/shared/utils/tokenStorage';
import { useAppDispatch } from '@/store/hooks';
import { setAuthenticated } from '../store';

export const RegistrationForm = () => {
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegistrationFields>({
    resolver: zodResolver(registrationSchema),
  });

  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const onFormSubmit: SubmitHandler<RegistrationFields> = async (data) => {
    const apiData = {
      login: data.login,
      name: data.name,
      surname: data.surname,
      email: data.email,
      phone_number: data.phone_number,
      password: data.password,
    };

    try {
      const authData = await registrationRequest(apiData);
      tokenStorage.setTokens(
        {
          accessToken: authData.access_token,
          refreshToken: authData.refresh_token,
        },
        'local',
      );
      dispatch(setAuthenticated());
      navigate('/');
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        setError('root.server', {
          type: 'server',
          message: 'Логин или почта пользователя уже заняты',
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
        placeholder="Ваш логин"
        label="Логин"
        {...register('login')}
        error={errors.login?.message}
      />
      <div className={styles.fullName}>
        <TextField
          placeholder="Ваше имя"
          label="Ваше имя"
          {...register('name')}
          error={errors.name?.message}
        />
        <TextField
          placeholder="Ваша фамилия"
          label="Ваша фамилия"
          {...register('surname')}
          error={errors.surname?.message}
        />
      </div>
      <TextField
        type="email"
        placeholder="Ваша почта"
        label="Ваша почта"
        {...register('email')}
        error={errors.email?.message}
      />
      <TextField
        type="tel"
        placeholder="+79991234567"
        label="Номер телефона"
        {...register('phone_number')}
        error={errors.phone_number?.message}
      />
      <PasswordField label="Пароль" {...register('password')} error={errors.password?.message} />
      <PasswordField
        label="Повторите пароль"
        {...register('confirmPassword')}
        error={errors.confirmPassword?.message}
      />
      <div className={styles.policy}>
        <Checkbox {...register('policy')} error={errors.policy?.message}>
          Я согласен с
          <a
            href="https://tamagotchi-official.com/gb/"
            target="_blank"
            rel="noopener noreferrer"
            className={styles.policyLink}>
            политикой конфиденциальности
          </a>{' '}
          и даю согласие на обработку персональных данных
        </Checkbox>
      </div>
      <Button title="Зарегистрироваться" type="submit" disabled={isSubmitting} />
      {errors.root?.server?.message && (
        <span className={styles.errorMessage}>{errors.root?.server?.message}</span>
      )}
    </form>
  );
};
