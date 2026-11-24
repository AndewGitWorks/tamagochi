package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
)

type PostgresProductStore struct{ db *sql.DB }

func NewPostgresProductStore(db *sql.DB) *PostgresProductStore { return &PostgresProductStore{db: db} }

const productColumns = `id, title, price, seria_title, COALESCE(bio,''), COALESCE(description,''), COALESCE(specifications,'null'::jsonb)`

func (s *PostgresProductStore) Create(ctx context.Context, product products.Product) (products.Product, error) {
	if product.SeriaTitle == "" {
		product.SeriaTitle = "NEWBIE"
	}
	if len(product.Specifications) == 0 {
		product.Specifications = []byte("null")
	}
	if err := product.Validate(); err != nil {
		return products.Product{}, err
	}
	err := s.db.QueryRowContext(ctx, `INSERT INTO tamagochi.products (title, price, seria_title, bio, description, specifications)
 VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+productColumns, product.Title, product.Price, product.SeriaTitle, product.Bio, product.Description, product.Specifications).Scan(&product.ID, &product.Title, &product.Price, &product.SeriaTitle, &product.Bio, &product.Description, &product.Specifications)
	if err != nil {
		return products.Product{}, mapCRUDError(err)
	}
	return product, nil
}

func (s *PostgresProductStore) FindByID(ctx context.Context, id int64) (products.Product, bool, error) {
	var product products.Product
	err := s.db.QueryRowContext(ctx, `SELECT `+productColumns+` FROM tamagochi.products WHERE id=$1`, id).Scan(&product.ID, &product.Title, &product.Price, &product.SeriaTitle, &product.Bio, &product.Description, &product.Specifications)
	if errors.Is(err, sql.ErrNoRows) {
		return products.Product{}, false, nil
	}
	if err != nil {
		return products.Product{}, false, err
	}
	return product, true, nil
}

func (s *PostgresProductStore) List(ctx context.Context, limit, offset int) ([]products.Product, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+productColumns+` FROM tamagochi.products ORDER BY id LIMIT $1 OFFSET $2`, pageSize(limit), offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]products.Product, 0)
	for rows.Next() {
		var product products.Product
		if err := rows.Scan(&product.ID, &product.Title, &product.Price, &product.SeriaTitle, &product.Bio, &product.Description, &product.Specifications); err != nil {
			return nil, err
		}
		result = append(result, product)
	}
	return result, rows.Err()
}

func (s *PostgresProductStore) Update(ctx context.Context, product products.Product) (products.Product, error) {
	if product.SeriaTitle == "" {
		product.SeriaTitle = "NEWBIE"
	}
	if len(product.Specifications) == 0 {
		product.Specifications = []byte("null")
	}
	if err := product.Validate(); err != nil {
		return products.Product{}, err
	}
	err := s.db.QueryRowContext(ctx, `UPDATE tamagochi.products SET title=$1, price=$2, seria_title=$3, bio=$4, description=$5, specifications=$6 WHERE id=$7 RETURNING `+productColumns, product.Title, product.Price, product.SeriaTitle, product.Bio, product.Description, product.Specifications, product.ID).Scan(&product.ID, &product.Title, &product.Price, &product.SeriaTitle, &product.Bio, &product.Description, &product.Specifications)
	if err != nil {
		return products.Product{}, mapCRUDError(err)
	}
	return product, nil
}

func (s *PostgresProductStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tamagochi.products WHERE id=$1`, id)
	return requireAffected(result, err)
}
