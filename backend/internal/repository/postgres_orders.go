package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
)

type PostgresOrderStore struct{ db *sql.DB }

const orderColumns = `id, user_id, number, total_price, delivery_address, created_at, status`

func NewPostgresOrderStore(db *sql.DB) *PostgresOrderStore { return &PostgresOrderStore{db: db} }

func (s *PostgresOrderStore) Create(ctx context.Context, order orders.Order) (orders.Order, error) {
	if order.Status == "" {
		order.Status = "PENDING"
	}
	if err := order.Validate(); err != nil {
		return orders.Order{}, err
	}
	if order.Status != "PENDING" {
		return orders.Order{}, domain.ErrValidation
	}
	err := s.db.QueryRowContext(ctx, `INSERT INTO tamagochi.orders (user_id, number, status)
 VALUES ($1, $2, $3) RETURNING `+orderColumns, order.UserID, order.Number, order.Status).Scan(&order.ID, &order.UserID, &order.Number, &order.TotalPrice, &order.DeliveryAddress, &order.CreatedAt, &order.Status)
	if err != nil {
		return orders.Order{}, mapCRUDError(err)
	}
	return order, nil
}

func (s *PostgresOrderStore) FindByID(ctx context.Context, id int64) (orders.Order, bool, error) {
	var order orders.Order
	err := s.db.QueryRowContext(ctx, `SELECT `+orderColumns+` FROM tamagochi.orders WHERE id=$1`, id).Scan(&order.ID, &order.UserID, &order.Number, &order.TotalPrice, &order.DeliveryAddress, &order.CreatedAt, &order.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return orders.Order{}, false, nil
	}
	if err != nil {
		return orders.Order{}, false, err
	}
	return order, true, nil
}

func (s *PostgresOrderStore) FindPendingByUser(ctx context.Context, userID int64) (orders.Order, bool, error) {
	var order orders.Order
	err := s.db.QueryRowContext(ctx, `SELECT `+orderColumns+` FROM tamagochi.orders WHERE user_id=$1 AND status='PENDING'`, userID).
		Scan(&order.ID, &order.UserID, &order.Number, &order.TotalPrice, &order.DeliveryAddress, &order.CreatedAt, &order.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return orders.Order{}, false, nil
	}
	if err != nil {
		return orders.Order{}, false, err
	}
	return order, true, nil
}

func (s *PostgresOrderStore) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]orders.Order, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+orderColumns+` FROM tamagochi.orders WHERE user_id=$1 ORDER BY id DESC LIMIT $2 OFFSET $3`, userID, pageSize(limit), offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]orders.Order, 0)
	for rows.Next() {
		var order orders.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.TotalPrice, &order.DeliveryAddress, &order.CreatedAt, &order.Status); err != nil {
			return nil, err
		}
		result = append(result, order)
	}
	return result, rows.Err()
}

func (s *PostgresOrderStore) Update(ctx context.Context, order orders.Order) (orders.Order, error) {
	if !orders.ValidStatus(order.Status) {
		return orders.Order{}, domain.ErrValidation
	}
	err := s.db.QueryRowContext(ctx, `UPDATE tamagochi.orders SET status=$1 WHERE id=$2 RETURNING `+orderColumns, order.Status, order.ID).Scan(&order.ID, &order.UserID, &order.Number, &order.TotalPrice, &order.DeliveryAddress, &order.CreatedAt, &order.Status)
	if err != nil {
		return orders.Order{}, mapCRUDError(err)
	}
	return order, nil
}

func (s *PostgresOrderStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tamagochi.orders WHERE id=$1 AND status IN ('PENDING','CANCELLED')`, id)
	return requireAffected(result, err)
}

func (s *PostgresOrderStore) Checkout(ctx context.Context, id, userID int64, address string) (orders.Order, error) {
	var order orders.Order
	err := s.db.QueryRowContext(ctx, `UPDATE tamagochi.orders
 SET status='AWAITING_PAYMENT', delivery_address=$3
 WHERE id=$1 AND user_id=$2 AND status='PENDING'
   AND EXISTS (SELECT 1 FROM tamagochi.order_items WHERE order_id=$1)
 RETURNING `+orderColumns, id, userID, address).
		Scan(&order.ID, &order.UserID, &order.Number, &order.TotalPrice, &order.DeliveryAddress, &order.CreatedAt, &order.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return orders.Order{}, domain.ErrOrderLocked
	}
	if err != nil {
		return orders.Order{}, mapCRUDError(err)
	}
	return order, nil
}

// Cancel checks ownership and status in the same statement as the update.
func (s *PostgresOrderStore) Cancel(ctx context.Context, id, userID int64) (orders.Order, error) {
	var order orders.Order
	err := s.db.QueryRowContext(ctx, `UPDATE tamagochi.orders SET status='CANCELLED'
	 WHERE id=$1 AND user_id=$2 AND status IN ('PENDING','AWAITING_PAYMENT','PAYMENT_FAILED')
 RETURNING `+orderColumns, id, userID).
		Scan(&order.ID, &order.UserID, &order.Number, &order.TotalPrice, &order.DeliveryAddress, &order.CreatedAt, &order.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return orders.Order{}, domain.ErrOrderLocked
	}
	if err != nil {
		return orders.Order{}, mapCRUDError(err)
	}
	return order, nil
}
