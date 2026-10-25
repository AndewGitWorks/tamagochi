package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

type UserCRUDStore interface {
	users.Store
	List(context.Context, int, int) ([]users.User, error)
}

type CRUDConfig struct {
	Users    UserCRUDStore
	Profiles users.ProfileStore
	Products products.Store
	Images   products.ImageStore
	Orders   orders.Store
	Items    orders.ItemStore
	Likes    products.LikedProductStore
}

type crudHandler struct{ CRUDConfig }
type userContextKey struct{}

func currentUser(r *http.Request) users.User {
	return r.Context().Value(userContextKey{}).(users.User)
}

func requireUser(auth *users.AuthService, admin bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "access token required"})
			return
		}
		user, err := auth.AuthenticateAccess(r.Context(), token)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		if admin && user.Role != "ADMIN" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin role required"})
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
	}
}

func registerCRUD(mux *http.ServeMux, auth *users.AuthService, config CRUDConfig) {
	h := &crudHandler{config}
	mux.HandleFunc("GET /products", h.listProducts)
	mux.HandleFunc("GET /products/{id}", h.getProduct)
	mux.HandleFunc("POST /products", requireUser(auth, true, h.createProduct))
	mux.HandleFunc("PUT /products/{id}", requireUser(auth, true, h.updateProduct))
	mux.HandleFunc("DELETE /products/{id}", requireUser(auth, true, h.deleteProduct))
	mux.HandleFunc("GET /products/{productID}/images", h.listImages)
	mux.HandleFunc("GET /products/{productID}/images/{id}", h.getImage)
	mux.HandleFunc("POST /products/{productID}/images", requireUser(auth, true, h.createImage))
	mux.HandleFunc("PUT /products/{productID}/images/{id}", requireUser(auth, true, h.updateImage))
	mux.HandleFunc("DELETE /products/{productID}/images/{id}", requireUser(auth, true, h.deleteImage))
	mux.HandleFunc("GET /users", requireUser(auth, true, h.listUsers))
	mux.HandleFunc("GET /users/me", requireUser(auth, false, h.getUser))
	mux.HandleFunc("PUT /users/me", requireUser(auth, false, h.updateUser))
	mux.HandleFunc("DELETE /users/me", requireUser(auth, false, h.deleteUser))
	mux.HandleFunc("POST /profiles/me", requireUser(auth, false, h.createProfile))
	mux.HandleFunc("GET /profiles/me", requireUser(auth, false, h.getProfile))
	mux.HandleFunc("PUT /profiles/me", requireUser(auth, false, h.updateProfile))
	mux.HandleFunc("DELETE /profiles/me", requireUser(auth, false, h.deleteProfile))
	mux.HandleFunc("POST /orders", requireUser(auth, false, h.createOrder))
	mux.HandleFunc("GET /orders", requireUser(auth, false, h.listOrders))
	mux.HandleFunc("GET /orders/{orderID}", requireUser(auth, false, h.getOrder))
	mux.HandleFunc("PUT /orders/{orderID}", requireUser(auth, false, h.updateOrder))
	mux.HandleFunc("DELETE /orders/{orderID}", requireUser(auth, false, h.deleteOrder))
	mux.HandleFunc("POST /orders/{orderID}/items", requireUser(auth, false, h.createItem))
	mux.HandleFunc("GET /orders/{orderID}/items", requireUser(auth, false, h.listItems))
	mux.HandleFunc("GET /orders/{orderID}/items/{id}", requireUser(auth, false, h.getItem))
	mux.HandleFunc("PUT /orders/{orderID}/items/{id}", requireUser(auth, false, h.updateItem))
	mux.HandleFunc("DELETE /orders/{orderID}/items/{id}", requireUser(auth, false, h.deleteItem))
	mux.HandleFunc("POST /likes/{productID}", requireUser(auth, false, h.createLike))
	mux.HandleFunc("GET /likes", requireUser(auth, false, h.listLikes))
	mux.HandleFunc("GET /likes/{productID}", requireUser(auth, false, h.getLike))
	mux.HandleFunc("DELETE /likes/{productID}", requireUser(auth, false, h.deleteLike))
}

func pathID(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil || id <= 0 {
		writeCRUDError(w, domain.ErrValidation)
		return 0, false
	}
	return id, true
}

func pagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, offset := 20, 0
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 100 {
			writeCRUDError(w, domain.ErrValidation)
			return 0, 0, false
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 {
			writeCRUDError(w, domain.ErrValidation)
			return 0, 0, false
		}
	}
	return limit, offset, true
}

func writeCRUDError(w http.ResponseWriter, err error) {
	status, message := http.StatusInternalServerError, "an unexpected error occurred"
	switch {
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, users.ErrUserNotFound):
		status, message = http.StatusNotFound, "entity was not found"
	case errors.Is(err, domain.ErrValidation), errors.Is(err, users.ErrValidation):
		status, message = http.StatusBadRequest, "invalid entity data"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, users.ErrUserConflict), errors.Is(err, users.ErrLoginTaken):
		status, message = http.StatusConflict, "entity conflicts with existing data"
	case errors.Is(err, domain.ErrOrderLocked):
		status, message = http.StatusConflict, "only pending orders can be edited"
	}
	writeJSON(w, status, map[string]string{"error": message})
}

func respond(w http.ResponseWriter, status int, value any, err error) {
	if err != nil {
		writeCRUDError(w, err)
		return
	}
	writeJSON(w, status, value)
}

func respondFound(w http.ResponseWriter, value any, found bool, err error) {
	if err != nil {
		writeCRUDError(w, err)
		return
	}
	if !found {
		writeCRUDError(w, domain.ErrNotFound)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func respondDeleted(w http.ResponseWriter, err error) {
	if err != nil {
		writeCRUDError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
