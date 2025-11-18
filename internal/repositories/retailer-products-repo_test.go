package repositories

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewRetailerProductsRepo(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	repo := NewRetailerProductsRepo(db)
	if repo == nil {
		t.Fatal("NewRetailerProductsRepo returned nil")
	}
	if repo.DB != db {
		t.Error("NewRetailerProductsRepo did not set DB correctly")
	}
}

func TestRetailerProductsRepo_GetProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewRetailerProductsRepo(db)

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "retailer_id", "name", "price", "stock_qty", "image_url", "description"}).
			AddRow(1, 1, "Test Product", 99.99, 10, "https://example.com/img.jpg", "Test description")
		mock.ExpectQuery("SELECT id, retailer_id, name, price, stock_qty, image_url, description").
			WillReturnRows(rows)

		products, err := repo.GetProducts()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(products) != 1 {
			t.Errorf("Expected 1 product, got %d", len(products))
		}
		if products[0].Name != "Test Product" {
			t.Errorf("Expected product name 'Test Product', got %s", products[0].Name)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Mock expectations were not met: %v", err)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "retailer_id", "name", "price", "stock_qty", "image_url", "description"})
		mock.ExpectQuery("SELECT id, retailer_id, name, price, stock_qty, image_url, description").
			WillReturnRows(rows)

		products, err := repo.GetProducts()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(products) != 0 {
			t.Errorf("Expected 0 products, got %d", len(products))
		}
	})
}

func TestRetailerProductsRepo_GetProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewRetailerProductsRepo(db)

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "retailer_id", "name", "price", "stock_qty", "image_url", "description"}).
			AddRow(1, 1, "Test Product", 99.99, 10, "https://example.com/img.jpg", "Test description")
		mock.ExpectQuery("SELECT id, retailer_id, name, price, stock_qty, image_url, description").
			WithArgs(1).
			WillReturnRows(rows)

		product, err := repo.GetProduct(1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if product.Id != 1 {
			t.Errorf("Expected product ID 1, got %d", product.Id)
		}
		if product.Name != "Test Product" {
			t.Errorf("Expected product name 'Test Product', got %s", product.Name)
		}
	})

	t.Run("product not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, retailer_id, name, price, stock_qty, image_url, description").
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		product, err := repo.GetProduct(999)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !IsError(err, ErrProductNotFound) {
			t.Errorf("Expected ErrProductNotFound, got %v", err)
		}
		if product.Id != 0 {
			t.Error("Expected empty product")
		}
	})
}

func TestRetailerProductsRepo_SearchProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewRetailerProductsRepo(db)

	t.Run("successful search", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "retailer_id", "name", "price", "stock_qty", "image_url", "description"}).
			AddRow(1, 1, "Telescope", 99.99, 10, "https://example.com/img.jpg", "A great telescope")
		mock.ExpectQuery("SELECT id, retailer_id, name, price, stock_qty, image_url, description").
			WithArgs("telescope").
			WillReturnRows(rows)

		products, err := repo.SearchProducts("telescope")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(products) != 1 {
			t.Errorf("Expected 1 product, got %d", len(products))
		}
	})
}

func IsError(err error, target error) bool {
	return err != nil && err.Error() == target.Error()
}
