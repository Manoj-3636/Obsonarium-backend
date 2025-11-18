package jsonutils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		data           Envelope
		status         int
		headers        http.Header
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful write",
			data:           Envelope{"message": "test"},
			status:         http.StatusOK,
			headers:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"test"}`,
		},
		{
			name:           "with headers",
			data:           Envelope{"data": "test"},
			status:         http.StatusCreated,
			headers:        http.Header{"X-Custom": []string{"value"}},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"data":"test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := WriteJSON(w, tt.data, tt.status, tt.headers)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if !bytes.Contains([]byte(body), []byte(tt.expectedBody)) {
				t.Errorf("Expected body to contain %s, got %s", tt.expectedBody, body)
			}

			if tt.headers != nil {
				if w.Header().Get("X-Custom") != "value" {
					t.Error("Expected custom header not set")
				}
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Error("Expected Content-Type application/json")
			}
		})
	}
}

func TestReadJSON(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		target        interface{}
		expectedError bool
	}{
		{
			name:          "valid JSON",
			body:          `{"name":"test","value":123}`,
			target:        &struct{ Name string `json:"name"`; Value int `json:"value"` }{},
			expectedError: false,
		},
		{
			name:          "invalid JSON",
			body:          `{invalid json}`,
			target:        &struct{}{},
			expectedError: true,
		},
		{
			name:          "empty body",
			body:          "",
			target:        &struct{}{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			err := readJSON(w, r, tt.target)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNewJSONutils(t *testing.T) {
	utils := NewJSONutils()

	if utils.Writer == nil {
		t.Error("Writer should not be nil")
	}
	if utils.Reader == nil {
		t.Error("Reader should not be nil")
	}
}

