package integration_test

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/database"
	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/security"
	"github.com/AndewGitWorks/tamagochi/backend/internal/repository"
)

func TestPostgresRegistrationCreatesProfileAndSupportsUserCRUD(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	db, err := database.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := repository.NewPostgresUserStore(db)
	login := "test" + strconv.FormatInt(time.Now().UnixNano()%1_000_000_000_000, 36)
	input := users.RegisterInput{
		Login: login, Name: "Test", Surname: "User", Email: login + "@example.test",
		PhoneNumber: "+15551234567", Password: "test-password-long-enough",
	}
	service := users.NewAuthService(users.AuthConfig{
		Store: store, RefreshSessions: repository.NewPostgresRefreshSessionStore(db),
		JWTSecret: []byte("integration-test-secret-at-least-32-characters"), Issuer: "test-api",
	})
	pair, err := service.Register(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	user, found, err := store.FindByLogin(ctx, login)
	if err != nil || !found {
		t.Fatalf("registered user not found: found=%v err=%v", found, err)
	}
	// Always remove the unique test user; the profile is deleted by ON DELETE CASCADE.
	t.Cleanup(func() { _ = store.Delete(context.Background(), user.ID) })
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("registration did not issue tokens")
	}
	if _, err := service.Refresh(ctx, pair.RefreshToken); err != nil {
		t.Fatalf("refresh token stored in PostgreSQL could not be consumed: %v", err)
	}
	if _, err := service.Refresh(ctx, pair.RefreshToken); !errors.Is(err, security.ErrInvalidToken) {
		t.Fatalf("expected used refresh token to be rejected, got %v", err)
	}
	var profileCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM tamagochi.profiles WHERE user_id=$1`, user.ID).Scan(&profileCount); err != nil {
		t.Fatal(err)
	}
	if profileCount != 1 {
		t.Fatalf("expected one profile created during registration, got %d", profileCount)
	}

	byID, found, err := store.FindByID(ctx, user.ID)
	if err != nil || !found || byID.Email != input.Email {
		t.Fatalf("FindByID returned user=%#v found=%v err=%v", byID, found, err)
	}
	updated := byID
	updated.Name = "Test Updated"
	updated.PasswordHash = "updated-test-password-hash"
	updated, err = store.Update(ctx, updated)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version <= user.Version || updated.Name != "Test Updated" {
		t.Fatalf("update did not advance user version: before=%#v after=%#v", user, updated)
	}
	if err := store.Delete(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM tamagochi.profiles WHERE user_id=$1`, user.ID).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("expected profile to cascade on user deletion, got %d", remaining)
	}
	if _, found, err := store.FindByID(ctx, user.ID); err != nil || found {
		t.Fatalf("deleted user still exists: found=%v err=%v", found, err)
	}
}
