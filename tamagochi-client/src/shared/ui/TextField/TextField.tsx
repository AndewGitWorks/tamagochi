import { useId, type InputHTMLAttributes, type ReactNode } from 'react';
import styles from './TextField.module.css';

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  error?: string;
  endAdornment?: ReactNode;
};

export const TextField = ({ label, error, endAdornment, ...props }: TextFieldProps) => {
  const generatedId = useId();
  const id = props.id || generatedId;
  return (
    <div className={styles.textFieldWrapper}>
      <div className={styles.labelWrapper}>
        <label htmlFor={id}>{label}:</label>
      </div>
      <div className={`${styles.inputWrapper} ${error ? styles.error : ''}`}>
        <input type="text" className={styles.input} {...props} id={id} />
        {endAdornment && <div className={styles.endAdornment}>{endAdornment}</div>}
      </div>
      {error && <span className={styles.errorMessage}>{error}</span>}
    </div>
  );
};
