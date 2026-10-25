package products

import (
	"context"
	"time"
)

type LikedProduct struct {
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

// The composite key is the whole relation; there is nothing to update.
type LikedProductStore interface {
	Create(context.Context, LikedProduct) (LikedProduct, error)
	Find(context.Context, int64, int64) (LikedProduct, bool, error)
	ListByUser(context.Context, int64, int, int) ([]LikedProduct, error)
	Delete(context.Context, int64, int64) error
}
