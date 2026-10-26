package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
)

func (h *crudHandler) ownedOrder(w http.ResponseWriter, r *http.Request) (orders.Order, bool) {
	id, ok := pathID(w, r, "orderID")
	if !ok {
		return orders.Order{}, false
	}
	order, found, err := h.Orders.FindByID(r.Context(), id)
	if err != nil {
		writeCRUDError(w, err)
		return orders.Order{}, false
	}
	user := currentUser(r)
	if !found || (order.UserID != user.ID && user.Role != "ADMIN") {
		writeCRUDError(w, domain.ErrNotFound)
		return orders.Order{}, false
	}
	return order, true
}

func (h *crudHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	// Order number, owner, status and total are assigned by the server.
	var input struct{}
	if !decodeJSON(w, r, &input) {
		return
	}
	var number [16]byte
	if _, err := rand.Read(number[:]); err != nil {
		writeCRUDError(w, err)
		return
	}
	order, err := h.Orders.Create(r.Context(), orders.Order{UserID: currentUser(r).ID, Number: hex.EncodeToString(number[:]), Status: "PENDING"})
	respond(w, http.StatusCreated, order, err)
}

func (h *crudHandler) getOrder(w http.ResponseWriter, r *http.Request) {
	order, ok := h.ownedOrder(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *crudHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.Orders.ListByUser(r.Context(), currentUser(r).ID, limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *crudHandler) updateOrder(w http.ResponseWriter, r *http.Request) {
	order, ok := h.ownedOrder(w, r)
	if !ok {
		return
	}
	var input struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if currentUser(r).Role != "ADMIN" {
		if input.Status != "CANCELLED" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only cancellation is allowed"})
			return
		}
		updated, err := h.Orders.Cancel(r.Context(), order.ID, currentUser(r).ID)
		respond(w, http.StatusOK, updated, err)
		return
	}
	order.Status = input.Status
	updated, err := h.Orders.Update(r.Context(), order)
	respond(w, http.StatusOK, updated, err)
}

func (h *crudHandler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	order, ok := h.ownedOrder(w, r)
	if !ok {
		return
	}
	if order.Status != "PENDING" && order.Status != "CANCELLED" {
		writeCRUDError(w, domain.ErrOrderLocked)
		return
	}
	respondDeleted(w, h.Orders.Delete(r.Context(), order.ID))
}

func (h *crudHandler) createItem(w http.ResponseWriter, r *http.Request) {
	order, ok := h.ownedOrder(w, r)
	if !ok {
		return
	}
	var input struct {
		ProductID int64 `json:"product_id"`
		Quantity  int64 `json:"quantity"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	item, err := h.Items.Create(r.Context(), orders.Item{OrderID: order.ID, ProductID: input.ProductID, Quantity: input.Quantity})
	respond(w, http.StatusCreated, item, err)
}

func (h *crudHandler) getItem(w http.ResponseWriter, r *http.Request) {
	order, ok := h.ownedOrder(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	item, found, err := h.Items.FindByID(r.Context(), id, order.ID)
	respondFound(w, item, found, err)
}

func (h *crudHandler) listItems(w http.ResponseWriter, r *http.Request) {
	order, ok := h.ownedOrder(w, r)
	if !ok {
		return
	}
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.Items.ListByOrder(r.Context(), order.ID, limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *crudHandler) updateItem(w http.ResponseWriter, r *http.Request) {
	order, ok := h.ownedOrder(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var input struct {
		Quantity int64 `json:"quantity"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	item, err := h.Items.Update(r.Context(), orders.Item{ID: id, OrderID: order.ID, Quantity: input.Quantity})
	respond(w, http.StatusOK, item, err)
}

func (h *crudHandler) deleteItem(w http.ResponseWriter, r *http.Request) {
	order, ok := h.ownedOrder(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	respondDeleted(w, h.Items.Delete(r.Context(), id, order.ID))
}
