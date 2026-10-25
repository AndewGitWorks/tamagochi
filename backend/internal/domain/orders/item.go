package orders

import (
	"context"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
)

type Item struct {
	ID        int64  `json:"id"`
	OrderID   int64  `json:"order_id"`
	ProductID int64  `json:"product_id"`
	Quantity  int64  `json:"quantity"`
	Price     string `json:"price"`
}

type ItemStore interface {
	Create(context.Context, Item) (Item, error)
	FindByID(context.Context, int64, int64) (Item, bool, error)
	ListByOrder(context.Context, int64, int, int) ([]Item, error)
	Update(context.Context, Item) (Item, error)
	Delete(context.Context, int64, int64) error
}

func (item Item) Validate() error {
	if item.OrderID <= 0 || item.ProductID <= 0 || item.Quantity <= 0 {
		return domain.ErrValidation
	}
	return nil
}
