package user_addresses

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func GetAddresses(addressesService *services.UserAddressesService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)
		if email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		addresses, err := addressesService.GetAddressesByEmail(email)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch addresses"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"addresses": addresses}, http.StatusOK, nil)
	}
}

func AddAddress(addressesService *services.UserAddressesService, writeJSON jsonutils.JSONwriter, readJSON jsonutils.JSONreader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)
		if email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		var requestBody struct {
			Label          string `json:"label"`
			Street_address string `json:"street_address"`
			City           string `json:"city"`
			State          string `json:"state"`
			Postal_code    string `json:"postal_code"`
			Country        string `json:"country"`
		}

		err := readJSON(w, r, &requestBody)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid request body"}, http.StatusBadRequest, nil)
			return
		}

		if requestBody.Street_address == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Street address is required"}, http.StatusBadRequest, nil)
			return
		}

		if requestBody.City == "" {
			writeJSON(w, jsonutils.Envelope{"error": "City is required"}, http.StatusBadRequest, nil)
			return
		}

		if requestBody.Postal_code == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Postal code is required"}, http.StatusBadRequest, nil)
			return
		}

		if requestBody.Country == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Country is required"}, http.StatusBadRequest, nil)
			return
		}

		address := &models.UserAddress{
			Label:          requestBody.Label,
			Street_address: requestBody.Street_address,
			City:           requestBody.City,
			State:          requestBody.State,
			Postal_code:    requestBody.Postal_code,
			Country:        requestBody.Country,
		}

		err = addressesService.AddAddress(email, address)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to add address"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"address": address}, http.StatusOK, nil)
	}
}

func RemoveAddress(addressesService *services.UserAddressesService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)
		if email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		addressIDParam := chi.URLParam(r, "id")
		addressID, err := strconv.Atoi(addressIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid address ID"}, http.StatusBadRequest, nil)
			return
		}

		err = addressesService.RemoveAddress(email, addressID)
		if err != nil {
			if errors.Is(err, repositories.ErrAddressNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Address not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to remove address"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "Address removed successfully"}, http.StatusOK, nil)
	}
}
