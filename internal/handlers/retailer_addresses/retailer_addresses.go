package retailer_addresses

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func GetAddresses(service *services.RetailerAddressesService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)

		addresses, err := service.GetAddresses(email)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"addresses": addresses}, http.StatusOK, nil)
	}
}

func AddAddress(service *services.RetailerAddressesService, writeJSON jsonutils.JSONwriter, readJSON jsonutils.JSONreader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)

		var address models.RetailerAddress
		if err := readJSON(w, r, &address); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		if err := service.CreateAddress(email, &address); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"address": address}, http.StatusCreated, nil)
	}
}

func RemoveAddress(service *services.RetailerAddressesService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)
		idStr := chi.URLParam(r, "id")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "invalid address ID"}, http.StatusBadRequest, nil)
			return
		}

		if err := service.DeleteAddress(email, id); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "address deleted"}, http.StatusOK, nil)
	}
}
