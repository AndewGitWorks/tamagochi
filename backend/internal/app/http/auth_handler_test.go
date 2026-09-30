package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

func TestAuthRoutes(t *testing.T) {
	service := users.NewAuthService(users.AuthConfig{
		Store: users.NewMemoryStore(), JWTSecret: []byte("another-long-test-secret-that-is-at-least-32-characters"), Issuer: "test-api",
	})
	router := NewRouter(service)

	register := request(router, http.MethodPost, "/register", `{"login":"sasha","name":"Sasha","surname":"Rossi","email":"sasha@example.com","phone_number":"+15551234567","password":"correct-horse-battery-staple"}`, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register status: %d; body: %s", register.Code, register.Body.String())
	}

	login := request(router, http.MethodPost, "/login", `{"login":"sasha","password":"correct-horse-battery-staple"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status: %d; body: %s", login.Code, login.Body.String())
	}
	var tokens users.TokenPair
	if err := json.Unmarshal(login.Body.Bytes(), &tokens); err != nil {
		t.Fatal(err)
	}

	refresh := request(router, http.MethodPost, "/refresh", "", "Bearer "+tokens.RefreshToken)
	if refresh.Code != http.StatusOK {
		t.Fatalf("refresh status: %d; body: %s", refresh.Code, refresh.Body.String())
	}
}

func request(router http.Handler, method, path, body, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func TestRegisterRejectsInvalidRequest(t *testing.T) {
	service := users.NewAuthService(users.AuthConfig{Store: users.NewMemoryStore(), JWTSecret: []byte("another-long-test-secret-that-is-at-least-32-characters"), Issuer: "test-api"})
	router := NewRouter(service)
	response := request(router, http.MethodPost, "/register", `{"login":"ab","name":"A","surname":"B","email":"bad","phone_number":"123","password":"short"}`, "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status: %d; body: %s", response.Code, response.Body.String())
	}
}
