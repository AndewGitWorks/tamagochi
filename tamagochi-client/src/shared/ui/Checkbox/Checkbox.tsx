import type { InputHTMLAttributes, ReactNode } from 'react';
import styles from './Checkbox.module.css';

type CheckboxProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  children?: ReactNode;
};

export const Checkbox = ({ label, checked, children, ...props }: CheckboxProps) => {
  return (
    <div className={styles.checkboxWrapper}>
      <input {...props} className={styles.checkbox} type="checkbox" id="checkbox" />
      <label htmlFor="checkbox">
        {label}
        {children}
      </label>
    </div>
  );
};
