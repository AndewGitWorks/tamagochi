package orders

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
)

// Actor is the minimum identity required to apply order access rules. The HTTP
// layer converts the authenticated user to this type before calling the service.
type Actor struct {
	ID   int64
	Role string
}

func (a Actor) IsAdmin() bool { return a.Role == "ADMIN" }

// Service is the business boundary for carts and orders. Repositories persist
// data and protect transactions; this service owns permissions and state rules.
type Service struct {
	orders Store
	items  ItemStore
	number func() (string, error)
}

// Cart is a pending order together with its current line items. A user has at
// most one cart; the database enforces that invariant with a partial index.
type Cart struct {
	Order Order  `json:"order"`
	Items []Item `json:"items"`
}

func NewService(orders Store, items ItemStore) *Service {
	return &Service{orders: orders, items: items, number: newOrderNumber}
}

func (s *Service) Create(ctx context.Context, userID int64) (Order, error) {
	if userID <= 0 {
		return Order{}, domain.ErrValidation
	}
	number, err := s.number()
	if err != nil {
		return Order{}, err
	}
	return s.orders.Create(ctx, Order{UserID: userID, Number: number, Status: "PENDING"})
}

// OpenCart returns the user's active cart or creates it. If two concurrent
// requests try to create it, the unique database index lets one succeed and
// the other request reads the newly created cart.
func (s *Service) OpenCart(ctx context.Context, userID int64) (Order, error) {
	return s.pendingCart(ctx, userID, true)
}

func (s *Service) GetCart(ctx context.Context, userID int64) (Cart, error) {
	if userID <= 0 {
		return Cart{}, domain.ErrValidation
	}
	order, err := s.pendingCart(ctx, userID, false)
	if err != nil {
		return Cart{}, err
	}
	items, err := s.items.ListByOrder(ctx, order.ID, 100, 0)
	if err != nil {
		return Cart{}, err
	}
	return Cart{Order: order, Items: items}, nil
}

// AddToCart increments an existing line instead of creating duplicate lines
// for the same product. The repository recalculates total_price atomically.
func (s *Service) AddToCart(ctx context.Context, userID, productID, quantity int64) (Cart, error) {
	if userID <= 0 || productID <= 0 || quantity <= 0 {
		return Cart{}, domain.ErrValidation
	}
	order, err := s.pendingCart(ctx, userID, true)
	if err != nil {
		return Cart{}, err
	}
	item, found, err := s.items.FindByProduct(ctx, order.ID, productID)
	if err != nil {
		return Cart{}, err
	}
	if found {
		if quantity > (1<<63-1)-item.Quantity {
			return Cart{}, domain.ErrValidation
		}
		if _, err = s.items.Update(ctx, Item{ID: item.ID, OrderID: order.ID, Quantity: item.Quantity + quantity}); err != nil {
			return Cart{}, err
		}
	} else if _, err = s.items.Create(ctx, Item{OrderID: order.ID, ProductID: productID, Quantity: quantity}); err != nil {
		return Cart{}, err
	}
	return s.GetCart(ctx, userID)
}

func (s *Service) UpdateCartItem(ctx context.Context, userID, itemID, quantity int64) (Cart, error) {
	if userID <= 0 || itemID <= 0 || quantity <= 0 {
		return Cart{}, domain.ErrValidation
	}
	order, err := s.pendingCart(ctx, userID, false)
	if err != nil {
		return Cart{}, err
	}
	if _, err = s.items.Update(ctx, Item{ID: itemID, OrderID: order.ID, Quantity: quantity}); err != nil {
		return Cart{}, err
	}
	return s.GetCart(ctx, userID)
}

func (s *Service) RemoveFromCart(ctx context.Context, userID, itemID int64) (Cart, error) {
	if userID <= 0 || itemID <= 0 {
		return Cart{}, domain.ErrValidation
	}
	order, err := s.pendingCart(ctx, userID, false)
	if err != nil {
		return Cart{}, err
	}
	if err = s.items.Delete(ctx, itemID, order.ID); err != nil {
		return Cart{}, err
	}
	return s.GetCart(ctx, userID)
}

func (s *Service) Get(ctx context.Context, actor Actor, orderID int64) (Order, error) {
	return s.owned(ctx, actor, orderID)
}

func (s *Service) List(ctx context.Context, userID int64, limit, offset int) ([]Order, error) {
	if userID <= 0 {
		return nil, domain.ErrValidation
	}
	return s.orders.ListByUser(ctx, userID, limit, offset)
}

