package orders

import (
	"context"
	"errors"
	"testing"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
)

type serviceOrderStore struct {
	order         Order
	updateCalls   int
	cancelCalls   int
	checkoutCalls int
}

func (s *serviceOrderStore) Create(_ context.Context, order Order) (Order, error) {
	s.order = order
	s.order.ID = 1
	return s.order, nil
}

func (s *serviceOrderStore) FindByID(_ context.Context, id int64) (Order, bool, error) {
	if id != s.order.ID {
		return Order{}, false, nil
	}
	return s.order, true, nil
}

func (s *serviceOrderStore) FindPendingByUser(_ context.Context, userID int64) (Order, bool, error) {
	if userID != s.order.UserID || s.order.Status != "PENDING" {
		return Order{}, false, nil
	}
	return s.order, true, nil
}

func (s *serviceOrderStore) ListByUser(context.Context, int64, int, int) ([]Order, error) {
	return nil, nil
}

func (s *serviceOrderStore) Checkout(_ context.Context, id, userID int64, address string) (Order, error) {
	s.checkoutCalls++
	if id != s.order.ID || userID != s.order.UserID || s.order.Status != "PENDING" {
		return Order{}, domain.ErrOrderLocked
	}
	s.order.Status = "AWAITING_PAYMENT"
	s.order.DeliveryAddress = address
	return s.order, nil
}

func (s *serviceOrderStore) Update(_ context.Context, order Order) (Order, error) {
	s.updateCalls++
	s.order = order
	return order, nil
}

func (s *serviceOrderStore) Cancel(_ context.Context, id, userID int64) (Order, error) {
	s.cancelCalls++
	if id != s.order.ID || userID != s.order.UserID || s.order.Status != "PENDING" {
		return Order{}, domain.ErrOrderLocked
	}
	s.order.Status = "CANCELLED"
	return s.order, nil
}

func (s *serviceOrderStore) Delete(context.Context, int64) error { return nil }

type serviceItemStore struct {
	nextID int64
	items  map[int64]Item
}

func (s *serviceItemStore) Create(_ context.Context, item Item) (Item, error) {
	s.nextID++
	item.ID = s.nextID
	s.items[item.ID] = item
	return item, nil
}

func (s *serviceItemStore) FindByID(_ context.Context, id, orderID int64) (Item, bool, error) {
	item, found := s.items[id]
	return item, found && item.OrderID == orderID, nil
}

func (s *serviceItemStore) FindByProduct(_ context.Context, orderID, productID int64) (Item, bool, error) {
	for _, item := range s.items {
		if item.OrderID == orderID && item.ProductID == productID {
			return item, true, nil
		}
	}
	return Item{}, false, nil
}

func (s *serviceItemStore) ListByOrder(_ context.Context, orderID int64, _, _ int) ([]Item, error) {
	result := make([]Item, 0)
	for _, item := range s.items {
		if item.OrderID == orderID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *serviceItemStore) Update(_ context.Context, item Item) (Item, error) {
	existing, found := s.items[item.ID]
	if !found || existing.OrderID != item.OrderID {
		return Item{}, domain.ErrNotFound
	}
	existing.Quantity = item.Quantity
	s.items[item.ID] = existing
	return existing, nil
}

func (s *serviceItemStore) Delete(_ context.Context, id, orderID int64) error {
	item, found := s.items[id]
	if !found || item.OrderID != orderID {
		return domain.ErrNotFound
	}
	delete(s.items, id)
	return nil
}

func TestServiceOnlyLetsOwnerCancelPendingOrder(t *testing.T) {
	store := &serviceOrderStore{order: Order{ID: 7, UserID: 12, Status: "PENDING"}}
	service := NewService(store, nil)

	updated, err := service.ChangeStatus(context.Background(), Actor{ID: 12, Role: "USER"}, 7, "CANCELLED")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "CANCELLED" || store.cancelCalls != 1 || store.updateCalls != 0 {
		t.Fatalf("unexpected cancellation: %+v, cancel=%d update=%d", updated, store.cancelCalls, store.updateCalls)
	}

	_, err = service.ChangeStatus(context.Background(), Actor{ID: 99, Role: "USER"}, 7, "CANCELLED")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected hidden foreign order, got %v", err)
	}
}

