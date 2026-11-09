package http

import (
	"net/http"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

// RouterConfig contains feature dependencies assembled by cmd/api. Individual
// handlers own their routes; this router only composes them.
type RouterConfig struct {
	Users    UserCRUDStore
	Profiles users.ProfileStore
	Products products.Store
	Images   products.ImageStore
	Orders   orders.Store
	Items    orders.ItemStore
	Likes    products.LikedProductStore
}

type routeRegistrar interface {
	RegisterRoutes(*http.ServeMux)
}

func NewRouter(auth *users.AuthService, config ...RouterConfig) http.Handler {
	mux := http.NewServeMux()
	registrars := []routeRegistrar{NewAuthHandler(auth)}
	if len(config) > 0 {
		dependencies := config[0]
		registrars = append(registrars,
			NewProductHandler(dependencies.Products, dependencies.Images, auth),
			NewUserHandler(dependencies.Users, dependencies.Profiles, auth),
			NewOrderHandler(dependencies.Orders, dependencies.Items, auth),
			NewLikedProductHandler(dependencies.Likes, auth),
		)
	}
	for _, registrar := range registrars {
		registrar.RegisterRoutes(mux)
	}
	return cors(mux)
}
