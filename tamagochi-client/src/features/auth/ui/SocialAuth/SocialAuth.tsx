import { VKLogoIcon, YandexLogoIcon } from '@/assets/icons';
import styles from './SocialAuth.module.css';

export const SocialAuth = () => {
  return (
    <section className={styles.socialAuth}>
      <h2 className={styles.title}>Или войти с помощь</h2>
      <div className={styles.buttons}>
        <button className={styles.btnProvider} aria-label="Войти с помощью VK">
          <VKLogoIcon />
        </button>
        <button className={styles.btnProvider} aria-label="Войти с помощью Шмяндекс">
          <YandexLogoIcon />
        </button>
      </div>
    </section>
  );
};
