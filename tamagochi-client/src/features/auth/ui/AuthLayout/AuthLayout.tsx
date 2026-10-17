import type { ReactNode } from 'react';
import styles from './AuthLayout.module.css';

type AuthLayoutProps = {
  children: ReactNode;
};

export const AuthLayout = ({ children }: AuthLayoutProps) => {
  return (
    <div className={styles.page}>
      <div className={styles.content}>{children}</div>
    </div>
  );
};
