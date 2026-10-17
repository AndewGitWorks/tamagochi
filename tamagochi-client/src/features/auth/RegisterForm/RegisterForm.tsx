import { Button, Checkbox, PasswordField, TextField } from '@/shared/ui';
import styles from './RegisterForm.module.css';
import { Link } from 'react-router';

export const RegisterForm = () => {
  return (
    <form className={styles.form}>
      <div className={styles.fullName}>
        <TextField placeholder="Твое имя, чудик" label="Ваше имя" />
        <TextField placeholder="Твоя фамилия, чудик" label="Ваша фамилия" />
      </div>
      <TextField placeholder="Твоя почта, чудик" label="Ваша почта" />
      <PasswordField label="Пароль" />
      <PasswordField label="Повторите пароль" />
      <div className={styles.policy}>
        <Checkbox>
          Я согласен с
          <Link to="/eggplant" className={styles.policyLink}>
            политикой конфиденциальности
          </Link>{' '}
          и даю согласие на обработку персональных данных
        </Checkbox>
      </div>
      <Button title="Зарегистрироваться" type="submit" />
    </form>
  );
};
