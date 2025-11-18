package repositories

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewCartRepo(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	repo := NewCartRepo(db)
	if repo == nil {
		t.Fatal("NewCartRepo returned nil")
	}
	if repo.DB != db {
		t.Error("NewCartRepo did not set DB correctly")
	}
}

func TestCartRepo_GetCartItemsByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCartRepo(db)

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"c.id", "c.user_id", "c.product_id", "c.quantity",
			"p.id", "p.retailer_id", "p.name", "p.price", "p.stock_qty", "p.image_url", "p.description",
		}).
			AddRow(1, 1, 1, 2, 1, 1, "Product", 99.99, 10, "https://example.com/img.jpg", "Description")
		mock.ExpectQuery("SELECT c.id, c.user_id, c.product_id, c.quantity").
			WithArgs(1).
			WillReturnRows(rows)

		items, err := repo.GetCartItemsByUserID(1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Errorf("Expected 1 cart item, got %d", len(items))
		}
		if items[0].Quantity != 2 {
			t.Errorf("Expected quantity 2, got %d", items[0].Quantity)
		}
	})
}

func TestCartRepo_AddCartItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCartRepo(db)

	t.Run("successful add", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"quantity"}).AddRow(3)
		mock.ExpectQuery("INSERT INTO cart_items").
			WithArgs(1, 1, 2).
			WillReturnRows(rows)

		qty, err := repo.AddCartItem(1, 1, 2)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if qty != 3 {
			t.Errorf("Expected quantity 3, got %d", qty)
		}
	})
}

func TestCartRepo_DecreaseCartItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCartRepo(db)

	t.Run("successful decrease", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"quantity"}).AddRow(1)
		mock.ExpectQuery("UPDATE cart_items").
			WithArgs(1, 1).
			WillReturnRows(rows)

		qty, err := repo.DecreaseCartItem(1, 1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if qty != 1 {
			t.Errorf("Expected quantity 1, got %d", qty)
		}
	})

	t.Run("item not found", func(t *testing.T) {
		mock.ExpectQuery("UPDATE cart_items").
			WithArgs(1, 999).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.DecreaseCartItem(1, 999)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !IsError(err, ErrCartItemNotFound) {
			t.Errorf("Expected ErrCartItemNotFound, got %v", err)
		}
	})
}

func TestCartRepo_RemoveCartItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCartRepo(db)

	t.Run("successful removal", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM cart_items").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveCartItem(1, 1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("item not found", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM cart_items").
			WithArgs(1, 999).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.RemoveCartItem(1, 999)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !IsError(err, ErrCartItemNotFound) {
			t.Errorf("Expected ErrCartItemNotFound, got %v", err)
		}
	})
}

func TestCartRepo_GetCartNumber(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCartRepo(db)

	t.Run("successful count", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"COALESCE(SUM(quantity),0)"}).AddRow(5)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(1).
			WillReturnRows(rows)

		count, err := repo.GetCartNumber(1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 5 {
			t.Errorf("Expected count 5, got %d", count)
		}
	})
}

