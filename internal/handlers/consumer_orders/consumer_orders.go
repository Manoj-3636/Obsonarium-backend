package consumer_orders

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// GetActiveOrders gets active orders for the authenticated retailer (excludes delivered/rejected items)
func GetActiveOrders(
	ordersService *services.ConsumerOrdersService,
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

		orders, err := ordersService.GetActiveOrdersByRetailerID(retailer.Id)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch orders"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"orders": orders}, http.StatusOK, nil)
	}
}

// GetHistoryOrders gets completed orders for the authenticated retailer (delivered/rejected items only)
func GetHistoryOrders(
	ordersService *services.ConsumerOrdersService,
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

		orders, err := ordersService.GetHistoryOrdersByRetailerID(retailer.Id)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch order history"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"orders": orders}, http.StatusOK, nil)
	}
}

// UpdateOrderItemStatus updates the status of an order item
func UpdateOrderItemStatus(
	ordersService *services.ConsumerOrdersService,
	retailersService *services.RetailersService,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
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

		itemIDParam := chi.URLParam(r, "item_id")
		itemID, err := strconv.Atoi(itemIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid item ID"}, http.StatusBadRequest, nil)
			return
		}

		var req struct {
			Status string `json:"status"`
		}
		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		if req.Status == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Status is required"}, http.StatusBadRequest, nil)
			return
		}

		err = ordersService.UpdateOrderItemStatus(itemID, retailer.Id, req.Status)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "Order item status updated successfully"}, http.StatusOK, nil)
	}
}
