package repositories

import (
	"Obsonarium-backend/internal/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewUserAddressesRepo(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	repo := NewUserAddressesRepo(db)
	if repo == nil {
		t.Fatal("NewUserAddressesRepo returned nil")
	}
	if repo.DB != db {
		t.Error("NewUserAddressesRepo did not set DB correctly")
	}
}

func TestUserAddressesRepo_GetAddressesByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserAddressesRepo(db)

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "label", "street_address", "city", "state", "postal_code", "country"}).
			AddRow(1, 1, "Home", "123 Main St", "City", "State", "12345", "USA").
			AddRow(2, 1, "Work", "456 Office Ave", "City", "State", "12345", "USA")
		mock.ExpectQuery("SELECT id, user_id, label, street_address, city, state, postal_code, country").
			WithArgs(1).
			WillReturnRows(rows)

		addresses, err := repo.GetAddressesByUserID(1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(addresses) != 2 {
			t.Errorf("Expected 2 addresses, got %d", len(addresses))
		}
		if addresses[0].Label != "Home" {
			t.Errorf("Expected first address label 'Home', got %s", addresses[0].Label)
		}
	})
}

func TestUserAddressesRepo_AddAddress(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserAddressesRepo(db)

	address := &models.UserAddress{
		User_id:        1,
		Label:          "Home",
		Street_address: "123 Main St",
		City:           "City",
		State:          "State",
		Postal_code:    "12345",
		Country:        "USA",
	}

	t.Run("successful add", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery("INSERT INTO user_addresses").
			WithArgs(address.User_id, address.Label, address.Street_address, address.City, address.State, address.Postal_code, address.Country).
			WillReturnRows(rows)

		err := repo.AddAddress(address)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if address.Id != 1 {
			t.Errorf("Expected address ID 1, got %d", address.Id)
		}
	})
}

func TestUserAddressesRepo_RemoveAddress(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserAddressesRepo(db)

	t.Run("successful removal", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM user_addresses").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveAddress(1, 1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("address not found", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM user_addresses").
			WithArgs(999, 1).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.RemoveAddress(1, 999)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !IsError(err, ErrAddressNotFound) {
			t.Errorf("Expected ErrAddressNotFound, got %v", err)
		}
	})
}

