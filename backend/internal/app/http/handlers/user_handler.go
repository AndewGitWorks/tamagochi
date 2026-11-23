package http

import (
	"net/http"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/security"
)

type UserHandler struct {
	Users    UserCRUDStore
	Profiles users.ProfileStore
	auth     *users.AuthService
}

func NewUserHandler(usersStore UserCRUDStore, profiles users.ProfileStore, auth *users.AuthService) *UserHandler {
	return &UserHandler{Users: usersStore, Profiles: profiles, auth: auth}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, currentUser(r))
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.Users.List(r.Context(), limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var input users.RegisterInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := users.ValidateUpdate(input); err != nil {
		writeCRUDError(w, err)
		return
	}
	user := currentUser(r)
	user.Name, user.Surname, user.Login = input.Name, input.Surname, input.Login
	user.Email, user.PhoneNumber = input.Email, input.PhoneNumber
	if input.Password != "" {
		hash, err := security.HashPassword(input.Password)
		if err != nil {
			writeCRUDError(w, err)
			return
		}
		user.PasswordHash = hash
	}
	updated, err := h.Users.Update(r.Context(), user)
	respond(w, http.StatusOK, updated, err)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	respondDeleted(w, h.Users.Delete(r.Context(), currentUser(r).ID))
}

type profileInput struct {
	Bio     string `json:"bio"`
	Address string `json:"address"`
}

func (h *UserHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var input profileInput
	if !decodeJSON(w, r, &input) {
		return
	}
	profile, err := h.Profiles.Create(r.Context(), users.Profile{UserID: currentUser(r).ID, Bio: input.Bio, Address: input.Address})
	respond(w, http.StatusCreated, profile, err)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	profile, found, err := h.Profiles.FindByUserID(r.Context(), currentUser(r).ID)
	respondFound(w, profile, found, err)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var input profileInput
	if !decodeJSON(w, r, &input) {
		return
	}
	profile, err := h.Profiles.Update(r.Context(), users.Profile{UserID: currentUser(r).ID, Bio: input.Bio, Address: input.Address})
	respond(w, http.StatusOK, profile, err)
}

func (h *UserHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	respondDeleted(w, h.Profiles.Delete(r.Context(), currentUser(r).ID))
}
