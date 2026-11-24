package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

type PostgresProfileStore struct{ db *sql.DB }

const profileColumns = `id, user_id, COALESCE(bio,''), COALESCE(address,''), (SELECT count(*) FROM tamagochi.orders WHERE user_id=tamagochi.profiles.user_id)`

func NewPostgresProfileStore(db *sql.DB) *PostgresProfileStore { return &PostgresProfileStore{db: db} }

func (s *PostgresProfileStore) Create(ctx context.Context, profile users.Profile) (users.Profile, error) {

	err := s.db.QueryRowContext(ctx, `INSERT INTO tamagochi.profiles (user_id, bio, address)
 VALUES ($1, $2, $3) RETURNING `+profileColumns, profile.UserID, profile.Bio, profile.Address).Scan(&profile.ID, &profile.UserID, &profile.Bio, &profile.Address, &profile.OrdersCount)
	if err != nil {
		return users.Profile{}, mapCRUDError(err)
	}
	return profile, nil
}

func (s *PostgresProfileStore) FindByUserID(ctx context.Context, userID int64) (users.Profile, bool, error) {
	var profile users.Profile
	err := s.db.QueryRowContext(ctx, `SELECT `+profileColumns+` FROM tamagochi.profiles WHERE user_id=$1`, userID).Scan(&profile.ID, &profile.UserID, &profile.Bio, &profile.Address, &profile.OrdersCount)
	if errors.Is(err, sql.ErrNoRows) {
		return users.Profile{}, false, nil
	}
	if err != nil {
		return users.Profile{}, false, err
	}
	return profile, true, nil
}

func (s *PostgresProfileStore) Update(ctx context.Context, profile users.Profile) (users.Profile, error) {

	err := s.db.QueryRowContext(ctx, `UPDATE tamagochi.profiles SET bio=$1, address=$2 WHERE user_id=$3 RETURNING `+profileColumns, profile.Bio, profile.Address, profile.UserID).Scan(&profile.ID, &profile.UserID, &profile.Bio, &profile.Address, &profile.OrdersCount)
	if err != nil {
		return users.Profile{}, mapCRUDError(err)
	}
	return profile, nil
}

func (s *PostgresProfileStore) Delete(ctx context.Context, userID int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tamagochi.profiles WHERE user_id=$1`, userID)
	return requireAffected(result, err)
}
