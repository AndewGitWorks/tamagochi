package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AndewGitWorks/tamagochi/backend/internal/domain"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
	"github.com/AndewGitWorks/tamagochi/backend/internal/repository"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestItemCreateRollsBackIfTotalUpdateFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT status FROM tamagochi.orders").WithArgs(int64(5)).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("PENDING"))
	mock.ExpectQuery("INSERT INTO tamagochi.order_items").WithArgs(int64(5), int64(9), int64(2)).WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price"}).AddRow(1, 5, 9, 2, "10.00"))
	mock.ExpectExec("UPDATE tamagochi.orders SET total_price").WithArgs(int64(5)).WillReturnError(errors.New("total failed"))
	mock.ExpectRollback()
	store := repository.NewPostgresOrderItemStore(db)
	if _, err := store.Create(context.Background(), orders.Item{OrderID: 5, ProductID: 9, Quantity: 2}); err == nil {
		t.Fatal("expected failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestItemUpdateRejectsPaidOrderBeforeChangingItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT status FROM tamagochi.orders").WithArgs(int64(5)).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("PAID"))
	mock.ExpectRollback()
	store := repository.NewPostgresOrderItemStore(db)
	_, err = store.Update(context.Background(), orders.Item{ID: 1, OrderID: 5, Quantity: 3})
	if !errors.Is(err, domain.ErrOrderLocked) {
		t.Fatalf("got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
