package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
)

func mapCRUDError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	var state sqlStateError
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "23505", "23503", "23001":
			return fmt.Errorf("%w: %v", domain.ErrConflict, err)
		case "23514", "23502", "22003", "22P02", "22001":
			return fmt.Errorf("%w: %v", domain.ErrValidation, err)
		}
	}
	return err
}

func requireAffected(result sql.Result, err error) error {
	if err != nil {
		return mapCRUDError(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func pageSize(limit int) int {
	if limit < 1 || limit > 100 {
		return 20
	}
	return limit
}
