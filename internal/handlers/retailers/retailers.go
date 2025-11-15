package retailers

import (
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func GetRetailer(retailersService *services.RetailersService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid retailer ID"}, http.StatusBadRequest, nil)
			return
		}

		retailer, err := retailersService.GetRetailer(id)
		if err != nil {
			if errors.Is(err, repositories.ErrRetailerNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Retailer not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch retailer"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"retailer": retailer}, http.StatusOK, nil)
	}
}
