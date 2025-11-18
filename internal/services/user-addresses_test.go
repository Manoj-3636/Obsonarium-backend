package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"errors"
	"testing"
)

// MockUserAddressesRepo is a mock implementation of IUserAddressesRepo
type MockUserAddressesRepo struct {
	GetAddressesByUserIDFunc func(userID int) ([]models.UserAddress, error)
	AddAddressFunc            func(address *models.UserAddress) error
	RemoveAddressFunc         func(userID int, addressID int) error
}

func (m *MockUserAddressesRepo) GetAddressesByUserID(userID int) ([]models.UserAddress, error) {
	if m.GetAddressesByUserIDFunc != nil {
		return m.GetAddressesByUserIDFunc(userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUserAddressesRepo) AddAddress(address *models.UserAddress) error {
	if m.AddAddressFunc != nil {
		return m.AddAddressFunc(address)
	}
	return errors.New("not implemented")
}

func (m *MockUserAddressesRepo) RemoveAddress(userID int, addressID int) error {
	if m.RemoveAddressFunc != nil {
		return m.RemoveAddressFunc(userID, addressID)
	}
	return errors.New("not implemented")
}

func TestNewUserAddressesService(t *testing.T) {
	mockAddressesRepo := &MockUserAddressesRepo{}
	mockUsersRepo := &MockUsersRepo{}

	service := NewUserAddressesService(mockAddressesRepo, mockUsersRepo)
	if service == nil {
		t.Fatal("NewUserAddressesService returned nil")
	}
	if service.addressesRepo != mockAddressesRepo {
		t.Error("NewUserAddressesService did not set addressesRepo correctly")
	}
	if service.usersRepo != mockUsersRepo {
		t.Error("NewUserAddressesService did not set usersRepo correctly")
	}
}

func TestUserAddressesService_GetAddressesByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		setupMocks    func() (*MockUserAddressesRepo, *MockUsersRepo)
		expectedCount int
		expectedError bool
	}{
		{
			name:  "successful retrieval",
			email: "test@example.com",
			setupMocks: func() (*MockUserAddressesRepo, *MockUsersRepo) {
				mockUsersRepo := &MockUsersRepo{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{Id: 1, Email: email}, nil
					},
				}
				mockAddressesRepo := &MockUserAddressesRepo{
					GetAddressesByUserIDFunc: func(userID int) ([]models.UserAddress, error) {
						return []models.UserAddress{
							{Id: 1, User_id: 1, Label: "Home", Street_address: "123 Main St"},
							{Id: 2, User_id: 1, Label: "Work", Street_address: "456 Office Ave"},
						}, nil
					},
				}
				return mockAddressesRepo, mockUsersRepo
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			setupMocks: func() (*MockUserAddressesRepo, *MockUsersRepo) {
				mockUsersRepo := &MockUsersRepo{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return nil, repositories.ErrUserNotFound
					},
				}
				return &MockUserAddressesRepo{}, mockUsersRepo
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAddressesRepo, mockUsersRepo := tt.setupMocks()
			service := NewUserAddressesService(mockAddressesRepo, mockUsersRepo)

			addresses, err := service.GetAddressesByEmail(tt.email)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(addresses) != tt.expectedCount {
					t.Errorf("Expected %d addresses, got %d", tt.expectedCount, len(addresses))
				}
			}
		})
	}
}

func TestUserAddressesService_AddAddress(t *testing.T) {
	mockUsersRepo := &MockUsersRepo{
		GetUserByEmailFunc: func(email string) (*models.User, error) {
			return &models.User{Id: 1, Email: email}, nil
		},
	}

	t.Run("successful add", func(t *testing.T) {
		mockAddressesRepo := &MockUserAddressesRepo{
			AddAddressFunc: func(address *models.UserAddress) error {
				address.Id = 1
				return nil
			},
		}

		service := NewUserAddressesService(mockAddressesRepo, mockUsersRepo)
		address := &models.UserAddress{
			Label:          "Home",
			Street_address: "123 Main St",
			City:           "City",
			Postal_code:    "12345",
			Country:        "USA",
		}

		err := service.AddAddress("test@example.com", address)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if address.User_id != 1 {
			t.Errorf("Expected User_id 1, got %d", address.User_id)
		}
		if address.Id != 1 {
			t.Errorf("Expected Id 1, got %d", address.Id)
		}
	})
}

func TestUserAddressesService_RemoveAddress(t *testing.T) {
	mockUsersRepo := &MockUsersRepo{
		GetUserByEmailFunc: func(email string) (*models.User, error) {
			return &models.User{Id: 1, Email: email}, nil
		},
	}

	t.Run("successful removal", func(t *testing.T) {
		mockAddressesRepo := &MockUserAddressesRepo{
			RemoveAddressFunc: func(userID int, addressID int) error {
				return nil
			},
		}

		service := NewUserAddressesService(mockAddressesRepo, mockUsersRepo)
		err := service.RemoveAddress("test@example.com", 1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("address not found", func(t *testing.T) {
		mockAddressesRepo := &MockUserAddressesRepo{
			RemoveAddressFunc: func(userID int, addressID int) error {
				return repositories.ErrAddressNotFound
			},
		}

		service := NewUserAddressesService(mockAddressesRepo, mockUsersRepo)
		err := service.RemoveAddress("test@example.com", 999)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !errors.Is(err, repositories.ErrAddressNotFound) {
			t.Errorf("Expected ErrAddressNotFound, got %v", err)
		}
	})
}

