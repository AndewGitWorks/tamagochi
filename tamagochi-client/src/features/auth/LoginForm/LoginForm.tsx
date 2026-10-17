import { Link } from 'react-router';
import { Button, Checkbox, PasswordField, TextField } from '@/shared/ui';

import styles from './LoginForm.module.css';

export const LoginForm = () => {
  return (
    <form className={styles.form}>
      <TextField placeholder="Введи логин, чудик" label="Логин" />
      <PasswordField label="Пароль" placeholder="Введи пароль, чудик" />
      <div className={styles.memory}>
        <Checkbox label="Запомнить меня" />
        <Link to="/forget-password" className={styles.link}>
          Забыл пароль
        </Link>
      </div>
      <Button title="Войти" type="submit" />
    </form>
  );
};
