package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

type UserCRUDStore interface {
	users.Store
	List(context.Context, int, int) ([]users.User, error)
}

type userContextKey struct{}

func currentUser(r *http.Request) users.User {
	return r.Context().Value(userContextKey{}).(users.User)
}

// RequireUser validates an access token and passes the current user to an endpoint.
// Route composition is kept in the routes package; handlers only implement endpoints.
func RequireUser(auth *users.AuthService, admin bool, next http.HandlerFunc) http.HandlerFunc {
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
