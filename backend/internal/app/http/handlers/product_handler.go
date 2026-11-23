package http

import (
	"encoding/json"
	"net/http"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

type productInput struct {
	Title          string          `json:"title"`
	Price          string          `json:"price"`
	SeriaTitle     string          `json:"seria_title"`
	Bio            string          `json:"bio"`
	Description    string          `json:"description"`
	Specifications json.RawMessage `json:"specifications"`
}

type ProductHandler struct {
	Products products.Store
	Images   products.ImageStore
	auth     *users.AuthService
}

func NewProductHandler(productsStore products.Store, images products.ImageStore, auth *users.AuthService) *ProductHandler {
	return &ProductHandler{Products: productsStore, Images: images, auth: auth}
}

func (input productInput) product(id int64) products.Product {
	return products.Product{ID: id, Title: input.Title, Price: input.Price, SeriaTitle: input.SeriaTitle,
		Bio: input.Bio, Description: input.Description, Specifications: input.Specifications}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var input productInput
	if !decodeJSON(w, r, &input) {
		return
	}
	product, err := h.Products.Create(r.Context(), input.product(0))
	respond(w, http.StatusCreated, product, err)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var input productInput
	if !decodeJSON(w, r, &input) {
		return
	}
	product, err := h.Products.Update(r.Context(), input.product(id))
	respond(w, http.StatusOK, product, err)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	product, found, err := h.Products.FindByID(r.Context(), id)
	respondFound(w, product, found, err)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.Products.List(r.Context(), limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	respondDeleted(w, h.Products.Delete(r.Context(), id))
}

func (h *ProductHandler) imageProduct(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := pathID(w, r, "productID")
	if !ok {
		return 0, false
	}
	_, found, err := h.Products.FindByID(r.Context(), id)
	if err != nil {
		writeCRUDError(w, err)
		return 0, false
	}
	if !found {
		writeCRUDError(w, domain.ErrNotFound)
		return 0, false
	}
	return id, true
}

type imageInput struct {
	StorageKey string `json:"storage_key"`
	SortOrder  int    `json:"sort_order"`
}

func (h *ProductHandler) CreateImage(w http.ResponseWriter, r *http.Request) {
	productID, ok := h.imageProduct(w, r)
	if !ok {
		return
	}
	var input imageInput
	if !decodeJSON(w, r, &input) {
		return
	}
	image, err := h.Images.Create(r.Context(), products.Image{ProductID: productID, StorageKey: input.StorageKey, SortOrder: input.SortOrder})
	respond(w, http.StatusCreated, image, err)
}

func (h *ProductHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	productID, ok := h.imageProduct(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var input imageInput
	if !decodeJSON(w, r, &input) {
		return
	}
	image, err := h.Images.Update(r.Context(), products.Image{ID: id, ProductID: productID, StorageKey: input.StorageKey, SortOrder: input.SortOrder})
	respond(w, http.StatusOK, image, err)
}

func (h *ProductHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	productID, ok := h.imageProduct(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	image, found, err := h.Images.FindByID(r.Context(), id, productID)
	respondFound(w, image, found, err)
}

func (h *ProductHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	productID, ok := h.imageProduct(w, r)
	if !ok {
		return
	}
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.Images.ListByProduct(r.Context(), productID, limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *ProductHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	productID, ok := h.imageProduct(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	respondDeleted(w, h.Images.Delete(r.Context(), id, productID))
}
