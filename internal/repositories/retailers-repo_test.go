package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewRetailersRepo(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	repo := NewRetailersRepo(db)
	if repo == nil {
		t.Fatal("NewRetailersRepo returned nil")
	}
	if repo.DB != db {
		t.Error("NewRetailersRepo did not set DB correctly")
	}
}

func TestRetailersRepo_GetRetailerByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewRetailersRepo(db)

	tests := []struct {
		name          string
		id            int
		setupMock     func()
		expectedRetailer *models.Retailer
		expectedError error
	}{
		{
			name: "successful retrieval",
			id:   1,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "phone", "address"}).
					AddRow(1, "Test Retailer", "test@retailer.com", "+1234567890", "123 Test St")
				mock.ExpectQuery("SELECT id, name, email, phone, address").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedRetailer: &models.Retailer{
				Id:      1,
				Name:    "Test Retailer",
				Email:   "test@retailer.com",
				Phone:   "+1234567890",
				Address: "123 Test St",
			},
			expectedError: nil,
		},
		{
			name: "retailer not found",
			id:   999,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, name, email, phone, address").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedRetailer: &models.Retailer{},
			expectedError:    ErrRetailerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			retailer, err := repo.GetRetailerByID(tt.id)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.expectedError)
				} else if !IsError(err, tt.expectedError) {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if retailer.Name != tt.expectedRetailer.Name {
					t.Errorf("Expected name %s, got %s", tt.expectedRetailer.Name, retailer.Name)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Mock expectations were not met: %v", err)
			}
		})
	}
}

