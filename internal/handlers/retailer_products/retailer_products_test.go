package retailer_products

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

// MockRetailerProductsServiceForTesting is a mock for testing
type MockRetailerProductsServiceForTesting struct {
	GetProductsFunc func(q string) ([]models.RetailerProduct, error)
	GetProductFunc  func(id int) (*models.RetailerProduct, error)
}

func TestGetProducts(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupService   func() *services.RetailerProductsService
		expectedStatus int
	}{
		{
			name:  "successful retrieval - no query",
			query: "",
			setupService: func() *services.RetailerProductsService {
				mockRepo := &MockRetailerProductsRepoForTesting{
					GetProductsFunc: func() ([]models.RetailerProduct, error) {
						return []models.RetailerProduct{
							{Id: 1, Name: "Product 1"},
						}, nil
					},
				}
				return services.NewRetailerProductsService(mockRepo)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "successful search - with query",
			query: "telescope",
			setupService: func() *services.RetailerProductsService {
				mockRepo := &MockRetailerProductsRepoForTesting{
					SearchProductsFunc: func(keyword string) ([]models.RetailerProduct, error) {
						return []models.RetailerProduct{
							{Id: 1, Name: "Telescope Product"},
						}, nil
					},
				}
				return services.NewRetailerProductsService(mockRepo)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "service error",
			query: "",
			setupService: func() *services.RetailerProductsService {
				mockRepo := &MockRetailerProductsRepoForTesting{
					GetProductsFunc: func() ([]models.RetailerProduct, error) {
						return nil, errors.New("database error")
					},
				}
				return services.NewRetailerProductsService(mockRepo)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			handler := GetProducts(service, jsonutils.WriteJSON)

			url := "/api/shop"
			if tt.query != "" {
				url += "?q=" + tt.query
			}
			r := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetProduct(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		setupService   func() *services.RetailerProductsService
		expectedStatus int
	}{
		{
			name:      "successful retrieval",
			productID: "1",
			setupService: func() *services.RetailerProductsService {
				mockRepo := &MockRetailerProductsRepoForTesting{
					GetProductFunc: func(id int) (*models.RetailerProduct, error) {
						return &models.RetailerProduct{Id: 1, Name: "Product 1"}, nil
					},
				}
				return services.NewRetailerProductsService(mockRepo)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "product not found",
			productID: "999",
			setupService: func() *services.RetailerProductsService {
				mockRepo := &MockRetailerProductsRepoForTesting{
					GetProductFunc: func(id int) (*models.RetailerProduct, error) {
						return &models.RetailerProduct{}, repositories.ErrProductNotFound
					},
				}
				return services.NewRetailerProductsService(mockRepo)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "invalid product ID",
			productID: "invalid",
			setupService: func() *services.RetailerProductsService {
				mockRepo := &MockRetailerProductsRepoForTesting{}
				return services.NewRetailerProductsService(mockRepo)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			handler := GetProduct(service, jsonutils.WriteJSON)

			// Use chi router to properly set URL params
			router := chi.NewRouter()
			router.Get("/api/shop/{id}", handler)

			r := httptest.NewRequest("GET", "/api/shop/"+tt.productID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// MockRetailerProductsRepoForTesting implements IRetailerProductsRepo
type MockRetailerProductsRepoForTesting struct {
	GetProductsFunc    func() ([]models.RetailerProduct, error)
	SearchProductsFunc func(keyword string) ([]models.RetailerProduct, error)
	GetProductFunc     func(id int) (*models.RetailerProduct, error)
}

func (m *MockRetailerProductsRepoForTesting) GetProducts() ([]models.RetailerProduct, error) {
	if m.GetProductsFunc != nil {
		return m.GetProductsFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockRetailerProductsRepoForTesting) SearchProducts(keyword string) ([]models.RetailerProduct, error) {
	if m.SearchProductsFunc != nil {
		return m.SearchProductsFunc(keyword)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRetailerProductsRepoForTesting) GetProduct(id int) (*models.RetailerProduct, error) {
	if m.GetProductFunc != nil {
		return m.GetProductFunc(id)
	}
	return nil, errors.New("not implemented")
}
