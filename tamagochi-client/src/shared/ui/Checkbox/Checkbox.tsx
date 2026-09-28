import type { InputHTMLAttributes } from 'react';
import styles from './Checkbox.module.css';

type CheckboxProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
};

export const Checkbox = ({ label, checked, ...props }: CheckboxProps) => {
  return (
    <div className={styles.checkboxWrapper}>
      
      <input {...props} className={styles.checkbox} type="checkbox" id="checkbox" />
      <label htmlFor="checkbox">{label}</label>
    </div>
  );
};
