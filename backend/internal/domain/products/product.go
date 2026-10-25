package products

import (
	"context"
	"encoding/json"
	"unicode/utf8"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
)

type Product struct {
	ID             int64           `json:"id"`
	Title          string          `json:"title"`
	Price          string          `json:"price"`
	SeriaTitle     string          `json:"seria_title"`
	Bio            string          `json:"bio"`
	Description    string          `json:"description"`
	Specifications json.RawMessage `json:"specifications"`
}

type Store interface {
	Create(context.Context, Product) (Product, error)
	FindByID(context.Context, int64) (Product, bool, error)
	List(context.Context, int, int) ([]Product, error)
	Update(context.Context, Product) (Product, error)
	Delete(context.Context, int64) error
}

func (p Product) Validate() error {
	if utf8.RuneCountInString(p.Title) < 3 || utf8.RuneCountInString(p.Title) > 100 ||
		utf8.RuneCountInString(p.SeriaTitle) < 3 || utf8.RuneCountInString(p.SeriaTitle) > 100 ||
		!domain.ValidPrice(p.Price) || (len(p.Specifications) > 0 && !json.Valid(p.Specifications)) {
		return domain.ErrValidation
	}
	return nil
}
