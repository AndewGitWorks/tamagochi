package users

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"time"

	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/security"
)

var (
	ErrValidation         = errors.New("invalid authentication request")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

var (
	loginPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,24}$`)
	phonePattern = regexp.MustCompile(`^\+[0-9]{9,14}$`)
)

type AuthConfig struct {
	Store           Store
	JWTSecret       []byte
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	RefreshSessions RefreshSessionStore
}

type AuthService struct {
	store           Store
	jwtSecret       []byte
	issuer          string
	accessTTL       time.Duration
	refreshTTL      time.Duration
	now             func() time.Time
	refreshSessions RefreshSessionStore
}

type RegisterInput struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type LoginInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	AccessExpiresAt  string `json:"access_expires_at"`
	RefreshExpiresAt string `json:"refresh_expires_at"`
}

func NewAuthService(config AuthConfig) *AuthService {
	if config.AccessTokenTTL <= 0 {
		config.AccessTokenTTL = 15 * time.Minute
	}
	if config.RefreshTokenTTL <= 0 {
		config.RefreshTokenTTL = 30 * 24 * time.Hour
	}
	if config.RefreshSessions == nil {
		config.RefreshSessions = NewMemoryRefreshSessionStore()
	}
	return &AuthService{store: config.Store, jwtSecret: config.JWTSecret, issuer: config.Issuer,
		accessTTL: config.AccessTokenTTL, refreshTTL: config.RefreshTokenTTL,
		now: time.Now, refreshSessions: config.RefreshSessions}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (TokenPair, error) {
	if err := validateRegistration(input); err != nil {
		return TokenPair{}, err
	}
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return TokenPair{}, err
	}
	user, err := s.store.Create(ctx, User{Login: input.Login, Name: input.Name, Surname: input.Surname,
		Email: input.Email, PhoneNumber: input.PhoneNumber, Role: "USER", PasswordHash: hash})
	if err != nil {
		return TokenPair{}, err
	}
	return s.issuePair(ctx, user)
}

func validateRegistration(input RegisterInput) error {
	if !loginPattern.MatchString(input.Login) || len([]rune(input.Name)) < 3 || len([]rune(input.Name)) > 100 ||
		len([]rune(input.Surname)) < 3 || len([]rune(input.Surname)) > 100 || len(input.Email) < 3 || len(input.Email) > 100 || len(input.Password) < 8 ||
		len(input.PhoneNumber) < 10 || len(input.PhoneNumber) > 15 || !phonePattern.MatchString(input.PhoneNumber) {
		return fmt.Errorf("%w: login, name, surname, email, phone number or password does not meet requirements", ErrValidation)
	}
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email {
		return fmt.Errorf("%w: email is invalid", ErrValidation)
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (TokenPair, error) {
	user, found, err := s.store.FindByLogin(ctx, input.Login)
	if err != nil {
		return TokenPair{}, err
	}
	if !found || !security.VerifyPassword(input.Password, user.PasswordHash) {
		return TokenPair{}, ErrInvalidCredentials
	}
	return s.issuePair(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := security.VerifyJWT(refreshToken, s.jwtSecret, s.issuer, "refresh", s.now())
	if err != nil {
		return TokenPair{}, err
	}
	id, parseErr := strconv.ParseInt(claims.Subject, 10, 64)
	if parseErr != nil {
		return TokenPair{}, security.ErrInvalidToken
	}
	consumed, err := s.refreshSessions.Consume(ctx, claims.ID, id)
	if err != nil {
		return TokenPair{}, err
	}
	if !consumed {
		return TokenPair{}, security.ErrInvalidToken
	}
	user, found, err := s.store.FindByID(ctx, id)
	if err != nil || !found || strconv.FormatInt(user.ID, 10) != claims.Subject || user.Login != claims.Login {
		return TokenPair{}, security.ErrInvalidToken
	}
	return s.issuePair(ctx, user)
}

func (s *AuthService) issuePair(ctx context.Context, user User) (TokenPair, error) {
	now := s.now().UTC()
	access, accessExpiry, _, err := s.issueToken(user, "access", s.accessTTL, now)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, refreshExpiry, refreshID, err := s.issueToken(user, "refresh", s.refreshTTL, now)
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.refreshSessions.Create(ctx, RefreshSession{ID: refreshID, UserID: user.ID, ExpiresAt: refreshExpiry}); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", AccessExpiresAt: accessExpiry.Format(time.RFC3339), RefreshExpiresAt: refreshExpiry.Format(time.RFC3339)}, nil
}

func (s *AuthService) issueToken(user User, tokenType string, ttl time.Duration, now time.Time) (token string, expiresAt time.Time, id string, err error) {
	id, err = randomID()
	if err != nil {
		return "", time.Time{}, "", err
	}
	expiresAt = now.Add(ttl)
	claims := security.Claims{Issuer: s.issuer, Subject: strconv.FormatInt(user.ID, 10), Login: user.Login, Name: user.Name, Role: user.Role,
		TimeCreation: now.Format(time.RFC3339), TimeExpiration: expiresAt.Format(time.RFC3339),
		IssuedAt: now.Unix(), ExpiresAt: expiresAt.Unix(), ID: id, TokenType: tokenType}
	token, err = security.SignJWT(claims, s.jwtSecret)
	return token, expiresAt, id, err
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
