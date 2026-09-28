import type { InputHTMLAttributes, ReactNode } from 'react';
import styles from './TextField.module.css';

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  error?: string;
  endAdornment?: ReactNode;
};

export const TextField = ({ label, error, endAdornment, ...props }: TextFieldProps) => {
  return (
    <div className={styles.textFieldWrapper}>
      <div className={styles.labelWrapper}>
        <label htmlFor={props.id}>{label}:</label>
      </div>
      <div className={`${styles.inputWrapper} ${!!error ? styles.error : ''}`}>
        <input className={styles.input} {...props} />
        {endAdornment && <div className={styles.endAdornment}>{endAdornment}</div>}
      </div>
      {error && <span className={styles.errorMessage}>{error}</span>}
    </div>
  );
};
