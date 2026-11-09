package http

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/security"
)

const maxRequestBodySize = 16 << 10

type AuthHandler struct {
	auth *users.AuthService
}

func NewAuthHandler(auth *users.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /register", h.register)
	mux.HandleFunc("POST /login", h.login)
	mux.HandleFunc("POST /refresh", h.refresh)
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var input users.RegisterInput
	if !decodeJSON(w, r, &input) {
		return
	}
	pair, err := h.auth.Register(r.Context(), input)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, pair)
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var input users.LoginInput
	if !decodeJSON(w, r, &input) {
		return
	}
	pair, err := h.auth.Login(r.Context(), input)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeAuthError(w, security.ErrInvalidToken)
		return
	}
	pair, err := h.auth.Refresh(r.Context(), token)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "request body must be valid JSON"})
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": "request body must contain one JSON value"})
		return false
	}
	return true
}

func bearerToken(value string) (string, bool) {
	parts := strings.Fields(value)
	returnToken := len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] != ""
	if !returnToken {
		return "", false
	}
	return parts[1], true
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, users.ErrValidation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "message": err.Error()})
	case errors.Is(err, users.ErrLoginTaken), errors.Is(err, users.ErrUserConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "login_taken", "message": err.Error()})
	case errors.Is(err, users.ErrInvalidCredentials), errors.Is(err, security.ErrInvalidToken):
		w.Header().Set("WWW-Authenticate", `Bearer realm="tamagochi-api"`)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication_error", "message": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "an unexpected error occurred"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