func (s *Service) ChangeStatus(ctx context.Context, actor Actor, orderID int64, target string) (Order, error) {
	order, err := s.owned(ctx, actor, orderID)
	if err != nil {
		return Order{}, err
	}
	if !actor.IsAdmin() {
		if target != "CANCELLED" {
			return Order{}, domain.ErrOrderTransition
		}
		return s.orders.Cancel(ctx, order.ID, actor.ID)
	}
	if !allowedTransition(order.Status, target) {
		return Order{}, domain.ErrOrderTransition
	}
	order.Status = target
	return s.orders.Update(ctx, order)
}

func (s *Service) Delete(ctx context.Context, actor Actor, orderID int64) error {
	order, err := s.owned(ctx, actor, orderID)
	if err != nil {
		return err
	}
	if order.Status != "PENDING" && order.Status != "CANCELLED" {
		return domain.ErrOrderLocked
	}
	return s.orders.Delete(ctx, order.ID)
}

func (s *Service) CreateItem(ctx context.Context, actor Actor, orderID, productID, quantity int64) (Item, error) {
	if _, err := s.owned(ctx, actor, orderID); err != nil {
		return Item{}, err
	}
	return s.items.Create(ctx, Item{OrderID: orderID, ProductID: productID, Quantity: quantity})
}

func (s *Service) GetItem(ctx context.Context, actor Actor, orderID, itemID int64) (Item, bool, error) {
	if _, err := s.owned(ctx, actor, orderID); err != nil {
		return Item{}, false, err
	}
	return s.items.FindByID(ctx, itemID, orderID)
}

func (s *Service) ListItems(ctx context.Context, actor Actor, orderID int64, limit, offset int) ([]Item, error) {
	if _, err := s.owned(ctx, actor, orderID); err != nil {
		return nil, err
	}
	return s.items.ListByOrder(ctx, orderID, limit, offset)
}

func (s *Service) UpdateItem(ctx context.Context, actor Actor, orderID, itemID, quantity int64) (Item, error) {
	if _, err := s.owned(ctx, actor, orderID); err != nil {
		return Item{}, err
	}
	return s.items.Update(ctx, Item{ID: itemID, OrderID: orderID, Quantity: quantity})
}

func (s *Service) DeleteItem(ctx context.Context, actor Actor, orderID, itemID int64) error {
	if _, err := s.owned(ctx, actor, orderID); err != nil {
		return err
	}
	return s.items.Delete(ctx, itemID, orderID)
}

func (s *Service) owned(ctx context.Context, actor Actor, orderID int64) (Order, error) {
	if actor.ID <= 0 || orderID <= 0 {
		return Order{}, domain.ErrValidation
	}
	order, found, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return Order{}, err
	}
	if !found || (!actor.IsAdmin() && order.UserID != actor.ID) {
		// Do not disclose whether an order belonging to another user exists.
		return Order{}, domain.ErrNotFound
	}
	return order, nil
}

func (s *Service) pendingCart(ctx context.Context, userID int64, create bool) (Order, error) {
	if userID <= 0 {
		return Order{}, domain.ErrValidation
	}
	cart, found, err := s.orders.FindPendingByUser(ctx, userID)
	if err != nil {
		return Order{}, err
	}
	if found {
		return cart, nil
	}
	if !create {
		return Order{}, domain.ErrNotFound
	}
	cart, err = s.Create(ctx, userID)
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, domain.ErrConflict) {
		return Order{}, err
	}
	cart, found, findErr := s.orders.FindPendingByUser(ctx, userID)
	if findErr != nil {
		return Order{}, findErr
	}
	if !found {
		return Order{}, err
	}
	return cart, nil
}

func allowedTransition(from, to string) bool {
	switch from {
	case "PENDING":
		return to == "PAID" || to == "PAYMENT_FAILED" || to == "CANCELLED"
	case "PAYMENT_FAILED":
		return to == "PENDING" || to == "CANCELLED"
	case "PAID":
		return to == "DELIVERING"
	case "DELIVERING":
		return to == "DELIVERED"
	}
	return false
}

func newOrderNumber() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

// Keep errors.Is useful for callers that want to identify terminal-state
// failures without leaking persistence details.
func IsTransitionError(err error) bool { return errors.Is(err, domain.ErrOrderTransition) }
