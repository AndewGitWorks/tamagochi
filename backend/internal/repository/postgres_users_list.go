package repository

import (
	"context"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

func (s *PostgresUserStore) List(ctx context.Context, limit, offset int) ([]users.User, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+userColumns+` FROM tamagochi.users ORDER BY id LIMIT $1 OFFSET $2`, pageSize(limit), offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]users.User, 0)
	for rows.Next() {
		var user users.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Surname, &user.Login, &user.Email, &user.PasswordHash, &user.PhoneNumber, &user.Version, &user.Role); err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	return result, rows.Err()
}
