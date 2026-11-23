package http

import (
	"encoding/json"
	"net/http"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
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
	service *products.Service
}

func NewProductHandler(service *products.Service) *ProductHandler {
	return &ProductHandler{service: service}
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
	product, err := h.service.Create(r.Context(), input.product(0))
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
	product, err := h.service.Update(r.Context(), input.product(id))
	respond(w, http.StatusOK, product, err)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	product, err := h.service.Get(r.Context(), id)
	respond(w, http.StatusOK, product, err)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	result, err := h.service.List(r.Context(), limit, offset)
	respond(w, http.StatusOK, result, err)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	respondDeleted(w, h.service.Delete(r.Context(), id))
}

func (h *ProductHandler) imageProduct(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return pathID(w, r, "productID")
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
	image, err := h.service.CreateImage(r.Context(), products.Image{ProductID: productID, StorageKey: input.StorageKey, SortOrder: input.SortOrder})
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
	image, err := h.service.UpdateImage(r.Context(), products.Image{ID: id, ProductID: productID, StorageKey: input.StorageKey, SortOrder: input.SortOrder})
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
	image, err := h.service.GetImage(r.Context(), productID, id)
	respond(w, http.StatusOK, image, err)
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
	result, err := h.service.ListImages(r.Context(), productID, limit, offset)
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
	respondDeleted(w, h.service.DeleteImage(r.Context(), productID, id))
}
