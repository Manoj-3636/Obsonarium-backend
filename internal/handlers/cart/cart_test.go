package cart

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockCartServiceForTesting wraps the service for testing
type MockCartServiceForTesting struct {
	GetCartItemsByEmailFunc func(email string) ([]models.CartItem, error)
	AddCartItemFunc          func(email string, productID int, quantity int) (int, error)
	RemoveCartItemFunc       func(email string, productID int) error
	GetCartNumberByEmailFunc func(email string) (int, error)
}

func createRequestWithContext(email string) *http.Request {
	r := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(r.Context(), auth.UserEmailKey, email)
	return r.WithContext(ctx)
}

func TestGetCart(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		setupService   func() *services.CartService
		expectedStatus int
	}{
		{
			name:  "unauthorized - no email",
			email: "",
			setupService: func() *services.CartService {
				// Create a service with mock repos
				mockCartRepo := &MockCartRepoForTesting{}
				mockUsersRepo := &MockUsersRepoForTesting{}
				return services.NewCartService(mockCartRepo, mockUsersRepo)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			handler := GetCart(service, jsonutils.WriteJSON)

			r := createRequestWithContext(tt.email)
			w := httptest.NewRecorder()

			handler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// Mock implementations for testing
type MockCartRepoForTesting struct{}

func (m *MockCartRepoForTesting) GetCartItemsByUserID(userID int) ([]models.CartItem, error) {
	return []models.CartItem{{Id: 1, User_id: userID, Product_id: 1, Quantity: 1}}, nil
}

func (m *MockCartRepoForTesting) AddCartItem(userID int, productID int, quantity int) (int, error) {
	return quantity, nil
}

func (m *MockCartRepoForTesting) RemoveCartItem(userID int, productID int) error {
	return nil
}

func (m *MockCartRepoForTesting) DecreaseCartItem(userID int, productID int) (int, error) {
	return 0, nil
}

func (m *MockCartRepoForTesting) GetCartNumber(userID int) (int, error) {
	return 1, nil
}

type MockUsersRepoForTesting struct{}

func (m *MockUsersRepoForTesting) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, repositories.ErrUserNotFound
	}
	return &models.User{Id: 1, Email: email}, nil
}

func (m *MockUsersRepoForTesting) UpsertUser(user *models.User) error {
	return nil
}

func TestAddCartItem(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		body           interface{}
		expectedStatus int
	}{
		{
			name:  "unauthorized",
			email: "",
			body:  map[string]interface{}{"product_id": 1, "quantity": 1},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "invalid product ID",
			email: "test@example.com",
			body: map[string]interface{}{
				"product_id": 0,
				"quantity":   1,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartRepo := &MockCartRepoForTesting{}
			mockUsersRepo := &MockUsersRepoForTesting{}
			service := services.NewCartService(mockCartRepo, mockUsersRepo)
			handler := AddCartItem(service, jsonutils.WriteJSON, jsonutils.NewJSONutils().Reader)

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

func TestRemoveCartItem(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		productID      string
		expectedStatus int
	}{
		{
			name:           "unauthorized",
			email:          "",
			productID:      "1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid product ID",
			email:          "test@example.com",
			productID:      "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartRepo := &MockCartRepoForTesting{}
			mockUsersRepo := &MockUsersRepoForTesting{}
			service := services.NewCartService(mockCartRepo, mockUsersRepo)
			handler := RemoveCartItem(service, jsonutils.WriteJSON)

			r := httptest.NewRequest("DELETE", "/api/cart/"+tt.productID, nil)
			r = r.WithContext(context.WithValue(r.Context(), auth.UserEmailKey, tt.email))
			w := httptest.NewRecorder()

			// Simulate chi URL param by setting it in context
			// In real usage, chi would set this
			if tt.productID != "invalid" {
				r = r.WithContext(context.WithValue(r.Context(), "product_id", tt.productID))
			}

			handler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetCartNumber(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		expectedStatus int
	}{
		{
			name:           "unauthorized",
			email:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "successful",
			email:          "test@example.com",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartRepo := &MockCartRepoForTesting{}
			mockUsersRepo := &MockUsersRepoForTesting{}
			service := services.NewCartService(mockCartRepo, mockUsersRepo)
			handler := GetCartNumber(service, jsonutils.WriteJSON)

			r := createRequestWithContext(tt.email)
			w := httptest.NewRecorder()

			handler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
