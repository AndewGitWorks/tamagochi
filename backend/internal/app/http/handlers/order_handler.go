package http

import (
	"net/http"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

type OrderHandler struct {
	service *orders.Service
	auth    *users.AuthService
}

func NewOrderHandler(service *orders.Service, auth *users.AuthService) *OrderHandler {
	return &OrderHandler{service: service, auth: auth}
}

func orderActor(r *http.Request) orders.Actor {
	user := currentUser(r)
	return orders.Actor{ID: user.ID, Role: user.Role}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var input struct{}
	if !decodeJSON(w, r, &input) {
		return
	}
	order, err := h.service.OpenCart(r.Context(), currentUser(r).ID)
	respond(w, http.StatusOK, order, err)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "orderID")
	if !ok {
		return
	}
	order, err := h.service.Get(r.Context(), orderActor(r), id)
	respond(w, http.StatusOK, order, err)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.service.List(r.Context(), currentUser(r).ID, limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "orderID")
	if !ok {
		return
	}
	var input struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	updated, err := h.service.ChangeStatus(r.Context(), orderActor(r), id, input.Status)
	respond(w, http.StatusOK, updated, err)
}

func (h *OrderHandler) PayOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "orderID")
	if !ok {
		return
	}
	var input struct {
		Approved bool `json:"approved"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	updated, err := h.service.Pay(r.Context(), orderActor(r), id, input.Approved)
	respond(w, http.StatusOK, updated, err)
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "orderID")
	if !ok {
		return
	}
	respondDeleted(w, h.service.Delete(r.Context(), orderActor(r), id))
}

func (h *OrderHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	orderID, ok := pathID(w, r, "orderID")
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
	item, err := h.service.CreateItem(r.Context(), orderActor(r), orderID, input.ProductID, input.Quantity)
	respond(w, http.StatusCreated, item, err)
}

func (h *OrderHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	orderID, ok := pathID(w, r, "orderID")
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	item, found, err := h.service.GetItem(r.Context(), orderActor(r), orderID, id)
	respondFound(w, item, found, err)
}

func (h *OrderHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	orderID, ok := pathID(w, r, "orderID")
	if !ok {
		return
	}
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListItems(r.Context(), orderActor(r), orderID, limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *OrderHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	orderID, ok := pathID(w, r, "orderID")
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
	item, err := h.service.UpdateItem(r.Context(), orderActor(r), orderID, id, input.Quantity)
	respond(w, http.StatusOK, item, err)
}

func (h *OrderHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	orderID, ok := pathID(w, r, "orderID")
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	respondDeleted(w, h.service.DeleteItem(r.Context(), orderActor(r), orderID, id))
}
