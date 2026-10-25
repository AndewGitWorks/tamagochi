package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
)

type PostgresOrderItemStore struct{ db *sql.DB }

func NewPostgresOrderItemStore(db *sql.DB) *PostgresOrderItemStore {
	return &PostgresOrderItemStore{db: db}
}

const itemColumns = `id, order_id, product_id, quantity, price`

func (s *PostgresOrderItemStore) Create(ctx context.Context, item orders.Item) (orders.Item, error) {
	if err := item.Validate(); err != nil {
		return orders.Item{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return orders.Item{}, err
	}
	defer tx.Rollback()
	if err = lockPendingOrder(ctx, tx, item.OrderID); err != nil {
		return orders.Item{}, err
	}
	// Keep the current product price as a snapshot; the client cannot set it.
	err = tx.QueryRowContext(ctx, `INSERT INTO tamagochi.order_items (order_id, product_id, quantity, price)
 SELECT $1, id, $3, price FROM tamagochi.products WHERE id=$2
 RETURNING `+itemColumns, item.OrderID, item.ProductID, item.Quantity).
		Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
	if err != nil {
		return orders.Item{}, mapCRUDError(err)
	}
	if err = updateOrderTotal(ctx, tx, item.OrderID); err != nil {
		return orders.Item{}, err
	}
	if err = tx.Commit(); err != nil {
		return orders.Item{}, mapCRUDError(err)
	}
	return item, nil
}

func (s *PostgresOrderItemStore) FindByID(ctx context.Context, id, orderID int64) (orders.Item, bool, error) {
	var item orders.Item
	err := s.db.QueryRowContext(ctx, `SELECT `+itemColumns+` FROM tamagochi.order_items WHERE id=$1 AND order_id=$2`, id, orderID).
		Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
	if errors.Is(err, sql.ErrNoRows) {
		return orders.Item{}, false, nil
	}
	if err != nil {
		return orders.Item{}, false, err
	}
	return item, true, nil
}

func (s *PostgresOrderItemStore) ListByOrder(ctx context.Context, orderID int64, limit, offset int) ([]orders.Item, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+itemColumns+` FROM tamagochi.order_items WHERE order_id=$1 ORDER BY id LIMIT $2 OFFSET $3`, orderID, pageSize(limit), offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]orders.Item, 0)
	for rows.Next() {
		var item orders.Item
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresOrderItemStore) Update(ctx context.Context, item orders.Item) (orders.Item, error) {
	if item.ID <= 0 || item.OrderID <= 0 || item.Quantity <= 0 {
		return orders.Item{}, domain.ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return orders.Item{}, err
	}
	defer tx.Rollback()
	if err = lockPendingOrder(ctx, tx, item.OrderID); err != nil {
		return orders.Item{}, err
	}
	// Product and price stay unchanged. To replace a product, remove the line and add a new one.
	err = tx.QueryRowContext(ctx, `UPDATE tamagochi.order_items SET quantity=$1 WHERE id=$2 AND order_id=$3 RETURNING `+itemColumns,
		item.Quantity, item.ID, item.OrderID).Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
	if err != nil {
		return orders.Item{}, mapCRUDError(err)
	}
	if err = updateOrderTotal(ctx, tx, item.OrderID); err != nil {
		return orders.Item{}, err
	}
	if err = tx.Commit(); err != nil {
		return orders.Item{}, mapCRUDError(err)
	}
	return item, nil
}

func (s *PostgresOrderItemStore) Delete(ctx context.Context, id, orderID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = lockPendingOrder(ctx, tx, orderID); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM tamagochi.order_items WHERE id=$1 AND order_id=$2`, id, orderID)
	if err = requireAffected(result, err); err != nil {
		return err
	}
	if err = updateOrderTotal(ctx, tx, orderID); err != nil {
		return err
	}
	return mapCRUDError(tx.Commit())
}

func lockPendingOrder(ctx context.Context, tx *sql.Tx, id int64) error {
	var status string
	err := tx.QueryRowContext(ctx, `SELECT status FROM tamagochi.orders WHERE id=$1 FOR UPDATE`, id).Scan(&status)
	if err != nil {
		return mapCRUDError(err)
	}
	if status != "PENDING" {
		return domain.ErrOrderLocked
	}
	return nil
}

func updateOrderTotal(ctx context.Context, tx *sql.Tx, orderID int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE tamagochi.orders SET total_price=
 (SELECT COALESCE(SUM(quantity * price), 0) FROM tamagochi.order_items WHERE order_id=$1) WHERE id=$1`, orderID)
	return mapCRUDError(err)
}
