package retailers

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

// MockRetailersRepoForTesting implements IRetailersRepo
type MockRetailersRepoForTesting struct {
	GetRetailerByIDFunc func(id int) (*models.Retailer, error)
}

func (m *MockRetailersRepoForTesting) GetRetailerByID(id int) (*models.Retailer, error) {
	if m.GetRetailerByIDFunc != nil {
		return m.GetRetailerByIDFunc(id)
	}
	return nil, errors.New("not implemented")
}

func TestGetRetailer(t *testing.T) {
	tests := []struct {
		name           string
		retailerID     string
		setupService   func() *services.RetailersService
		expectedStatus int
	}{
		{
			name:       "successful retrieval",
			retailerID: "1",
			setupService: func() *services.RetailersService {
				mockRepo := &MockRetailersRepoForTesting{
					GetRetailerByIDFunc: func(id int) (*models.Retailer, error) {
						return &models.Retailer{Id: 1, Name: "Test Retailer"}, nil
					},
				}
				return services.NewRetailersService(mockRepo)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "retailer not found",
			retailerID: "999",
			setupService: func() *services.RetailersService {
				mockRepo := &MockRetailersRepoForTesting{
					GetRetailerByIDFunc: func(id int) (*models.Retailer, error) {
						return &models.Retailer{}, repositories.ErrRetailerNotFound
					},
				}
				return services.NewRetailersService(mockRepo)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "invalid retailer ID",
			retailerID: "invalid",
			setupService: func() *services.RetailersService {
				mockRepo := &MockRetailersRepoForTesting{}
				return services.NewRetailersService(mockRepo)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			handler := GetRetailer(service, jsonutils.WriteJSON)

			router := chi.NewRouter()
			router.Get("/api/retailers/{id}", handler)

			r := httptest.NewRequest("GET", "/api/retailers/"+tt.retailerID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

