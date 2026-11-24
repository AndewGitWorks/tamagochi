package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

// PostgresRefreshSessionStore keeps only refresh-token IDs. The JWT itself is
// never stored, so a database leak cannot be used as a bearer credential.
type PostgresRefreshSessionStore struct{ db *sql.DB }

func NewPostgresRefreshSessionStore(db *sql.DB) *PostgresRefreshSessionStore {
	return &PostgresRefreshSessionStore{db: db}
}

func (s *PostgresRefreshSessionStore) Create(ctx context.Context, session users.RefreshSession) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO tamagochi.refresh_sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)`, session.ID, session.UserID, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("save refresh session: %w", err)
	}
	return nil
}

// Consume deletes a non-expired session in one statement. Only one concurrent
// refresh request can receive a successful result for the same token ID.
func (s *PostgresRefreshSessionStore) Consume(ctx context.Context, id string, userID int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tamagochi.refresh_sessions
		WHERE id = $1 AND user_id = $2 AND expires_at > CURRENT_TIMESTAMP`, id, userID)
	if err != nil {
		return false, fmt.Errorf("consume refresh session: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read refresh session result: %w", err)
	}
	return affected == 1, nil
}
