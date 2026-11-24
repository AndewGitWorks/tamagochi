package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AndewGitWorks/tamagochi/backend/internal/app/http/routes"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/security"
)

type productSpy struct {
	products.Store
	created bool
}

type productsImageStub struct{ products.ImageStore }

type userCRUDStub struct{ users.Store }

func (userCRUDStub) List(context.Context, int, int) ([]users.User, error) { return nil, nil }

type profileStub struct{ users.ProfileStore }

func (s *productSpy) Create(_ context.Context, product products.Product) (products.Product, error) {
	s.created = true
	product.ID = 1
	return product, nil
}

type orderStub struct{ orders.Store }

func (s orderStub) FindByID(context.Context, int64) (orders.Order, bool, error) {
	return orders.Order{ID: 1, UserID: 999, Status: "PENDING"}, true, nil
}

func crudTestAuth(t *testing.T, role string) (*users.AuthService, string, string) {
	t.Helper()
	store := users.NewMemoryStore()
	user, err := store.Create(context.Background(), users.User{Name: "Test", Surname: "User", Login: "tester", Email: "test@example.test", Role: role, PasswordHash: "never expose this"})
	if err != nil {
		t.Fatal(err)
	}
	secret := []byte("a-test-secret-with-at-least-32-characters")
	auth := users.NewAuthService(users.AuthConfig{Store: store, JWTSecret: secret, Issuer: "test"})
	sign := func(kind string) string {
		token, err := security.SignJWT(security.Claims{Issuer: "test", Subject: strconv.FormatInt(user.ID, 10), Login: user.Login, Name: user.Name, Role: role,
			IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Hour).Unix(), ID: "token-id", TokenType: kind}, secret)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	return auth, sign("access"), sign("refresh")
}

func TestCatalogWritesRequireAdminAccessToken(t *testing.T) {
	for _, tc := range []struct {
		name, role, tokenKind string
		status                int
	}{
		{"anonymous", "USER", "none", http.StatusUnauthorized},
		{"user", "USER", "access", http.StatusForbidden},
		{"refresh token", "ADMIN", "refresh", http.StatusUnauthorized},
		{"admin", "ADMIN", "access", http.StatusCreated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auth, access, refresh := crudTestAuth(t, tc.role)
			spy := &productSpy{}
			router := routes.NewRouter(auth, routes.Config{CatalogService: products.NewService(spy, productsImageStub{})})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(`{"title":"Товар","price":"19.99"}`))
			if tc.tokenKind == "access" {
				req.Header.Set("Authorization", "Bearer "+access)
			}
			if tc.tokenKind == "refresh" {
				req.Header.Set("Authorization", "Bearer "+refresh)
			}
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
			}
			if spy.created != (tc.status == http.StatusCreated) {
				t.Fatal("unexpected store call")
			}
		})
	}
}

func TestCannotReadAnotherUsersOrderOrItems(t *testing.T) {
	auth, token, _ := crudTestAuth(t, "USER")
	router := routes.NewRouter(auth, routes.Config{OrderService: orders.NewService(orderStub{}, nil)})
	for _, path := range []string{"/api/v1/orders/1", "/api/v1/orders/1/items", "/api/v1/orders/1/items/2"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != http.StatusNotFound {
			t.Fatalf("path=%s status=%d", path, res.Code)
		}
	}
}

func TestUserResponseOmitsPasswordHash(t *testing.T) {
	auth, token, _ := crudTestAuth(t, "USER")
	router := routes.NewRouter(auth, routes.Config{Users: userCRUDStub{}, UserService: users.NewUserService(userCRUDStub{}), ProfileService: users.NewProfileService(profileStub{})})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatal(res.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 || body["login"] != "tester" {
		t.Fatalf("body=%s", res.Body.String())
	}
	if strings.Contains(res.Body.String(), "never expose this") {
		t.Fatal("response leaked password hash")
	}
}

func TestProductBodyRejectsImmutableFieldsAndTrailingJSON(t *testing.T) {
	auth, token, _ := crudTestAuth(t, "ADMIN")
	for _, body := range []string{
		`{"id":123,"title":"Product","price":"1.00"}`,
		`{"title":"Product","price":"1.00"} {}`,
	} {
		spy := &productSpy{}
		router := routes.NewRouter(auth, routes.Config{CatalogService: products.NewService(spy, productsImageStub{})})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest || spy.created {
			t.Fatalf("status=%d created=%v", res.Code, spy.created)
		}
	}
}
