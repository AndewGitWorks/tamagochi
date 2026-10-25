package users

import "context"

type ProfileStore interface {
	Create(context.Context, Profile) (Profile, error)
	FindByUserID(context.Context, int64) (Profile, bool, error)
	Update(context.Context, Profile) (Profile, error)
	Delete(context.Context, int64) error
}

// Update keeps the same validation rules as registration. Password changes
// are optional; the HTTP handler hashes a new password before saving it.
func ValidateUpdate(input RegisterInput) error {
	if input.Password == "" {
		input.Password = "unchanged-password"
	}
	return validateRegistration(input)
}
