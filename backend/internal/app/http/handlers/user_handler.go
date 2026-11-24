package http

import (
	"net/http"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

type UserHandler struct {
	Users    UserCRUDStore
	users    *users.UserService
	profiles *users.ProfileService
}

func NewUserHandler(usersStore UserCRUDStore, userService *users.UserService, profileService *users.ProfileService) *UserHandler {
	return &UserHandler{Users: usersStore, users: userService, profiles: profileService}
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
	updated, err := h.users.UpdateSelf(r.Context(), currentUser(r), input)
	respond(w, http.StatusOK, updated, err)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	respondDeleted(w, h.users.DeleteSelf(r.Context(), currentUser(r).ID))
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
	profile, err := h.profiles.Save(r.Context(), users.Profile{UserID: currentUser(r).ID, Bio: input.Bio, Address: input.Address})
	respond(w, http.StatusOK, profile, err)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.profiles.Get(r.Context(), currentUser(r).ID)
	respond(w, http.StatusOK, profile, err)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var input profileInput
	if !decodeJSON(w, r, &input) {
		return
	}
	profile, err := h.profiles.Save(r.Context(), users.Profile{UserID: currentUser(r).ID, Bio: input.Bio, Address: input.Address})
	respond(w, http.StatusOK, profile, err)
}

func (h *UserHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	respondDeleted(w, h.profiles.Clear(r.Context(), currentUser(r).ID))
}
