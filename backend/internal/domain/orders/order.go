package orders

import (
	"context"
	"time"
	"unicode/utf8"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
)

type Order struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Number          string    `json:"number"`
	TotalPrice      string    `json:"total_price"`
	DeliveryAddress string    `json:"delivery_address"`
	CreatedAt       time.Time `json:"created_at"`
	Status          string    `json:"status"`
}

type Store interface {
	Create(context.Context, Order) (Order, error)
	FindByID(context.Context, int64) (Order, bool, error)
	FindPendingByUser(context.Context, int64) (Order, bool, error)
	ListByUser(context.Context, int64, int, int) ([]Order, error)
	Checkout(context.Context, int64, int64, string) (Order, error)
	Update(context.Context, Order) (Order, error)
	Cancel(context.Context, int64, int64) (Order, error)
	Delete(context.Context, int64) error
}

func ValidStatus(status string) bool {
	switch status {
	case "PENDING", "AWAITING_PAYMENT", "PAID", "PAYMENT_FAILED", "DELIVERING", "DELIVERED", "CANCELLED":
		return true
	}
	return false
}

func (order Order) Validate() error {
	if order.UserID <= 0 || utf8.RuneCountInString(order.Number) < 3 || utf8.RuneCountInString(order.Number) > 64 || !ValidStatus(order.Status) {
		return domain.ErrValidation
	}
	return nil
}
