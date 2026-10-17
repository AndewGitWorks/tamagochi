import { AuthLayout } from '@/features/auth/ui/AuthLayout/AuthLayout';
import styles from './LoginPage.module.css';
import { LoginForm } from '@/features/auth/LoginForm/LoginForm';
import { AuthSwitch } from '@/features/auth/ui/AuthSwitch';
import { SocialAuth } from '@/features/auth/ui/SocialAuth/SocialAuth';

export const LoginPage = () => {
  return (
    <AuthLayout>
      <header className={styles.headerLogin}>
        <h1 className={styles.title}>Авторизация</h1>
        <p className={styles.description}>Заполните данные, чтобы войти в аккаунт</p>
      </header>
      <LoginForm />
      <AuthSwitch to="/register" text="У вас нет акккаунта?" linkText="Зарегистрироваться" />
      <SocialAuth />
    </AuthLayout>
  );
};
