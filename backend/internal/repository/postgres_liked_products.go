package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
)

type PostgresLikedProductStore struct{ db *sql.DB }

const likeColumns = `user_id, product_id, created_at`

func NewPostgresLikedProductStore(db *sql.DB) *PostgresLikedProductStore {
	return &PostgresLikedProductStore{db: db}
}

func (s *PostgresLikedProductStore) Create(ctx context.Context, like products.LikedProduct) (products.LikedProduct, error) {

	err := s.db.QueryRowContext(ctx, `INSERT INTO tamagochi.user_liked_products (user_id, product_id)
 VALUES ($1, $2) RETURNING `+likeColumns, like.UserID, like.ProductID).Scan(&like.UserID, &like.ProductID, &like.CreatedAt)
	if err != nil {
		return products.LikedProduct{}, mapCRUDError(err)
	}
	return like, nil
}

func (s *PostgresLikedProductStore) Find(ctx context.Context, userID, productID int64) (products.LikedProduct, bool, error) {
	var like products.LikedProduct
	err := s.db.QueryRowContext(ctx, `SELECT `+likeColumns+` FROM tamagochi.user_liked_products WHERE user_id=$1 AND product_id=$2`, userID, productID).Scan(&like.UserID, &like.ProductID, &like.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return products.LikedProduct{}, false, nil
	}
	if err != nil {
		return products.LikedProduct{}, false, err
	}
	return like, true, nil
}

func (s *PostgresLikedProductStore) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]products.LikedProduct, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+likeColumns+` FROM tamagochi.user_liked_products WHERE user_id=$1 ORDER BY created_at DESC, product_id LIMIT $2 OFFSET $3`, userID, pageSize(limit), offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	likes := make([]products.LikedProduct, 0)
	for rows.Next() {
		var like products.LikedProduct
		if err := rows.Scan(&like.UserID, &like.ProductID, &like.CreatedAt); err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}
	return likes, rows.Err()
}

func (s *PostgresLikedProductStore) Delete(ctx context.Context, userID, productID int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tamagochi.user_liked_products WHERE user_id=$1 AND product_id=$2`, userID, productID)
	return requireAffected(result, err)
}
