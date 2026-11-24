package http

import (
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"net/http"
)

type LikedProductHandler struct {
	service *products.FavoriteService
}

func NewLikedProductHandler(service *products.FavoriteService) *LikedProductHandler {
	return &LikedProductHandler{service: service}
}

func (h *LikedProductHandler) CreateLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	like, err := h.service.Add(r.Context(), currentUser(r).ID, productID)
	respond(w, http.StatusOK, like, err)
}

func (h *LikedProductHandler) GetLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	like, err := h.service.Get(r.Context(), currentUser(r).ID, productID)
	respond(w, http.StatusOK, like, err)
}

func (h *LikedProductHandler) ListLikes(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.service.List(r.Context(), currentUser(r).ID, limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *LikedProductHandler) DeleteLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	respondDeleted(w, h.service.Remove(r.Context(), currentUser(r).ID, productID))
}
