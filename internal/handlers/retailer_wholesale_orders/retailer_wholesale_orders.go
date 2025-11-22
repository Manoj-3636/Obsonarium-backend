package retailer_wholesale_orders

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
)

// GetOrders gets all orders for the authenticated retailer
func GetOrders(
	ordersService *services.RetailerWholesaleOrdersService,
	retailersService *services.RetailersService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get retailer email from context (set by RequireRetailer middleware)
		email, ok := r.Context().Value(auth.UserEmailKey).(string)
		if !ok || email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		// Get retailer ID
		retailer, err := retailersService.GetRetailerByEmail(email)
		if err != nil {
			if err == repositories.ErrRetailerNotFound {
				writeJSON(w, jsonutils.Envelope{"error": "Retailer not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch retailer"}, http.StatusInternalServerError, nil)
			return
		}

		orders, err := ordersService.GetOrdersByRetailerID(retailer.Id)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch orders"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"orders": orders}, http.StatusOK, nil)
	}
}

