import { AuthLayout } from '@/features/auth/ui/AuthLayout/AuthLayout';
import styles from './RegisterPage.module.css';
import { RegisterForm } from '@/features/auth/RegisterForm/RegisterForm';
import { AuthSwitch } from '@/features/auth/ui/AuthSwitch';
import { SocialAuth } from '@/features/auth/ui/SocialAuth/SocialAuth';

export const RegisterPage = () => {
  return (
    <AuthLayout>
      <header className={styles.headerRegister}>
        <h1 className={styles.title}>Регистрация</h1>
        <p className={styles.description}>Заполните данные, чтобы зарегистрировать свой аккаунт</p>
      </header>
      <RegisterForm />
      <AuthSwitch to="/login" text="У вас есть аккаунта?" linkText="Войти" />
      <SocialAuth />
    </AuthLayout>
  );
};
