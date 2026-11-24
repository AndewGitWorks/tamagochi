package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
)

type PostgresProductImageStore struct{ db *sql.DB }

const imageColumns = `id, product_id, storage_key, sort_order, created_at`

func NewPostgresProductImageStore(db *sql.DB) *PostgresProductImageStore {
	return &PostgresProductImageStore{db: db}
}

func (s *PostgresProductImageStore) Create(ctx context.Context, image products.Image) (products.Image, error) {
	if err := image.Validate(); err != nil {
		return products.Image{}, err
	}
	err := s.db.QueryRowContext(ctx, `INSERT INTO tamagochi.product_images (product_id, storage_key, sort_order)
 VALUES ($1, $2, $3) RETURNING `+imageColumns, image.ProductID, image.StorageKey, image.SortOrder).Scan(&image.ID, &image.ProductID, &image.StorageKey, &image.SortOrder, &image.CreatedAt)
	if err != nil {
		return products.Image{}, mapCRUDError(err)
	}
	return image, nil
}

func (s *PostgresProductImageStore) FindByID(ctx context.Context, id, productID int64) (products.Image, bool, error) {
	var image products.Image
	err := s.db.QueryRowContext(ctx, `SELECT `+imageColumns+` FROM tamagochi.product_images WHERE id=$1 AND product_id=$2`, id, productID).Scan(&image.ID, &image.ProductID, &image.StorageKey, &image.SortOrder, &image.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return products.Image{}, false, nil
	}
	if err != nil {
		return products.Image{}, false, err
	}
	return image, true, nil
}

func (s *PostgresProductImageStore) ListByProduct(ctx context.Context, productID int64, limit, offset int) ([]products.Image, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+imageColumns+` FROM tamagochi.product_images WHERE product_id=$1 ORDER BY sort_order, id LIMIT $2 OFFSET $3`, productID, pageSize(limit), offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	images := make([]products.Image, 0)
	for rows.Next() {
		var image products.Image
		if err := rows.Scan(&image.ID, &image.ProductID, &image.StorageKey, &image.SortOrder, &image.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, rows.Err()
}

func (s *PostgresProductImageStore) Update(ctx context.Context, image products.Image) (products.Image, error) {
	if err := image.Validate(); err != nil {
		return products.Image{}, err
	}
	err := s.db.QueryRowContext(ctx, `UPDATE tamagochi.product_images SET storage_key=$1, sort_order=$2 WHERE id=$3 AND product_id=$4 RETURNING `+imageColumns, image.StorageKey, image.SortOrder, image.ID, image.ProductID).Scan(&image.ID, &image.ProductID, &image.StorageKey, &image.SortOrder, &image.CreatedAt)
	if err != nil {
		return products.Image{}, mapCRUDError(err)
	}
	return image, nil
}

func (s *PostgresProductImageStore) Delete(ctx context.Context, id, productID int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tamagochi.product_images WHERE id=$1 AND product_id=$2`, id, productID)
	return requireAffected(result, err)
}
