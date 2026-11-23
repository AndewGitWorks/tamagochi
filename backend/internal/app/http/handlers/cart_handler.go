package http

import (
	"net/http"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
)

// CartHandler exposes the active PENDING order as a shopping cart. It contains
// no business decisions; those remain in orders.Service.
type CartHandler struct {
	service *orders.Service
}

func NewCartHandler(service *orders.Service) *CartHandler {
	return &CartHandler{service: service}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	cart, err := h.service.GetCart(r.Context(), currentUser(r).ID)
	respond(w, http.StatusOK, cart, err)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProductID int64 `json:"product_id"`
		Quantity  int64 `json:"quantity"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	cart, err := h.service.AddToCart(r.Context(), currentUser(r).ID, input.ProductID, input.Quantity)
	respond(w, http.StatusOK, cart, err)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var input struct {
		Quantity int64 `json:"quantity"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	cart, err := h.service.UpdateCartItem(r.Context(), currentUser(r).ID, itemID, input.Quantity)
	respond(w, http.StatusOK, cart, err)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	itemID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	cart, err := h.service.RemoveFromCart(r.Context(), currentUser(r).ID, itemID)
	respond(w, http.StatusOK, cart, err)
}

func (h *CartHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DeliveryAddress string `json:"delivery_address"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	order, err := h.service.Checkout(r.Context(), currentUser(r).ID, input.DeliveryAddress)
	respond(w, http.StatusOK, order, err)
}
