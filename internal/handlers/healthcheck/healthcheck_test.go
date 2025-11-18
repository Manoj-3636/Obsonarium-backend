package healthcheck

import (
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHealthCheckHandler(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		expectedStatus int
		expectedEnv    string
	}{
		{
			name:           "production environment",
			env:            "prod",
			expectedStatus: http.StatusOK,
			expectedEnv:    "prod",
		},
		{
			name:           "development environment",
			env:            "dev",
			expectedStatus: http.StatusOK,
			expectedEnv:    "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHealthCheckHandler(tt.env, jsonutils.WriteJSON)

			r := httptest.NewRequest("GET", "/api/healthcheck", nil)
			w := httptest.NewRecorder()

			handler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Error("Expected Content-Type application/json")
			}
		})
	}
}

