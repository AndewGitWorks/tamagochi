package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

type PostgresUserStore struct{ db *sql.DB }

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore { return &PostgresUserStore{db: db} }

const userColumns = `id, name, surname, login, email, password_hashed, phone_number, version, role`

func (s *PostgresUserStore) Create(ctx context.Context, user users.User) (users.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return users.User{}, fmt.Errorf("begin user registration: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if user.Role == "" {
		user.Role = "USER"
	}
	if user.Version == 0 {
		user.Version = 1
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO tamagochi.users
		(name, surname, login, email, password_hashed, phone_number, version, role)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		user.Name, user.Surname, user.Login, user.Email, user.PasswordHash, user.PhoneNumber, user.Version, user.Role).Scan(&user.ID)
	if err != nil {
		return users.User{}, mapConflict(fmt.Errorf("insert user: %w", err))
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO tamagochi.profiles (user_id) VALUES ($1)`, user.ID); err != nil {
		return users.User{}, fmt.Errorf("create initial user profile: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return users.User{}, fmt.Errorf("commit user registration: %w", err)
	}
	return user, nil
}

func (s *PostgresUserStore) FindByLogin(ctx context.Context, login string) (users.User, bool, error) {
	return s.find(ctx, `SELECT `+userColumns+` FROM tamagochi.users WHERE login=$1`, login)
}

func (s *PostgresUserStore) FindByID(ctx context.Context, id int64) (users.User, bool, error) {
	return s.find(ctx, `SELECT `+userColumns+` FROM tamagochi.users WHERE id=$1`, id)
}

func (s *PostgresUserStore) find(ctx context.Context, query string, arg any) (users.User, bool, error) {
	var user users.User
	err := s.db.QueryRowContext(ctx, query, arg).Scan(&user.ID, &user.Name, &user.Surname, &user.Login, &user.Email,
		&user.PasswordHash, &user.PhoneNumber, &user.Version, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return users.User{}, false, nil
	}
	if err != nil {
		return users.User{}, false, fmt.Errorf("read user: %w", err)
	}
	return user, true, nil
}

func (s *PostgresUserStore) Update(ctx context.Context, user users.User) (users.User, error) {
	if user.Role == "" {
		user.Role = "USER"
	}
	row := s.db.QueryRowContext(ctx, `UPDATE tamagochi.users SET name=$1,surname=$2,login=$3,email=$4,
		password_hashed=$5,phone_number=$6,role=$7,version=version+1 WHERE id=$8
		RETURNING id,name,surname,login,email,password_hashed,phone_number,version,role`,
		user.Name, user.Surname, user.Login, user.Email, user.PasswordHash, user.PhoneNumber, user.Role, user.ID)
	var updated users.User
	err := row.Scan(&updated.ID, &updated.Name, &updated.Surname, &updated.Login, &updated.Email, &updated.PasswordHash, &updated.PhoneNumber, &updated.Version, &updated.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return users.User{}, users.ErrUserNotFound
	}
	if err != nil {
		return users.User{}, mapConflict(fmt.Errorf("update user: %w", err))
	}
	return updated, nil
}

func (s *PostgresUserStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tamagochi.users WHERE id=$1`, id)
	if err != nil {
		return mapCRUDError(fmt.Errorf("delete user: %w", err))
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if count == 0 {
		return users.ErrUserNotFound
	}
	return nil
}

type sqlStateError interface{ SQLState() string }

func mapConflict(err error) error {
	var stateErr sqlStateError
	if errors.As(err, &stateErr) && stateErr.SQLState() == "23505" {
		return fmt.Errorf("%w: %v", users.ErrUserConflict, err)
	}
	return err
}
