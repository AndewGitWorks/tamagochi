import type { InputHTMLAttributes } from 'react';
import styles from './Checkbox.module.css';

type CheckboxProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  error?: string;
};

export const Checkbox = ({ label, children, error, ...props }: CheckboxProps) => {
  return (
    <div className={styles.checkboxContainer}>
      <div className={styles.checkboxWrapper}>
        <input {...props} className={styles.checkbox} type="checkbox" id="checkbox" />
        <label htmlFor="checkbox">
          {label}
          {label && children && ' '}
          {children}
        </label>
      </div>
      {error && <span className={styles.errorMessage}>{error}</span>}
    </div>
  );
};
