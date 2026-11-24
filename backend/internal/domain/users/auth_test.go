package users

import (
	"context"
	"errors"
	"testing"

	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/security"
)

const testSecret = "a-long-test-secret-that-is-at-least-32-characters"

func newTestService() *AuthService {
	return NewAuthService(AuthConfig{
		Store: NewMemoryStore(), JWTSecret: []byte(testSecret), Issuer: "test-api",
	})
}

func TestRegisterIssuesRequiredClaims(t *testing.T) {
	service := newTestService()
	pair, err := service.Register(context.Background(), RegisterInput{Login: "maria", Name: "Maria", Surname: "Rossi", Email: "maria@example.com", PhoneNumber: "+15551234567", Password: "correct-horse-battery-staple"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := security.VerifyJWT(pair.AccessToken, []byte(testSecret), "test-api", "access", service.now())
	if err != nil {
		t.Fatal(err)
	}
	if pair.TokenType != "Bearer" || claims.Login != "maria" || claims.Name != "Maria" || claims.Role != "USER" || claims.TimeCreation == "" || claims.TimeExpiration == "" {
		t.Fatalf("unexpected access claims: %#v", claims)
	}
}

func TestLoginAndRefreshRotation(t *testing.T) {
	service := newTestService()
	if _, err := service.Register(context.Background(), RegisterInput{Login: "maria", Name: "Maria", Surname: "Rossi", Email: "maria@example.com", PhoneNumber: "+15551234567", Password: "correct-horse-battery-staple"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(context.Background(), LoginInput{Login: "maria", Password: "not-the-password"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	login, err := service.Login(context.Background(), LoginInput{Login: "maria", Password: "correct-horse-battery-staple"})
	if err != nil {
		t.Fatal(err)
	}
	refreshed, err := service.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.RefreshToken == login.RefreshToken {
		t.Fatal("refresh token was not rotated")
	}
	if _, err := service.Refresh(context.Background(), login.RefreshToken); !errors.Is(err, security.ErrInvalidToken) {
		t.Fatalf("expected reused refresh token to be rejected, got %v", err)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	valid := RegisterInput{Login: "maria", Name: "Maria", Surname: "Rossi", Email: "maria@example.com", PhoneNumber: "+15551234567", Password: "correct-horse-battery-staple"}
	cases := []struct {
		name   string
		mutate func(*RegisterInput)
	}{
		{"short login", func(v *RegisterInput) { v.Login = "ab" }},
		{"short name", func(v *RegisterInput) { v.Name = "Al" }},
		{"short surname", func(v *RegisterInput) { v.Surname = "Li" }},
		{"invalid email", func(v *RegisterInput) { v.Email = "not-an-email" }},
		{"invalid phone", func(v *RegisterInput) { v.PhoneNumber = "12345" }},
		{"short password", func(v *RegisterInput) { v.Password = "short" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.mutate(&input)
			service := newTestService()
			if _, err := service.Register(context.Background(), input); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected ErrValidation, got %v", err)
			}
			if _, found, err := service.store.FindByLogin(context.Background(), input.Login); err != nil || found {
				t.Fatalf("invalid user was persisted: found=%v err=%v", found, err)
			}
		})
	}
}
