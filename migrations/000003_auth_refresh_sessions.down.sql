DROP TABLE IF EXISTS tamagochi.refresh_sessions;

ALTER TABLE tamagochi.users
    ALTER COLUMN password_hashed TYPE VARCHAR(64);
