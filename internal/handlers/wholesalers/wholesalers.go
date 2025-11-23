package wholesalers

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

func GetWholesaler(wholesalersService *services.WholesalersService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid wholesaler ID"}, http.StatusBadRequest, nil)
			return
		}

		wholesaler, err := wholesalersService.GetWholesaler(id)
		if err != nil {
			if errors.Is(err, repositories.ErrWholesalerNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Wholesaler not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch wholesaler"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"wholesaler": wholesaler}, http.StatusOK, nil)
	}
}

// GetCurrentWholesaler gets the current authenticated wholesaler's profile
func GetCurrentWholesaler(wholesalersService *services.WholesalersService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(auth.UserEmailKey).(string)
		if !ok || email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		wholesaler, err := wholesalersService.GetWholesalerByEmail(email)
		if err != nil {
			if errors.Is(err, repositories.ErrWholesalerNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Wholesaler not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch wholesaler"}, http.StatusInternalServerError, nil)
			return
		}

		// Check onboarding status using the service method
		onboarded, err := wholesalersService.IsOnboarded(email)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to check onboarding status"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{
			"wholesaler": wholesaler,
			"onboarded":  onboarded,
		}, http.StatusOK, nil)
	}
}

type UpdateWholesalerRequest struct {
	BusinessName string   `json:"business_name"`
	Phone        string   `json:"phone"`
	Address      string   `json:"address"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

// UpdateCurrentWholesaler updates the current authenticated wholesaler's profile (onboarding)
func UpdateCurrentWholesaler(wholesalersService *services.WholesalersService, writeJSON jsonutils.JSONwriter, readJSON jsonutils.JSONreader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(auth.UserEmailKey).(string)
		if !ok || email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		var req UpdateWholesalerRequest
		err := readJSON(w, r, &req)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid request body"}, http.StatusBadRequest, nil)
			return
		}

		// Validation
		req.BusinessName = strings.TrimSpace(req.BusinessName)
		req.Phone = strings.TrimSpace(req.Phone)
		req.Address = strings.TrimSpace(req.Address)

		if req.BusinessName == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Business name is required"}, http.StatusBadRequest, nil)
			return
		}

		if req.Phone == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Phone is required"}, http.StatusBadRequest, nil)
			return
		}

		// Validate phone number format - support Indian phone numbers
		// Remove all non-digit characters for validation
		phoneDigits := regexp.MustCompile(`\D`).ReplaceAllString(req.Phone, "")

		// Handle Indian phone number formats
		// If 12 digits and starts with 91, remove 91
		if len(phoneDigits) == 12 && strings.HasPrefix(phoneDigits, "91") {
			phoneDigits = phoneDigits[2:]
		}
		// If 11 digits and starts with 0, remove 0
		if len(phoneDigits) == 11 && strings.HasPrefix(phoneDigits, "0") {
			phoneDigits = phoneDigits[1:]
		}

		// Phone must be exactly 10 digits after sanitization
		if len(phoneDigits) != 10 {
			writeJSON(w, jsonutils.Envelope{"error": "Phone number must be a valid 10-digit Indian number"}, http.StatusBadRequest, nil)
			return
		}

		// Validate phone contains only digits (check the digits-only version)
		phonePattern := regexp.MustCompile(`^\d{10}$`)
		if !phonePattern.MatchString(phoneDigits) {
			writeJSON(w, jsonutils.Envelope{"error": "Phone number must contain only digits"}, http.StatusBadRequest, nil)
			return
		}

		// Use the sanitized digits-only version for storage
		req.Phone = phoneDigits

		if req.Address == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Address is required"}, http.StatusBadRequest, nil)
			return
		}

		// Name comes from Google OAuth, not from user input
		wholesaler, err := wholesalersService.UpdateWholesaler(email, req.BusinessName, req.Phone, req.Address, req.Latitude, req.Longitude)
		if err != nil {
			if errors.Is(err, repositories.ErrWholesalerNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Wholesaler not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to update wholesaler"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{
			"wholesaler": wholesaler,
			"onboarded": true,
			"message":   "Profile updated successfully",
		}, http.StatusOK, nil)
	}
}