func TestServiceRejectsUserPaymentAndInvalidAdminTransition(t *testing.T) {
	store := &serviceOrderStore{order: Order{ID: 7, UserID: 12, Status: "PENDING"}}
	service := NewService(store, nil)

	_, err := service.ChangeStatus(context.Background(), Actor{ID: 12, Role: "USER"}, 7, "PAID")
	if !errors.Is(err, domain.ErrOrderTransition) {
		t.Fatalf("expected user payment rejection, got %v", err)
	}
	if store.updateCalls != 0 || store.cancelCalls != 0 {
		t.Fatal("forbidden status change reached persistence")
	}

	_, err = service.ChangeStatus(context.Background(), Actor{ID: 1, Role: "ADMIN"}, 7, "DELIVERED")
	if !errors.Is(err, domain.ErrOrderTransition) {
		t.Fatalf("expected skipped transition rejection, got %v", err)
	}
}

func TestServiceAllowsOnlyValidAdminStatusSequence(t *testing.T) {
	store := &serviceOrderStore{order: Order{ID: 7, UserID: 12, Status: "PAID"}}
	service := NewService(store, nil)
	admin := Actor{ID: 1, Role: "ADMIN"}

	for _, status := range []string{"DELIVERING", "DELIVERED"} {
		updated, err := service.ChangeStatus(context.Background(), admin, 7, status)
		if err != nil {
			t.Fatalf("transition to %s: %v", status, err)
		}
		if updated.Status != status {
			t.Fatalf("status = %s, want %s", updated.Status, status)
		}
	}
	if store.updateCalls != 2 {
		t.Fatalf("updates = %d, want 2", store.updateCalls)
	}
}

func TestServiceAddToCartCreatesOneLineAndThenIncrementsIt(t *testing.T) {
	store := &serviceOrderStore{order: Order{ID: 7, UserID: 12, Status: "PENDING"}}
	items := &serviceItemStore{items: make(map[int64]Item)}
	service := NewService(store, items)

	first, err := service.AddToCart(context.Background(), 12, 44, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 1 || first.Items[0].Quantity != 2 {
		t.Fatalf("first cart: %+v", first)
	}

	second, err := service.AddToCart(context.Background(), 12, 44, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Items) != 1 || second.Items[0].Quantity != 5 {
		t.Fatalf("second cart: %+v", second)
	}
}

func TestServiceCheckoutRequiresItemsAndAllowsPaymentRetry(t *testing.T) {
	store := &serviceOrderStore{order: Order{ID: 7, UserID: 12, Status: "PENDING"}}
	items := &serviceItemStore{items: make(map[int64]Item)}
	service := NewService(store, items)

	_, err := service.Checkout(context.Background(), 12, "Moscow, Test street 1")
	if !errors.Is(err, domain.ErrValidation) || store.checkoutCalls != 0 {
		t.Fatalf("empty checkout: err=%v calls=%d", err, store.checkoutCalls)
	}
	items.items[1] = Item{ID: 1, OrderID: 7, ProductID: 44, Quantity: 1, Price: "10.00"}

	checkedOut, err := service.Checkout(context.Background(), 12, "Moscow, Test street 1")
	if err != nil {
		t.Fatal(err)
	}
	if checkedOut.Status != "AWAITING_PAYMENT" || checkedOut.DeliveryAddress != "Moscow, Test street 1" {
		t.Fatalf("checkout: %+v", checkedOut)
	}

	failed, err := service.Pay(context.Background(), Actor{ID: 12, Role: "USER"}, 7, false)
	if err != nil || failed.Status != "PAYMENT_FAILED" {
		t.Fatalf("failed payment: order=%+v err=%v", failed, err)
	}
	paid, err := service.Pay(context.Background(), Actor{ID: 12, Role: "USER"}, 7, true)
	if err != nil || paid.Status != "PAID" {
		t.Fatalf("payment retry: order=%+v err=%v", paid, err)
	}
}
