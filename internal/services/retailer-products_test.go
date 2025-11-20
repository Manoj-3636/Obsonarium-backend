package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"errors"
	"testing"
)

// MockRetailerProductsRepo is a mock implementation of IRetailerProductsRepo
type MockRetailerProductsRepo struct {
	GetProductsFunc           func() ([]models.RetailerProduct, error)
	SearchProductsFunc        func(keyword string) ([]models.RetailerProduct, error)
	GetProductFunc            func(id int) (*models.RetailerProduct, error)
	GetProductsByRetailerIDFunc func(retailerID int) ([]models.RetailerProduct, error)
}

func (m *MockRetailerProductsRepo) GetProducts() ([]models.RetailerProduct, error) {
	if m.GetProductsFunc != nil {
		return m.GetProductsFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockRetailerProductsRepo) SearchProducts(keyword string) ([]models.RetailerProduct, error) {
	if m.SearchProductsFunc != nil {
		return m.SearchProductsFunc(keyword)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRetailerProductsRepo) GetProduct(id int) (*models.RetailerProduct, error) {
	if m.GetProductFunc != nil {
		return m.GetProductFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRetailerProductsRepo) GetProductsByRetailerID(retailerID int) ([]models.RetailerProduct, error) {
	if m.GetProductsByRetailerIDFunc != nil {
		return m.GetProductsByRetailerIDFunc(retailerID)
	}
	return nil, errors.New("not implemented")
}

func TestNewRetailerProductsService(t *testing.T) {
	mockRepo := &MockRetailerProductsRepo{}
	service := NewRetailerProductsService(mockRepo)

	if service == nil {
		t.Fatal("NewRetailerProductsService returned nil")
	}
	if service.productsRepo != mockRepo {
		t.Error("NewRetailerProductsService did not set productsRepo correctly")
	}
}

func TestRetailerProductsService_GetProducts(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		setupMock     func() *MockRetailerProductsRepo
		expectedCount int
		expectedError bool
	}{
		{
			name:  "get products without query",
			query: "",
			setupMock: func() *MockRetailerProductsRepo {
				return &MockRetailerProductsRepo{
					GetProductsFunc: func() ([]models.RetailerProduct, error) {
						return []models.RetailerProduct{
							{Id: 1, Name: "Product 1"},
							{Id: 2, Name: "Product 2"},
						}, nil
					},
				}
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:  "search products with query",
			query: "telescope",
			setupMock: func() *MockRetailerProductsRepo {
				return &MockRetailerProductsRepo{
					SearchProductsFunc: func(keyword string) ([]models.RetailerProduct, error) {
						return []models.RetailerProduct{
							{Id: 1, Name: "Telescope Product"},
						}, nil
					},
				}
			},
			expectedCount: 1,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.setupMock()
			service := NewRetailerProductsService(mockRepo)

			products, err := service.GetProducts(tt.query)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(products) != tt.expectedCount {
					t.Errorf("Expected %d products, got %d", tt.expectedCount, len(products))
				}
			}
		})
	}
}

func TestRetailerProductsService_GetProduct(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		setupMock     func() *MockRetailerProductsRepo
		expectedError error
	}{
		{
			name: "successful retrieval",
			id:   1,
			setupMock: func() *MockRetailerProductsRepo {
				return &MockRetailerProductsRepo{
					GetProductFunc: func(id int) (*models.RetailerProduct, error) {
						return &models.RetailerProduct{Id: 1, Name: "Test Product"}, nil
					},
				}
			},
			expectedError: nil,
		},
		{
			name: "product not found",
			id:   999,
			setupMock: func() *MockRetailerProductsRepo {
				return &MockRetailerProductsRepo{
					GetProductFunc: func(id int) (*models.RetailerProduct, error) {
						return &models.RetailerProduct{}, repositories.ErrProductNotFound
					},
				}
			},
			expectedError: repositories.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.setupMock()
			service := NewRetailerProductsService(mockRepo)

			product, err := service.GetProduct(tt.id)

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
				if product.Id != tt.id {
					t.Errorf("Expected product ID %d, got %d", tt.id, product.Id)
				}
			}
		})
	}
}

