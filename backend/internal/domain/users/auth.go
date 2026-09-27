package users

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/security"
)

var (
	ErrValidation         = errors.New("invalid authentication request")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

var loginPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,24}$`)

type AuthConfig struct {
	Store           Store
	JWTSecret       []byte
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type AuthService struct {
	store           Store
	jwtSecret       []byte
	issuer          string
	accessTTL       time.Duration
	refreshTTL      time.Duration
	now             func() time.Time
	refreshSessions map[string]string
	mu              sync.Mutex
}

type RegisterInput struct {
	Login    string `json:"login"`
	Name     string `json:"name"`
	Password string `json:"password"`
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
	return &AuthService{
		store:           config.Store,
		jwtSecret:       config.JWTSecret,
		issuer:          config.Issuer,
		accessTTL:       config.AccessTokenTTL,
		refreshTTL:      config.RefreshTokenTTL,
		now:             time.Now,
		refreshSessions: make(map[string]string),
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (TokenPair, error) {
	if !loginPattern.MatchString(input.Login) || len([]rune(input.Name)) < 2 || len([]rune(input.Name)) > 100 || len(input.Password) < 8 {
		return TokenPair{}, fmt.Errorf("%w: login must be 3-24 allowed characters, name 2-100 characters, password at least 8 characters", ErrValidation)
	}
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return TokenPair{}, err
	}
	user, err := s.store.Create(ctx, User{Login: input.Login, Name: input.Name, Role: "USER", PasswordHash: hash})
	if err != nil {
		return TokenPair{}, err
	}
	return s.issuePair(user)
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (TokenPair, error) {
	user, found, err := s.store.FindByLogin(ctx, input.Login)
	if err != nil {
		return TokenPair{}, err
	}
	if !found || !security.VerifyPassword(input.Password, user.PasswordHash) {
		return TokenPair{}, ErrInvalidCredentials
	}
	return s.issuePair(user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := security.VerifyJWT(refreshToken, s.jwtSecret, s.issuer, "refresh", s.now())
	if err != nil {
		return TokenPair{}, err
	}
	s.mu.Lock()
	userID, found := s.refreshSessions[claims.ID]
	if found && userID == claims.Subject {
		delete(s.refreshSessions, claims.ID) // Refresh tokens are single-use.
	}
	s.mu.Unlock()
	if !found || userID != claims.Subject {
		return TokenPair{}, security.ErrInvalidToken
	}
	user, found, err := s.store.FindByLogin(ctx, claims.Login)
	if err != nil || !found || user.ID != claims.Subject {
		return TokenPair{}, security.ErrInvalidToken
	}
	return s.issuePair(user)
}

func (s *AuthService) issuePair(user User) (TokenPair, error) {
	now := s.now().UTC()
	access, accessExpiry, _, err := s.issueToken(user, "access", s.accessTTL, now)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, refreshExpiry, refreshID, err := s.issueToken(user, "refresh", s.refreshTTL, now)
	if err != nil {
		return TokenPair{}, err
	}
	s.mu.Lock()
	s.refreshSessions[refreshID] = user.ID
	s.mu.Unlock()
	return TokenPair{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry}, nil
}

func (s *AuthService) issueToken(user User, tokenType string, ttl time.Duration, now time.Time) (token, expiration, id string, err error) {
	id, err = randomID()
	if err != nil {
		return "", "", "", err
	}
	expiresAt := now.Add(ttl)
	claims := security.Claims{
		Issuer: s.issuer, Subject: user.ID, Login: user.Login, Name: user.Name, Role: user.Role,
		TimeCreation: now.Format(time.RFC3339), TimeExpiration: expiresAt.Format(time.RFC3339),
		IssuedAt: now.Unix(), ExpiresAt: expiresAt.Unix(), ID: id, TokenType: tokenType,
	}
	token, err = security.SignJWT(claims, s.jwtSecret)
	return token, claims.TimeExpiration, id, err
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
