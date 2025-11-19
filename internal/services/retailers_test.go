package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"errors"
	"testing"
)

// MockRetailersRepo is a mock implementation of IRetailersRepo
type MockRetailersRepo struct {
	GetRetailerByIDFunc    func(id int) (*models.Retailer, error)
	UpsertRetailerFunc     func(retailer *models.Retailer) error
	GetRetailerByEmailFunc func(email string) (*models.Retailer, error)
	UpdateRetailerFunc     func(retailer *models.Retailer) error
}

func (m *MockRetailersRepo) GetRetailerByID(id int) (*models.Retailer, error) {
	if m.GetRetailerByIDFunc != nil {
		return m.GetRetailerByIDFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRetailersRepo) UpsertRetailer(retailer *models.Retailer) error {
	if m.UpsertRetailerFunc != nil {
		return m.UpsertRetailerFunc(retailer)
	}
	return errors.New("not implemented")
}

func (m *MockRetailersRepo) GetRetailerByEmail(email string) (*models.Retailer, error) {
	if m.GetRetailerByEmailFunc != nil {
		return m.GetRetailerByEmailFunc(email)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRetailersRepo) UpdateRetailer(retailer *models.Retailer) error {
	if m.UpdateRetailerFunc != nil {
		return m.UpdateRetailerFunc(retailer)
	}
	return errors.New("not implemented")
}

func TestNewRetailersService(t *testing.T) {
	mockRepo := &MockRetailersRepo{}
	service := NewRetailersService(mockRepo)

	if service == nil {
		t.Fatal("NewRetailersService returned nil")
	}
	if service.retailersRepo != mockRepo {
		t.Error("NewRetailersService did not set retailersRepo correctly")
	}
}

func TestRetailersService_GetRetailer(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		setupMock     func() *MockRetailersRepo
		expectedError error
	}{
		{
			name: "successful retrieval",
			id:   1,
			setupMock: func() *MockRetailersRepo {
				return &MockRetailersRepo{
					GetRetailerByIDFunc: func(id int) (*models.Retailer, error) {
						return &models.Retailer{
							Id:      1,
							Name:    "Test Retailer",
							Email:   "test@retailer.com",
							Phone:   "+1234567890",
							Address: "123 Test St",
						}, nil
					},
				}
			},
			expectedError: nil,
		},
		{
			name: "retailer not found",
			id:   999,
			setupMock: func() *MockRetailersRepo {
				return &MockRetailersRepo{
					GetRetailerByIDFunc: func(id int) (*models.Retailer, error) {
						return &models.Retailer{}, repositories.ErrRetailerNotFound
					},
				}
			},
			expectedError: repositories.ErrRetailerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.setupMock()
			service := NewRetailersService(mockRepo)

			retailer, err := service.GetRetailer(tt.id)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if retailer.Id != tt.id {
					t.Errorf("Expected retailer ID %d, got %d", tt.id, retailer.Id)
				}
			}
		})
	}
}
