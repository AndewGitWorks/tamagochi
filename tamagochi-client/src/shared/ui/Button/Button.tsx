import type { ButtonHTMLAttributes } from 'react';
import styles from './Button.module.css';

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  title: string;
};

export const Button = ({ title, ...props }: ButtonProps) => {
  return (
    <>
      <button className={styles.button} {...props}>
        <span className={styles.buttonTitle}>{title}</span>
      </button>
    </>
  );
};
