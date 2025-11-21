package services

import (
	"Obsonarium-backend/internal/models"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// MockUsersRepo is a mock implementation of IUsersRepo
type MockUsersRepo struct {
	GetUserByEmailFunc func(email string) (*models.User, error)
	GetUserByIDFunc    func(id int) (*models.User, error)
	UpsertUserFunc     func(user *models.User) error
}

func (m *MockUsersRepo) GetUserByEmail(email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUsersRepo) GetUserByID(id int) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUsersRepo) UpsertUser(user *models.User) error {
	if m.UpsertUserFunc != nil {
		return m.UpsertUserFunc(user)
	}
	return errors.New("not implemented")
}

// MockWholesalersRepo is a mock implementation of IWholesalersRepo
type MockWholesalersRepo struct {
	UpsertWholesalerFunc func(wholesaler *models.Wholesaler) error
}

func (m *MockWholesalersRepo) GetWholesalerByID(id int) (*models.Wholesaler, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWholesalersRepo) UpsertWholesaler(wholesaler *models.Wholesaler) error {
	if m.UpsertWholesalerFunc != nil {
		return m.UpsertWholesalerFunc(wholesaler)
	}
	return errors.New("not implemented")
}

func (m *MockWholesalersRepo) GetWholesalerByEmail(email string) (*models.Wholesaler, error) {
	return nil, errors.New("not implemented")
}

func (m *MockWholesalersRepo) UpdateWholesaler(wholesaler *models.Wholesaler) error {
	return errors.New("not implemented")
}

func TestNewAuthService(t *testing.T) {
	os.Setenv("LOCOSYNC_SIGNING", "test-key")
	defer os.Unsetenv("LOCOSYNC_SIGNING")

	mockRepo := &MockUsersRepo{}
	mockRetailersRepo := &MockRetailersRepo{}
	mockWholesalersRepo := &MockWholesalersRepo{}
	service := NewAuthService(mockRepo, mockRetailersRepo, mockWholesalersRepo)

	if service == nil {
		t.Fatal("NewAuthService returned nil")
	}
	if service.usersRepo != mockRepo {
		t.Error("NewAuthService did not set usersRepo correctly")
	}
	if service.selfSigningKey != "test-key" {
		t.Error("NewAuthService did not set selfSigningKey correctly")
	}
}

func TestAuthService_CreateJWT(t *testing.T) {
	os.Setenv("LOCOSYNC_SIGNING", "test-secret-key-for-jwt-creation")
	defer os.Unsetenv("LOCOSYNC_SIGNING")

	mockRepo := &MockUsersRepo{}
	mockRetailersRepo := &MockRetailersRepo{}
	mockWholesalersRepo := &MockWholesalersRepo{}
	service := NewAuthService(mockRepo, mockRetailersRepo, mockWholesalersRepo)

	user := &models.User{
		Id:    1,
		Email: "test@example.com",
		Name:  "Test User",
	}

	token, err := service.CreateJWT(user)
	if err != nil {
		t.Fatalf("CreateJWT returned error: %v", err)
	}

	if token == "" {
		t.Fatal("CreateJWT returned empty token")
	}

	// Verify the token can be parsed
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key-for-jwt-creation"), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Parsed token is not valid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}

	if claims["sub"] != user.Email {
		t.Errorf("Expected sub claim %s, got %v", user.Email, claims["sub"])
	}

	if claims["iss"] != "Obsonarium" {
		t.Errorf("Expected iss claim 'Obsonarium', got %v", claims["iss"])
	}
}

func TestAuthService_VerifySelfToken(t *testing.T) {
	os.Setenv("LOCOSYNC_SIGNING", "test-secret-key-for-verification")
	defer os.Unsetenv("LOCOSYNC_SIGNING")

	mockRepo := &MockUsersRepo{}
	mockRetailersRepo := &MockRetailersRepo{}
	mockWholesalersRepo := &MockWholesalersRepo{}
	service := NewAuthService(mockRepo, mockRetailersRepo, mockWholesalersRepo)

	user := &models.User{
		Email: "test@example.com",
	}

	// Create a valid token
	token, err := service.CreateJWT(user)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Verify the token
	claims, err := service.VerifySelfToken(token)
	if err != nil {
		t.Fatalf("VerifySelfToken returned error: %v", err)
	}

	if claims == nil {
		t.Fatal("VerifySelfToken returned nil claims")
	}

	if (*claims)["sub"] != user.Email {
		t.Errorf("Expected sub claim %s, got %v", user.Email, (*claims)["sub"])
	}
}

func TestAuthService_VerifySelfToken_Expired(t *testing.T) {
	os.Setenv("LOCOSYNC_SIGNING", "test-secret-key")
	defer os.Unsetenv("LOCOSYNC_SIGNING")

	mockRepo := &MockUsersRepo{}
	mockRetailersRepo := &MockRetailersRepo{}
	mockWholesalersRepo := &MockWholesalersRepo{}
	service := NewAuthService(mockRepo, mockRetailersRepo, mockWholesalersRepo)

	// Create an expired token
	claims := jwt.MapClaims{
		"sub": "test@example.com",
		"exp": time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
		"iss": "Obsonarium",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	_, err := service.VerifySelfToken(tokenString)
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}

	// JWT library wraps the error, so check if it contains our error or the expired message
	if !errors.Is(err, ErrSelfTokenExpired) && !errors.Is(err, ErrSelfTokenVerify) {
		// Check if error message contains expired
		if err.Error() == "" || (err.Error() != "self token expired" && !contains(err.Error(), "expired")) {
			t.Errorf("Expected expired token error, got %v", err)
		}
	}
}

func TestAuthService_UpsertUser(t *testing.T) {
	os.Setenv("LOCOSYNC_SIGNING", "test-key")
	defer os.Unsetenv("LOCOSYNC_SIGNING")

	t.Run("successful upsert", func(t *testing.T) {
		mockRepo := &MockUsersRepo{
			UpsertUserFunc: func(user *models.User) error {
				if user.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %s", user.Email)
				}
				return nil
			},
		}

		mockRetailersRepo := &MockRetailersRepo{}
		mockWholesalersRepo := &MockWholesalersRepo{}
		service := NewAuthService(mockRepo, mockRetailersRepo, mockWholesalersRepo)
		err := service.UpsertUser("test@example.com", "Test User", "https://example.com/pfp.jpg")
		if err != nil {
			t.Errorf("UpsertUser returned error: %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &MockUsersRepo{
			UpsertUserFunc: func(user *models.User) error {
				return errors.New("database error")
			},
		}

		mockRetailersRepo := &MockRetailersRepo{}
		mockWholesalersRepo := &MockWholesalersRepo{}
		service := NewAuthService(mockRepo, mockRetailersRepo, mockWholesalersRepo)
		err := service.UpsertUser("test@example.com", "Test User", "https://example.com/pfp.jpg")
		if err == nil {
			t.Fatal("Expected error from UpsertUser, got nil")
		}
	})
}
