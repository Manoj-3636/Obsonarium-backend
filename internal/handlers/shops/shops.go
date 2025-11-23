package shops

import (
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
	"strconv"
)

// GetNearbyShops returns shops within a radius of the given coordinates
// GET /api/shops/nearby?lat=<lat>&lon=<lon>&radius=<km>
func GetNearbyShops(shopsService *services.ShopsService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")
		radiusStr := r.URL.Query().Get("radius")

		if latStr == "" || lonStr == "" {
			writeJSON(w, jsonutils.Envelope{"error": "lat and lon parameters are required"}, http.StatusBadRequest, nil)
			return
		}

		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid lat parameter"}, http.StatusBadRequest, nil)
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid lon parameter"}, http.StatusBadRequest, nil)
			return
		}

		// Default radius to 10km if not provided
		radiusKm := 10.0
		if radiusStr != "" {
			radiusKm, err = strconv.ParseFloat(radiusStr, 64)
			if err != nil || radiusKm <= 0 {
				writeJSON(w, jsonutils.Envelope{"error": "Invalid radius parameter"}, http.StatusBadRequest, nil)
				return
			}
		}

		shops, err := shopsService.GetNearbyShops(lat, lon, radiusKm)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch nearby shops"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"shops": shops}, http.StatusOK, nil)
	}
}


