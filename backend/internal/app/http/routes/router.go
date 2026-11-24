// Package routes owns the public HTTP contract of the API. Handlers do not
// register themselves: every URL and middleware assignment is visible here.
package routes

import (
	"log"
	"net/http"

	apphttp "github.com/AndewGitWorks/tamagochi/backend/internal/app/http/handlers"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

const APIV1Prefix = "/api/v1"

// Config supplies the dependencies for optional API modules. Authentication is
// always available; a module is registered only when all of its dependencies
// have been assembled by cmd/api.
type Config struct {
	Users           apphttp.UserCRUDStore
	UserService     *users.UserService
	ProfileService  *users.ProfileService
	CatalogService  *products.Service
	OrderService    *orders.Service
	FavoriteService *products.FavoriteService
}

func NewRouter(auth *users.AuthService, configs ...Config) http.Handler {
	api := http.NewServeMux()
	authHandler := apphttp.NewAuthHandler(auth)
	api.HandleFunc("POST /register", authHandler.Register)
	api.HandleFunc("POST /login", authHandler.Login)
	api.HandleFunc("POST /refresh", authHandler.Refresh)
	api.HandleFunc("GET /health", authHandler.HealthCheck)
	if len(configs) > 0 {
		registerResources(api, auth, configs[0])
	}

	root := http.NewServeMux()
	root.Handle(APIV1Prefix+"/", http.StripPrefix(APIV1Prefix, api))
	return cors(root)
}

func registerResources(api *http.ServeMux, auth *users.AuthService, config Config) {
	if config.CatalogService != nil {
		productsHandler := apphttp.NewProductHandler(config.CatalogService)
		api.HandleFunc("GET /products", productsHandler.ListProducts)
		api.HandleFunc("GET /products/{id}", productsHandler.GetProduct)
		api.HandleFunc("POST /products", apphttp.RequireUser(auth, true, productsHandler.CreateProduct))
		api.HandleFunc("PUT /products/{id}", apphttp.RequireUser(auth, true, productsHandler.UpdateProduct))
		api.HandleFunc("DELETE /products/{id}", apphttp.RequireUser(auth, true, productsHandler.DeleteProduct))
		api.HandleFunc("GET /products/{productID}/images", productsHandler.ListImages)
		api.HandleFunc("GET /products/{productID}/images/{id}", productsHandler.GetImage)
		api.HandleFunc("POST /products/{productID}/images", apphttp.RequireUser(auth, true, productsHandler.CreateImage))
		api.HandleFunc("PUT /products/{productID}/images/{id}", apphttp.RequireUser(auth, true, productsHandler.UpdateImage))
		api.HandleFunc("DELETE /products/{productID}/images/{id}", apphttp.RequireUser(auth, true, productsHandler.DeleteImage))
	}

	if config.Users != nil && config.UserService != nil && config.ProfileService != nil {
		usersHandler := apphttp.NewUserHandler(config.Users, config.UserService, config.ProfileService)
		api.HandleFunc("GET /users", apphttp.RequireUser(auth, true, usersHandler.ListUsers))
		api.HandleFunc("GET /users/me", apphttp.RequireUser(auth, false, usersHandler.GetUser))
		api.HandleFunc("PUT /users/me", apphttp.RequireUser(auth, false, usersHandler.UpdateUser))
		api.HandleFunc("DELETE /users/me", apphttp.RequireUser(auth, false, usersHandler.DeleteUser))
		api.HandleFunc("POST /profiles/me", apphttp.RequireUser(auth, false, usersHandler.CreateProfile))
		api.HandleFunc("GET /profiles/me", apphttp.RequireUser(auth, false, usersHandler.GetProfile))
		api.HandleFunc("PUT /profiles/me", apphttp.RequireUser(auth, false, usersHandler.UpdateProfile))
		api.HandleFunc("DELETE /profiles/me", apphttp.RequireUser(auth, false, usersHandler.DeleteProfile))
	}

	if config.OrderService != nil {
		cartHandler := apphttp.NewCartHandler(config.OrderService)
		api.HandleFunc("GET /cart", apphttp.RequireUser(auth, false, cartHandler.GetCart))
		api.HandleFunc("POST /cart/items", apphttp.RequireUser(auth, false, cartHandler.AddItem))
		api.HandleFunc("PUT /cart/items/{id}", apphttp.RequireUser(auth, false, cartHandler.UpdateItem))
		api.HandleFunc("DELETE /cart/items/{id}", apphttp.RequireUser(auth, false, cartHandler.RemoveItem))
		api.HandleFunc("POST /cart/checkout", apphttp.RequireUser(auth, false, cartHandler.Checkout))

		ordersHandler := apphttp.NewOrderHandler(config.OrderService, auth)
		api.HandleFunc("POST /orders", apphttp.RequireUser(auth, false, ordersHandler.CreateOrder))
		api.HandleFunc("GET /orders", apphttp.RequireUser(auth, false, ordersHandler.ListOrders))
		api.HandleFunc("GET /orders/{orderID}", apphttp.RequireUser(auth, false, ordersHandler.GetOrder))
		api.HandleFunc("PUT /orders/{orderID}", apphttp.RequireUser(auth, false, ordersHandler.UpdateOrder))
		api.HandleFunc("POST /orders/{orderID}/payment", apphttp.RequireUser(auth, false, ordersHandler.PayOrder))
		api.HandleFunc("DELETE /orders/{orderID}", apphttp.RequireUser(auth, false, ordersHandler.DeleteOrder))
		api.HandleFunc("POST /orders/{orderID}/items", apphttp.RequireUser(auth, false, ordersHandler.CreateItem))
		api.HandleFunc("GET /orders/{orderID}/items", apphttp.RequireUser(auth, false, ordersHandler.ListItems))
		api.HandleFunc("GET /orders/{orderID}/items/{id}", apphttp.RequireUser(auth, false, ordersHandler.GetItem))
		api.HandleFunc("PUT /orders/{orderID}/items/{id}", apphttp.RequireUser(auth, false, ordersHandler.UpdateItem))
		api.HandleFunc("DELETE /orders/{orderID}/items/{id}", apphttp.RequireUser(auth, false, ordersHandler.DeleteItem))
	}

	if config.FavoriteService != nil {
		likesHandler := apphttp.NewLikedProductHandler(config.FavoriteService)
		api.HandleFunc("POST /likes/{productID}", apphttp.RequireUser(auth, false, likesHandler.CreateLike))
		api.HandleFunc("GET /likes", apphttp.RequireUser(auth, false, likesHandler.ListLikes))
		api.HandleFunc("GET /likes/{productID}", apphttp.RequireUser(auth, false, likesHandler.GetLike))
		api.HandleFunc("DELETE /likes/{productID}", apphttp.RequireUser(auth, false, likesHandler.DeleteLike))
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
