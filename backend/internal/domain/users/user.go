package users

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
)

var ErrLoginTaken = errors.New("a user with this login already exists")

// User is deliberately independent of any database model.
type User struct {
	ID           string
	Login        string
	Name         string
	Role         string
	PasswordHash string
}

// Store is the boundary to replace with a database repository later.
type Store interface {
	Create(ctx context.Context, user User) (User, error)
	FindByLogin(ctx context.Context, login string) (User, bool, error)
}

// MemoryStore is temporary persistence for local development before DB models exist.
type MemoryStore struct {
	mu      sync.RWMutex
	byLogin map[string]User
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{byLogin: make(map[string]User)}
}

func (s *MemoryStore) Create(_ context.Context, user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byLogin[user.Login]; exists {
		return User{}, ErrLoginTaken
	}
	user.ID = newID()
	s.byLogin[user.Login] = user
	return user, nil
}

func (s *MemoryStore) FindByLogin(_ context.Context, login string) (User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.byLogin[login]
	return user, ok, nil
}

func newID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}
