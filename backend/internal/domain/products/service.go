package products

import (
	"context"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
)

// Service owns catalog rules; HTTP handlers only decode and encode requests.
type Service struct {
	products Store
	images   ImageStore
}

func NewService(products Store, images ImageStore) *Service {
	return &Service{products: products, images: images}
}

func (s *Service) Create(ctx context.Context, product Product) (Product, error) {
	return s.products.Create(ctx, product)
}
func (s *Service) Update(ctx context.Context, product Product) (Product, error) {
	if product.ID <= 0 {
		return Product{}, domain.ErrValidation
	}
	return s.products.Update(ctx, product)
}
func (s *Service) Get(ctx context.Context, id int64) (Product, error) {
	if id <= 0 {
		return Product{}, domain.ErrValidation
	}
	p, found, err := s.products.FindByID(ctx, id)
	if err != nil || !found {
		if err != nil {
			return Product{}, err
		}
		return Product{}, domain.ErrNotFound
	}
	return p, nil
}
func (s *Service) List(ctx context.Context, limit, offset int) ([]Product, error) {
	return s.products.List(ctx, limit, offset)
}
func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrValidation
	}
	return s.products.Delete(ctx, id)
}

func (s *Service) CreateImage(ctx context.Context, image Image) (Image, error) {
	if _, err := s.Get(ctx, image.ProductID); err != nil {
		return Image{}, err
	}
	return s.images.Create(ctx, image)
}
func (s *Service) UpdateImage(ctx context.Context, image Image) (Image, error) {
	if _, err := s.Get(ctx, image.ProductID); err != nil {
		return Image{}, err
	}
	return s.images.Update(ctx, image)
}
func (s *Service) GetImage(ctx context.Context, productID, id int64) (Image, error) {
	if _, err := s.Get(ctx, productID); err != nil {
		return Image{}, err
	}
	image, found, err := s.images.FindByID(ctx, id, productID)
	if err != nil || !found {
		if err != nil {
			return Image{}, err
		}
		return Image{}, domain.ErrNotFound
	}
	return image, nil
}
func (s *Service) ListImages(ctx context.Context, productID int64, limit, offset int) ([]Image, error) {
	if _, err := s.Get(ctx, productID); err != nil {
		return nil, err
	}
	return s.images.ListByProduct(ctx, productID, limit, offset)
}
func (s *Service) DeleteImage(ctx context.Context, productID, id int64) error {
	if _, err := s.Get(ctx, productID); err != nil {
		return err
	}
	return s.images.Delete(ctx, id, productID)
}

// FavoriteService makes adding and removing favorites idempotent.
type FavoriteService struct {
	products Store
	likes    LikedProductStore
}

func NewFavoriteService(products Store, likes LikedProductStore) *FavoriteService {
	return &FavoriteService{products: products, likes: likes}
}
func (s *FavoriteService) Add(ctx context.Context, userID, productID int64) (LikedProduct, error) {
	if userID <= 0 || productID <= 0 {
		return LikedProduct{}, domain.ErrValidation
	}
	if _, found, err := s.products.FindByID(ctx, productID); err != nil {
		return LikedProduct{}, err
	} else if !found {
		return LikedProduct{}, domain.ErrNotFound
	}
	if like, found, err := s.likes.Find(ctx, userID, productID); err != nil {
		return LikedProduct{}, err
	} else if found {
		return like, nil
	}
	return s.likes.Create(ctx, LikedProduct{UserID: userID, ProductID: productID})
}
func (s *FavoriteService) Get(ctx context.Context, userID, productID int64) (LikedProduct, error) {
	like, found, err := s.likes.Find(ctx, userID, productID)
	if err != nil || !found {
		if err != nil {
			return LikedProduct{}, err
		}
		return LikedProduct{}, domain.ErrNotFound
	}
	return like, nil
}
func (s *FavoriteService) List(ctx context.Context, userID int64, limit, offset int) ([]LikedProduct, error) {
	if userID <= 0 {
		return nil, domain.ErrValidation
	}
	return s.likes.ListByUser(ctx, userID, limit, offset)
}
func (s *FavoriteService) Remove(ctx context.Context, userID, productID int64) error {
	if userID <= 0 || productID <= 0 {
		return domain.ErrValidation
	}
	_, found, err := s.likes.Find(ctx, userID, productID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return s.likes.Delete(ctx, userID, productID)
}
