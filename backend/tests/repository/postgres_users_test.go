package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"github.com/AndewGitWorks/tamagochi/backend/internal/repository"
	"github.com/DATA-DOG/go-sqlmock"
)

func newStore(t *testing.T) (*repository.PostgresUserStore, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return repository.NewPostgresUserStore(db), mock
}

func TestCreateUserAndProfileCommitTogether(t *testing.T) {
	store, mock := newStore(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO tamagochi.users")).WithArgs("Maria", "Rossi", "maria", "maria@example.com", "hashed-password", "+15551234567", int64(1), "USER").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO tamagochi.profiles (user_id) VALUES ($1)")).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	created, err := store.Create(context.Background(), users.User{Name: "Maria", Surname: "Rossi", Login: "maria", Email: "maria@example.com", PasswordHash: "hashed-password", PhoneNumber: "+15551234567"})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 42 || created.Version != 1 || created.Role != "USER" {
		t.Fatalf("unexpected created user: %#v", created)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRollsBackWhenProfileInsertFails(t *testing.T) {
	store, mock := newStore(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO tamagochi.users")).WithArgs("Maria", "Rossi", "maria", "maria@example.com", "hashed-password", "+15551234567", int64(1), "USER").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO tamagochi.profiles (user_id) VALUES ($1)")).WithArgs(int64(42)).WillReturnError(errors.New("profile insert failed"))
	mock.ExpectRollback()
	_, err := store.Create(context.Background(), users.User{Name: "Maria", Surname: "Rossi", Login: "maria", Email: "maria@example.com", PasswordHash: "hashed-password", PhoneNumber: "+15551234567"})
	if err == nil {
		t.Fatal("expected registration transaction to fail")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindByLogin(t *testing.T) {
	store, mock := newStore(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT " + "id, name, surname, login, email, password_hashed, phone_number, version, role" + " FROM tamagochi.users WHERE login=$1")).WithArgs("maria").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "surname", "login", "email", "password_hashed", "phone_number", "version", "role"}).AddRow(42, "Maria", "Rossi", "maria", "maria@example.com", "hash", "+15551234567", 1, "USER"))
	user, found, err := store.FindByLogin(context.Background(), "maria")
	if err != nil || !found || user.ID != 42 || user.Login != "maria" {
		t.Fatalf("user=%#v found=%v err=%v", user, found, err)
	}
}

func TestFindByIDNotFound(t *testing.T) {
	store, mock := newStore(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, surname, login, email, password_hashed, phone_number, version, role FROM tamagochi.users WHERE id=$1")).WithArgs(int64(42)).WillReturnError(sql.ErrNoRows)
	_, found, err := store.FindByID(context.Background(), 42)
	if err != nil || found {
		t.Fatalf("expected not found, found=%v err=%v", found, err)
	}
}

func TestUpdateUser(t *testing.T) {
	store, mock := newStore(t)
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE tamagochi.users SET name=$1,surname=$2,login=$3,email=$4,")).WithArgs("Maria Updated", "Rossi", "maria", "maria@example.com", "new-hash", "+15551234567", "USER", int64(42)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "surname", "login", "email", "password_hashed", "phone_number", "version", "role"}).AddRow(42, "Maria Updated", "Rossi", "maria", "maria@example.com", "new-hash", "+15551234567", 2, "USER"))
	user, err := store.Update(context.Background(), users.User{ID: 42, Name: "Maria Updated", Surname: "Rossi", Login: "maria", Email: "maria@example.com", PasswordHash: "new-hash", PhoneNumber: "+15551234567"})
	if err != nil || user.Version != 2 || user.Name != "Maria Updated" {
		t.Fatalf("user=%#v err=%v", user, err)
	}
}

func TestDeleteUserAndMissingUser(t *testing.T) {
	store, mock := newStore(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM tamagochi.users WHERE id=$1")).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.Delete(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM tamagochi.users WHERE id=$1")).WithArgs(int64(99)).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := store.Delete(context.Background(), 99); !errors.Is(err, users.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
