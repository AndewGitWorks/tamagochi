package http

import (
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"net/http"
)

type LikedProductHandler struct {
	Likes products.LikedProductStore
	auth  *users.AuthService
}

func NewLikedProductHandler(likes products.LikedProductStore, auth *users.AuthService) *LikedProductHandler {
	return &LikedProductHandler{Likes: likes, auth: auth}
}

func (h *LikedProductHandler) CreateLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	like, err := h.Likes.Create(r.Context(), products.LikedProduct{UserID: currentUser(r).ID, ProductID: productID})
	respond(w, http.StatusCreated, like, err)
}

func (h *LikedProductHandler) GetLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	like, found, err := h.Likes.Find(r.Context(), currentUser(r).ID, productID)
	respondFound(w, like, found, err)
}

func (h *LikedProductHandler) ListLikes(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.Likes.ListByUser(r.Context(), currentUser(r).ID, limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *LikedProductHandler) DeleteLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	respondDeleted(w, h.Likes.Delete(r.Context(), currentUser(r).ID, productID))
}
