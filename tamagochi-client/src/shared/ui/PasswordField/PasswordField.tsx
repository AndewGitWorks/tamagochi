import { EyeClosedIcon, EyeOpenIcon } from '@/assets/icons';
import { TextField } from '@/shared/ui/TextField';
import { useState, type InputHTMLAttributes } from 'react';

import styles from './PasswordField.module.css';

type PasswordFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  error?: string;
};

export const PasswordField = ({ label, error, ...props }: PasswordFieldProps) => {
  const [isPasswordVisible, setIsPasswordVisible] = useState(false);

  const handleVisibiliity = () => {
    setIsPasswordVisible((prev) => !prev);
  };

  return (
    <TextField
      {...props}
      label={label}
      error={error}
      type={isPasswordVisible ? 'text' : 'password'}
      endAdornment={
        <button
          type="button"
          className={styles.eyeBtn}
          aria-label={isPasswordVisible ? 'Скрыть пароль' : 'Показать пароль'}
          onClick={handleVisibiliity}>
          {isPasswordVisible ? <EyeClosedIcon /> : <EyeOpenIcon />}
        </button>
      }
    />
  );
};
