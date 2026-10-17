import { Link } from 'react-router';
import styles from './AuthSwitch.module.css';

type AuthSwitchProps = {
  text: string;
  linkText: string;
  to: string;
};

export const AuthSwitch = ({ text, linkText, to }: AuthSwitchProps) => {
  return (
    <p className={styles.switch}>
      {text}
      <Link to={to} className={styles.link}>
        {linkText}
      </Link>
    </p>
  );
};
