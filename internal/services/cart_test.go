package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"errors"
	"testing"
)

// MockCartRepo is a mock implementation of ICartRepo
type MockCartRepo struct {
	GetCartItemsByUserIDFunc func(userID int) ([]models.CartItem, error)
	AddCartItemFunc          func(userID int, productID int, quantity int) (int, error)
	RemoveCartItemFunc       func(userID int, productID int) error
	DecreaseCartItemFunc     func(userID int, productID int) (int, error)
	GetCartNumberFunc        func(userID int) (int, error)
}

func (m *MockCartRepo) GetCartItemsByUserID(userID int) ([]models.CartItem, error) {
	if m.GetCartItemsByUserIDFunc != nil {
		return m.GetCartItemsByUserIDFunc(userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockCartRepo) AddCartItem(userID int, productID int, quantity int) (int, error) {
	if m.AddCartItemFunc != nil {
		return m.AddCartItemFunc(userID, productID, quantity)
	}
	return 0, errors.New("not implemented")
}

func (m *MockCartRepo) RemoveCartItem(userID int, productID int) error {
	if m.RemoveCartItemFunc != nil {
		return m.RemoveCartItemFunc(userID, productID)
	}
	return errors.New("not implemented")
}

func (m *MockCartRepo) DecreaseCartItem(userID int, productID int) (int, error) {
	if m.DecreaseCartItemFunc != nil {
		return m.DecreaseCartItemFunc(userID, productID)
	}
	return 0, errors.New("not implemented")
}

func (m *MockCartRepo) GetCartNumber(userID int) (int, error) {
	if m.GetCartNumberFunc != nil {
		return m.GetCartNumberFunc(userID)
	}
	return 0, errors.New("not implemented")
}

func TestNewCartService(t *testing.T) {
	mockCartRepo := &MockCartRepo{}
	mockUsersRepo := &MockUsersRepo{}

	service := NewCartService(mockCartRepo, mockUsersRepo)
	if service == nil {
		t.Fatal("NewCartService returned nil")
	}
	if service.cartRepo != mockCartRepo {
		t.Error("NewCartService did not set cartRepo correctly")
	}
	if service.usersRepo != mockUsersRepo {
		t.Error("NewCartService did not set usersRepo correctly")
	}
}

func TestCartService_GetCartItemsByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		setupMocks    func() (*MockCartRepo, *MockUsersRepo)
		expectedError bool
	}{
		{
			name:  "successful retrieval",
			email: "test@example.com",
			setupMocks: func() (*MockCartRepo, *MockUsersRepo) {
				mockUsersRepo := &MockUsersRepo{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{Id: 1, Email: email}, nil
					},
				}
				mockCartRepo := &MockCartRepo{
					GetCartItemsByUserIDFunc: func(userID int) ([]models.CartItem, error) {
						return []models.CartItem{
							{Id: 1, User_id: 1, Product_id: 1, Quantity: 2},
						}, nil
					},
				}
				return mockCartRepo, mockUsersRepo
			},
			expectedError: false,
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			setupMocks: func() (*MockCartRepo, *MockUsersRepo) {
				mockUsersRepo := &MockUsersRepo{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return nil, repositories.ErrUserNotFound
					},
				}
				return &MockCartRepo{}, mockUsersRepo
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartRepo, mockUsersRepo := tt.setupMocks()
			service := NewCartService(mockCartRepo, mockUsersRepo)

			items, err := service.GetCartItemsByEmail(tt.email)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(items) == 0 {
					t.Error("Expected cart items, got empty slice")
				}
			}
		})
	}
}

func TestCartService_AddCartItem(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		productID     int
		quantity      int
		setupMocks    func() (*MockCartRepo, *MockUsersRepo)
		expectedQty   int
		expectedError bool
	}{
		{
			name:      "add item with quantity 1",
			email:     "test@example.com",
			productID: 1,
			quantity:  1,
			setupMocks: func() (*MockCartRepo, *MockUsersRepo) {
				mockUsersRepo := &MockUsersRepo{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{Id: 1, Email: email}, nil
					},
				}
				mockCartRepo := &MockCartRepo{
					AddCartItemFunc: func(userID int, productID int, quantity int) (int, error) {
						return 1, nil
					},
				}
				return mockCartRepo, mockUsersRepo
			},
			expectedQty:   1,
			expectedError: false,
		},
		{
			name:      "decrease item with quantity -1",
			email:     "test@example.com",
			productID: 1,
			quantity:  -1,
			setupMocks: func() (*MockCartRepo, *MockUsersRepo) {
				mockUsersRepo := &MockUsersRepo{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{Id: 1, Email: email}, nil
					},
				}
				mockCartRepo := &MockCartRepo{
					DecreaseCartItemFunc: func(userID int, productID int) (int, error) {
						return 0, nil
					},
				}
				return mockCartRepo, mockUsersRepo
			},
			expectedQty:   0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartRepo, mockUsersRepo := tt.setupMocks()
			service := NewCartService(mockCartRepo, mockUsersRepo)

			qty, err := service.AddCartItem(tt.email, tt.productID, tt.quantity)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if qty != tt.expectedQty {
					t.Errorf("Expected quantity %d, got %d", tt.expectedQty, qty)
				}
			}
		})
	}
}

func TestCartService_RemoveCartItem(t *testing.T) {
	mockUsersRepo := &MockUsersRepo{
		GetUserByEmailFunc: func(email string) (*models.User, error) {
			return &models.User{Id: 1, Email: email}, nil
		},
	}

	t.Run("successful removal", func(t *testing.T) {
		mockCartRepo := &MockCartRepo{
			RemoveCartItemFunc: func(userID int, productID int) error {
				return nil
			},
		}

		service := NewCartService(mockCartRepo, mockUsersRepo)
		err := service.RemoveCartItem("test@example.com", 1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("cart item not found", func(t *testing.T) {
		mockCartRepo := &MockCartRepo{
			RemoveCartItemFunc: func(userID int, productID int) error {
				return repositories.ErrCartItemNotFound
			},
		}

		service := NewCartService(mockCartRepo, mockUsersRepo)
		err := service.RemoveCartItem("test@example.com", 1)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !errors.Is(err, repositories.ErrCartItemNotFound) {
			t.Errorf("Expected ErrCartItemNotFound, got %v", err)
		}
	})
}

func TestCartService_GetCartNumberByEmail(t *testing.T) {
	mockUsersRepo := &MockUsersRepo{
		GetUserByEmailFunc: func(email string) (*models.User, error) {
			return &models.User{Id: 1, Email: email}, nil
		},
	}

	mockCartRepo := &MockCartRepo{
		GetCartNumberFunc: func(userID int) (int, error) {
			return 5, nil
		},
	}

	service := NewCartService(mockCartRepo, mockUsersRepo)
	count, err := service.GetCartNumberByEmail("test@example.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}
