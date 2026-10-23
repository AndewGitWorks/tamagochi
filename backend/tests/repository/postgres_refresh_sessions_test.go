package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"github.com/AndewGitWorks/tamagochi/backend/internal/repository"
	"github.com/DATA-DOG/go-sqlmock"
)

func newRefreshStore(t *testing.T) (*repository.PostgresRefreshSessionStore, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return repository.NewPostgresRefreshSessionStore(db), mock
}

func TestCreateRefreshSession(t *testing.T) {
	store, mock := newRefreshStore(t)
	expiresAt := time.Now().Add(time.Hour).UTC().Truncate(time.Microsecond)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO tamagochi.refresh_sessions (id, user_id, expires_at)")).
		WithArgs("refresh-id", int64(42), expiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.Create(context.Background(), users.RefreshSession{ID: "refresh-id", UserID: 42, ExpiresAt: expiresAt}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConsumeRefreshSessionIsSingleUse(t *testing.T) {
	store, mock := newRefreshStore(t)
	query := regexp.QuoteMeta("DELETE FROM tamagochi.refresh_sessions WHERE id = $1 AND user_id = $2 AND expires_at > CURRENT_TIMESTAMP")
	mock.ExpectExec(query).WithArgs("refresh-id", int64(42)).WillReturnResult(sqlmock.NewResult(0, 1))
	consumed, err := store.Consume(context.Background(), "refresh-id", 42)
	if err != nil || !consumed {
		t.Fatalf("consumed=%v err=%v", consumed, err)
	}
	mock.ExpectExec(query).WithArgs("refresh-id", int64(42)).WillReturnResult(sqlmock.NewResult(0, 0))
	consumed, err = store.Consume(context.Background(), "refresh-id", 42)
	if err != nil || consumed {
		t.Fatalf("consumed=%v err=%v", consumed, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
