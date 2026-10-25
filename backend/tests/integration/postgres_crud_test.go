package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/database"
	"github.com/AndewGitWorks/tamagochi/backend/internal/repository"
)

func TestPostgresCRUD(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	db, err := database.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	userStore := repository.NewPostgresUserStore(db)
	profileStore := repository.NewPostgresProfileStore(db)
	productStore := repository.NewPostgresProductStore(db)
	imageStore := repository.NewPostgresProductImageStore(db)
	orderStore := repository.NewPostgresOrderStore(db)
	itemStore := repository.NewPostgresOrderItemStore(db)
	likeStore := repository.NewPostgresLikedProductStore(db)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	user, err := userStore.Create(ctx, users.User{Name: "Test", Surname: "User", Login: "crud" + suffix, Email: suffix + "@example.test", PasswordHash: "test-hash", PhoneNumber: "+15551234567"})
	check(err)
	var productID int64
	// Cleanup also runs when an assertion fails midway through the test.
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM tamagochi.orders WHERE user_id=$1`, user.ID)
		_ = userStore.Delete(ctx, user.ID)
		if productID != 0 {
			_ = productStore.Delete(ctx, productID)
		}
	}()
	product, err := productStore.Create(ctx, products.Product{Title: "Product " + suffix, Price: "12.34", Specifications: json.RawMessage(`{"color":"red"}`)})
	check(err)
	productID = product.ID
	if product.SeriaTitle != "NEWBIE" || product.Price != "12.34" {
		t.Fatalf("unexpected product: %+v", product)
	}
	readProduct, found, err := productStore.FindByID(ctx, product.ID)
	check(err)
	if !found || string(readProduct.Specifications) == "null" {
		t.Fatal("product JSON was not stored")
	}
	product.Title = "Updated " + suffix
	product, err = productStore.Update(ctx, product)
	check(err)
	catalog, err := productStore.List(ctx, 100, 0)
	check(err)
	if len(catalog) == 0 {
		t.Fatal("empty catalog")
	}
	image, err := imageStore.Create(ctx, products.Image{ProductID: product.ID, StorageKey: "test/" + suffix + ".png"})
	check(err)
	image.SortOrder = 2
	image, err = imageStore.Update(ctx, image)
	check(err)
	_, found, err = imageStore.FindByID(ctx, image.ID, product.ID+1)
	check(err)
	if found {
		t.Fatal("image was accessible through another product")
	}
	images, err := imageStore.ListByProduct(ctx, product.ID, 20, 0)
	check(err)
	if len(images) != 1 || images[0].SortOrder != 2 {
		t.Fatalf("images=%+v", images)
	}
	_, found, err = imageStore.FindByID(ctx, image.ID, product.ID)
	check(err)
	if !found {
		t.Fatal("missing image")
	}
	check(imageStore.Delete(ctx, image.ID, product.ID))
	if err := imageStore.Delete(ctx, image.ID, product.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing image delete: %v", err)
	}
	// Add another image to verify product deletion cascades.
	image, err = imageStore.Create(ctx, products.Image{ProductID: product.ID, StorageKey: "test/" + suffix + "-cascade.png"})
	check(err)

	_, err = likeStore.Create(ctx, products.LikedProduct{UserID: user.ID, ProductID: product.ID})
	check(err)
	_, err = likeStore.Create(ctx, products.LikedProduct{UserID: user.ID, ProductID: product.ID})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate like: %v", err)
	}
	_, found, err = likeStore.Find(ctx, user.ID, product.ID)
	check(err)
	if !found {
		t.Fatal("missing like")
	}
	likes, err := likeStore.ListByUser(ctx, user.ID, 20, 0)
	check(err)
	if len(likes) != 1 {
		t.Fatalf("likes=%+v", likes)
	}
	check(likeStore.Delete(ctx, user.ID, product.ID))
	_, err = likeStore.Create(ctx, products.LikedProduct{UserID: user.ID, ProductID: product.ID})
	check(err)

	order, err := orderStore.Create(ctx, orders.Order{UserID: user.ID, Number: "order-" + suffix})
	check(err)
	item, err := itemStore.Create(ctx, orders.Item{OrderID: order.ID, ProductID: product.ID, Quantity: 2, Price: "0.01"})
	check(err)
	if item.Price != "12.34" {
		t.Fatalf("client price was accepted: %+v", item)
	}
	product.Price = "99.99"
	_, err = productStore.Update(ctx, product)
	check(err)
	order, found, err = orderStore.FindByID(ctx, order.ID)
	check(err)
	if !found || order.TotalPrice != "24.68" {
		t.Fatalf("incorrect total: %+v", order)
	}
	if err := productStore.Delete(ctx, product.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("ordered product must be protected: %v", err)
	}
	item.Quantity = 3
	item, err = itemStore.Update(ctx, item)
	check(err)
	if item.Price != "12.34" {
		t.Fatalf("snapshot price changed: %+v", item)
	}
	_, found, err = itemStore.FindByID(ctx, item.ID, order.ID+1)
	check(err)
	if found {
		t.Fatal("item was accessible through another order")
	}
	_, found, err = itemStore.FindByID(ctx, item.ID, order.ID)
	check(err)
	if !found {
		t.Fatal("missing item")
	}
	lines, err := itemStore.ListByOrder(ctx, order.ID, 20, 0)
	check(err)
	if len(lines) != 1 {
		t.Fatalf("items=%+v", lines)
	}
	order, _, err = orderStore.FindByID(ctx, order.ID)
	check(err)
	if order.TotalPrice != "37.02" {
		t.Fatalf("total after update: %s", order.TotalPrice)
	}
	// Overflow of the order total must roll back the quantity change too.
	tooLarge := item
	tooLarge.Quantity = 1_000_000_000
	if _, err := itemStore.Update(ctx, tooLarge); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("overflow must be rejected: %v", err)
	}
	unchanged, _, err := itemStore.FindByID(ctx, item.ID, order.ID)
	check(err)
	if unchanged.Quantity != 3 {
		t.Fatalf("failed total update changed quantity: %+v", unchanged)
	}
	check(itemStore.Delete(ctx, item.ID, order.ID))
	order, _, err = orderStore.FindByID(ctx, order.ID)
	check(err)
	if order.TotalPrice != "0.00" {
		t.Fatalf("total after delete: %s", order.TotalPrice)
	}
	item, err = itemStore.Create(ctx, orders.Item{OrderID: order.ID, ProductID: product.ID, Quantity: 1})
	check(err)
	order.Status = "PAID"
	_, err = orderStore.Update(ctx, order)
	check(err)
	if err := itemStore.Delete(ctx, item.ID, order.ID); !errors.Is(err, domain.ErrOrderLocked) {
		t.Fatalf("paid order could be edited: %v", err)
	}
	_, err = orderStore.Cancel(ctx, order.ID, user.ID)
	if !errors.Is(err, domain.ErrOrderLocked) {
		t.Fatalf("paid order could be cancelled by owner: %v", err)
	}
	order.Status = "PENDING"
	_, err = orderStore.Update(ctx, order)
	check(err)
	_, err = orderStore.Cancel(ctx, order.ID, user.ID)
	check(err)
	ordersList, err := orderStore.ListByUser(ctx, user.ID, 20, 0)
	check(err)
	if len(ordersList) != 1 {
		t.Fatalf("orders=%+v", ordersList)
	}
	profile, found, err := profileStore.FindByUserID(ctx, user.ID)
	check(err)
	if !found || profile.OrdersCount != 1 {
		t.Fatalf("profile count: %+v", profile)
	}
	profile.Bio, profile.Address = "Updated bio", "New address"
	profile, err = profileStore.Update(ctx, profile)
	check(err)
	check(profileStore.Delete(ctx, user.ID))
	profile, err = profileStore.Create(ctx, users.Profile{UserID: user.ID, Bio: profile.Bio, Address: profile.Address})
	check(err)
	if profile.OrdersCount != 1 {
		t.Fatalf("recreated profile lost count: %+v", profile)
	}
	check(orderStore.Delete(ctx, order.ID))
	_, found, err = itemStore.FindByID(ctx, item.ID, order.ID)
	check(err)
	if found {
		t.Fatal("order item did not cascade")
	}
	profile, _, err = profileStore.FindByUserID(ctx, user.ID)
	check(err)
	if profile.OrdersCount != 0 {
		t.Fatalf("count after order deletion: %+v", profile)
	}
	check(productStore.Delete(ctx, product.ID))
	_, found, err = imageStore.FindByID(ctx, image.ID, product.ID)
	check(err)
	if found {
		t.Fatal("image did not cascade")
	}
	_, found, err = likeStore.Find(ctx, user.ID, product.ID)
	check(err)
	if found {
		t.Fatal("like did not cascade")
	}
	productID = 0
}
