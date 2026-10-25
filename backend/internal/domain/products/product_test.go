package products_test

import (
	"encoding/json"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"testing"
)

func TestProductValidation(t *testing.T) {
	for _, price := range []string{"0", "0.01", "12.34", "99999999.99"} {
		product := products.Product{Title: "Товар", SeriaTitle: "NEWBIE", Price: price, Specifications: json.RawMessage(`{"color":"red"}`)}
		if err := product.Validate(); err != nil {
			t.Errorf("price=%q: %v", price, err)
		}
	}
	for _, price := range []string{"", "-1", "NaN", "1e2", "0.001", "100000000", "01.00"} {
		if err := (products.Product{Title: "Товар", SeriaTitle: "NEWBIE", Price: price}).Validate(); err == nil {
			t.Errorf("accepted price=%q", price)
		}
	}
	if err := (products.Product{Title: "ab", SeriaTitle: "NEWBIE", Price: "1"}).Validate(); err == nil {
		t.Fatal("accepted a short title")
	}
}
