package users

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/security"
)

type UserService struct{ store Store }

func NewUserService(store Store) *UserService { return &UserService{store: store} }
func (s *UserService) UpdateSelf(ctx context.Context, current User, input RegisterInput) (User, error) {
	if current.ID <= 0 {
		return User{}, domain.ErrValidation
	}
	if err := ValidateUpdate(input); err != nil {
		return User{}, err
	}
	current.Name, current.Surname, current.Login, current.Email, current.PhoneNumber = input.Name, input.Surname, input.Login, input.Email, input.PhoneNumber
	if input.Password != "" {
		hash, err := security.HashPassword(input.Password)
		if err != nil {
			return User{}, err
		}
		current.PasswordHash = hash
	}
	return s.store.Update(ctx, current)
}
func (s *UserService) DeleteSelf(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return domain.ErrValidation
	}
	return s.store.Delete(ctx, userID)
}

type ProfileService struct{ store ProfileStore }

func NewProfileService(store ProfileStore) *ProfileService { return &ProfileService{store: store} }
func validateProfile(p Profile) error {
	if p.UserID <= 0 || utf8.RuneCountInString(p.Bio) > 1000 || utf8.RuneCountInString(p.Address) > 500 {
		return domain.ErrValidation
	}
	return nil
}
func (s *ProfileService) Get(ctx context.Context, userID int64) (Profile, error) {
	p, found, err := s.store.FindByUserID(ctx, userID)
	if err != nil || !found {
		if err != nil {
			return Profile{}, err
		}
		return Profile{}, domain.ErrNotFound
	}
	return p, nil
}
func (s *ProfileService) Save(ctx context.Context, profile Profile) (Profile, error) {
	profile.Bio, profile.Address = strings.TrimSpace(profile.Bio), strings.TrimSpace(profile.Address)
	if err := validateProfile(profile); err != nil {
		return Profile{}, err
	}
	existing, found, err := s.store.FindByUserID(ctx, profile.UserID)
	if err != nil {
		return Profile{}, err
	}
	if found {
		profile.ID, profile.OrdersCount = existing.ID, existing.OrdersCount
		return s.store.Update(ctx, profile)
	}
	return s.store.Create(ctx, profile)
}
func (s *ProfileService) Clear(ctx context.Context, userID int64) error {
	_, err := s.Save(ctx, Profile{UserID: userID})
	return err
}
