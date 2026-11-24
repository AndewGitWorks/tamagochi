package products

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
)

type Image struct {
	ID         int64     `json:"id"`
	ProductID  int64     `json:"product_id"`
	StorageKey string    `json:"storage_key"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}

type ImageStore interface {
	Create(context.Context, Image) (Image, error)
	FindByID(context.Context, int64, int64) (Image, bool, error)
	ListByProduct(context.Context, int64, int, int) ([]Image, error)
	Update(context.Context, Image) (Image, error)
	Delete(context.Context, int64, int64) error
}

func (image Image) Validate() error {
	if image.ProductID <= 0 || strings.TrimSpace(image.StorageKey) == "" || utf8.RuneCountInString(image.StorageKey) > 500 || image.SortOrder < 0 {
		return domain.ErrValidation
	}
	return nil
}
