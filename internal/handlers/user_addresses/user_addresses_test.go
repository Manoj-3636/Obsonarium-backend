package user_addresses

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

// MockUserAddressesRepoForTesting implements IUserAddressesRepo
type MockUserAddressesRepoForTesting struct {
	GetAddressesByUserIDFunc func(userID int) ([]models.UserAddress, error)
	AddAddressFunc            func(address *models.UserAddress) error
	RemoveAddressFunc         func(userID int, addressID int) error
}

func (m *MockUserAddressesRepoForTesting) GetAddressesByUserID(userID int) ([]models.UserAddress, error) {
	if m.GetAddressesByUserIDFunc != nil {
		return m.GetAddressesByUserIDFunc(userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUserAddressesRepoForTesting) AddAddress(address *models.UserAddress) error {
	if m.AddAddressFunc != nil {
		return m.AddAddressFunc(address)
	}
	return errors.New("not implemented")
}

func (m *MockUserAddressesRepoForTesting) RemoveAddress(userID int, addressID int) error {
	if m.RemoveAddressFunc != nil {
		return m.RemoveAddressFunc(userID, addressID)
	}
	return errors.New("not implemented")
}

// MockUsersRepoForTesting implements IUsersRepo
type MockUsersRepoForTesting struct {
	GetUserByEmailFunc func(email string) (*models.User, error)
	UpsertUserFunc     func(user *models.User) error
}

func (m *MockUsersRepoForTesting) GetUserByEmail(email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUsersRepoForTesting) UpsertUser(user *models.User) error {
	if m.UpsertUserFunc != nil {
		return m.UpsertUserFunc(user)
	}
	return errors.New("not implemented")
}

func createRequestWithEmail(email string) *http.Request {
	r := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(r.Context(), auth.UserEmailKey, email)
	return r.WithContext(ctx)
}

func TestGetAddresses(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		setupService   func() *services.UserAddressesService
		expectedStatus int
	}{
		{
			name:  "successful retrieval",
			email: "test@example.com",
			setupService: func() *services.UserAddressesService {
				mockAddressesRepo := &MockUserAddressesRepoForTesting{
					GetAddressesByUserIDFunc: func(userID int) ([]models.UserAddress, error) {
						return []models.UserAddress{
							{Id: 1, User_id: userID, Label: "Home", Street_address: "123 Main St"},
						}, nil
					},
				}
				mockUsersRepo := &MockUsersRepoForTesting{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{Id: 1, Email: email}, nil
					},
				}
				return services.NewUserAddressesService(mockAddressesRepo, mockUsersRepo)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "unauthorized",
			email: "",
			setupService: func() *services.UserAddressesService {
				return services.NewUserAddressesService(&MockUserAddressesRepoForTesting{}, &MockUsersRepoForTesting{})
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			handler := GetAddresses(service, jsonutils.WriteJSON)

			r := createRequestWithEmail(tt.email)
			w := httptest.NewRecorder()

			handler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAddAddress(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		body           interface{}
		expectedStatus int
	}{
		{
			name:  "unauthorized",
			email: "",
			body: map[string]interface{}{
				"street_address": "123 Main St",
				"city":           "City",
				"postal_code":    "12345",
				"country":        "USA",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "missing required fields",
			email: "test@example.com",
			body: map[string]interface{}{
				"city": "City",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAddressesRepo := &MockUserAddressesRepoForTesting{
				AddAddressFunc: func(address *models.UserAddress) error {
					address.Id = 1
					return nil
				},
			}
			mockUsersRepo := &MockUsersRepoForTesting{
				GetUserByEmailFunc: func(email string) (*models.User, error) {
					if email == "" {
						return nil, repositories.ErrUserNotFound
					}
					return &models.User{Id: 1, Email: email}, nil
				},
			}
			service := services.NewUserAddressesService(mockAddressesRepo, mockUsersRepo)
			handler := AddAddress(service, jsonutils.WriteJSON, jsonutils.NewJSONutils().Reader)

			bodyBytes, _ := json.Marshal(tt.body)
			r := httptest.NewRequest("POST", "/", bytes.NewBuffer(bodyBytes))
			r = r.WithContext(context.WithValue(r.Context(), auth.UserEmailKey, tt.email))
			w := httptest.NewRecorder()

			handler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRemoveAddress(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		addressID      string
		setupService   func() *services.UserAddressesService
		expectedStatus int
	}{
		{
			name:      "unauthorized",
			email:     "",
			addressID: "1",
			setupService: func() *services.UserAddressesService {
				return services.NewUserAddressesService(&MockUserAddressesRepoForTesting{}, &MockUsersRepoForTesting{})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "invalid address ID",
			email:     "test@example.com",
			addressID: "invalid",
			setupService: func() *services.UserAddressesService {
				mockUsersRepo := &MockUsersRepoForTesting{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{Id: 1, Email: email}, nil
					},
				}
				return services.NewUserAddressesService(&MockUserAddressesRepoForTesting{}, mockUsersRepo)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "address not found",
			email:     "test@example.com",
			addressID: "999",
			setupService: func() *services.UserAddressesService {
				mockAddressesRepo := &MockUserAddressesRepoForTesting{
					RemoveAddressFunc: func(userID int, addressID int) error {
						return repositories.ErrAddressNotFound
					},
				}
				mockUsersRepo := &MockUsersRepoForTesting{
					GetUserByEmailFunc: func(email string) (*models.User, error) {
						return &models.User{Id: 1, Email: email}, nil
					},
				}
				return services.NewUserAddressesService(mockAddressesRepo, mockUsersRepo)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			handler := RemoveAddress(service, jsonutils.WriteJSON)

			router := chi.NewRouter()
			router.Delete("/api/addresses/{id}", handler)

			r := httptest.NewRequest("DELETE", "/api/addresses/"+tt.addressID, nil)
			r = r.WithContext(context.WithValue(r.Context(), auth.UserEmailKey, tt.email))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

