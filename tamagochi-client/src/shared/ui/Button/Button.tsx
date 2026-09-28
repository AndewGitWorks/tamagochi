import type { ButtonHTMLAttributes, ReactNode } from 'react';
import styles from './Button.module.css';

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  title: string;
  children?: ReactNode;
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
