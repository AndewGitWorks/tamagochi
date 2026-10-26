ALTER TABLE tamagochi.users
    ALTER COLUMN password_hashed TYPE TEXT;

CREATE TABLE tamagochi.refresh_sessions (
    id VARCHAR(64) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES tamagochi.users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX refresh_sessions_expires_at_idx
    ON tamagochi.refresh_sessions (expires_at);
