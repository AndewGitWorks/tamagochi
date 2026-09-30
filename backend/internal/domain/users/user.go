package users

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrLoginTaken   = errors.New("a user with this login already exists")
	ErrUserConflict = errors.New("a user with these unique fields already exists")
	ErrUserNotFound = errors.New("user was not found")
)

type User struct {
	ID           int64
	Name         string
	Surname      string
	Login        string
	Email        string
	PasswordHash string
	PhoneNumber  string
	Version      int64
	Role         string
}

type Profile struct {
	ID          int64
	UserID      int64
	Bio         string
	Address     string
	OrdersCount int64
}

// Store is the persistence boundary for user CRUD. The PostgreSQL implementation
// also creates the initial profile atomically; MemoryStore models users for unit tests.
type Store interface {
	Create(ctx context.Context, user User) (User, error)
	FindByLogin(ctx context.Context, login string) (User, bool, error)
	FindByID(ctx context.Context, id int64) (User, bool, error)
	Update(ctx context.Context, user User) (User, error)
	Delete(ctx context.Context, id int64) error
}

// MemoryStore is used by isolated service tests and local development.
type MemoryStore struct {
	mu      sync.RWMutex
	nextID  int64
	byID    map[int64]User
	byLogin map[string]int64
	byEmail map[string]int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{byID: make(map[int64]User), byLogin: make(map[string]int64), byEmail: make(map[string]int64)}
}

func (s *MemoryStore) Create(_ context.Context, user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byLogin[user.Login]; exists {
		return User{}, ErrLoginTaken
	}
	if _, exists := s.byEmail[user.Email]; exists {
		return User{}, ErrUserConflict
	}
	s.nextID++
	user.ID, user.Version = s.nextID, 1
	if user.Role == "" {
		user.Role = "USER"
	}
	s.byID[user.ID], s.byLogin[user.Login], s.byEmail[user.Email] = user, user.ID, user.ID
	return user, nil
}

func (s *MemoryStore) FindByLogin(_ context.Context, login string) (User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byLogin[login]
	if !ok {
		return User{}, false, nil
	}
	return s.byID[id], true, nil
}

func (s *MemoryStore) FindByID(_ context.Context, id int64) (User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.byID[id]
	return user, ok, nil
}

func (s *MemoryStore) Update(_ context.Context, user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.byID[user.ID]
	if !ok {
		return User{}, ErrUserNotFound
	}
	if id, exists := s.byLogin[user.Login]; exists && id != user.ID {
		return User{}, ErrLoginTaken
	}
	if id, exists := s.byEmail[user.Email]; exists && id != user.ID {
		return User{}, ErrUserConflict
	}
	delete(s.byLogin, old.Login)
	delete(s.byEmail, old.Email)
	user.Version = old.Version + 1
	if user.Role == "" {
		user.Role = old.Role
	}
	s.byID[user.ID], s.byLogin[user.Login], s.byEmail[user.Email] = user, user.ID, user.ID
	return user, nil
}

func (s *MemoryStore) Delete(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.byID[id]
	if !ok {
		return ErrUserNotFound
	}
	delete(s.byID, id)
	delete(s.byLogin, user.Login)
	delete(s.byEmail, user.Email)
	return nil
}
