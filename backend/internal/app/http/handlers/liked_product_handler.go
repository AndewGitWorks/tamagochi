package http

import (
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"net/http"
)

func (h *crudHandler) createLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	like, err := h.Likes.Create(r.Context(), products.LikedProduct{UserID: currentUser(r).ID, ProductID: productID})
	respond(w, http.StatusCreated, like, err)
}

func (h *crudHandler) getLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	like, found, err := h.Likes.Find(r.Context(), currentUser(r).ID, productID)
	respondFound(w, like, found, err)
}

func (h *crudHandler) listLikes(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.Likes.ListByUser(r.Context(), currentUser(r).ID, limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *crudHandler) deleteLike(w http.ResponseWriter, r *http.Request) {
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}
	respondDeleted(w, h.Likes.Delete(r.Context(), currentUser(r).ID, productID))
}
