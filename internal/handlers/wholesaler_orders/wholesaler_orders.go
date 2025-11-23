package wholesaler_orders

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

// GetActiveOrders gets active orders for the authenticated wholesaler (excludes delivered/rejected items)
func GetActiveOrders(
	ordersService *services.RetailerWholesaleOrdersService,
	wholesalersService *services.WholesalersService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get wholesaler email from context (set by RequireWholesaler middleware)
		email, ok := r.Context().Value(auth.UserEmailKey).(string)
		if !ok || email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		// Get wholesaler ID
		wholesaler, err := wholesalersService.GetWholesalerByEmail(email)
		if err != nil {
			if err == repositories.ErrWholesalerNotFound {
				writeJSON(w, jsonutils.Envelope{"error": "Wholesaler not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch wholesaler"}, http.StatusInternalServerError, nil)
			return
		}

		orders, err := ordersService.GetActiveOrdersByWholesalerID(wholesaler.Id)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch orders"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"orders": orders}, http.StatusOK, nil)
	}
}

// GetHistoryOrders gets completed orders for the authenticated wholesaler (delivered/rejected items only)
func GetHistoryOrders(
	ordersService *services.RetailerWholesaleOrdersService,
	wholesalersService *services.WholesalersService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get wholesaler email from context (set by RequireWholesaler middleware)
		email, ok := r.Context().Value(auth.UserEmailKey).(string)
		if !ok || email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		// Get wholesaler ID
		wholesaler, err := wholesalersService.GetWholesalerByEmail(email)
		if err != nil {
			if err == repositories.ErrWholesalerNotFound {
				writeJSON(w, jsonutils.Envelope{"error": "Wholesaler not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch wholesaler"}, http.StatusInternalServerError, nil)
			return
		}

		orders, err := ordersService.GetHistoryOrdersByWholesalerID(wholesaler.Id)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch order history"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"orders": orders}, http.StatusOK, nil)
	}
}

// UpdateOrderItemStatus updates the status of an order item
func UpdateOrderItemStatus(
	ordersService *services.RetailerWholesaleOrdersService,
	wholesalersService *services.WholesalersService,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get wholesaler email from context (set by RequireWholesaler middleware)
		email, ok := r.Context().Value(auth.UserEmailKey).(string)
		if !ok || email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		// Get wholesaler ID
		wholesaler, err := wholesalersService.GetWholesalerByEmail(email)
		if err != nil {
			if err == repositories.ErrWholesalerNotFound {
				writeJSON(w, jsonutils.Envelope{"error": "Wholesaler not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch wholesaler"}, http.StatusInternalServerError, nil)
			return
		}

		itemIDParam := chi.URLParam(r, "item_id")
		itemID, err := strconv.Atoi(itemIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid item ID"}, http.StatusBadRequest, nil)
			return
		}

		var req struct {
			Status       string `json:"status"`
			DeliveryDate string `json:"delivery_date,omitempty"`
		}
		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		if req.Status == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Status is required"}, http.StatusBadRequest, nil)
			return
		}

		var deliveryDate *time.Time
		if req.DeliveryDate != "" && req.Status == "accepted" {
			parsedDate, err := time.Parse("2006-01-02", req.DeliveryDate)
			if err != nil {
				writeJSON(w, jsonutils.Envelope{"error": "Invalid delivery date format. Use YYYY-MM-DD"}, http.StatusBadRequest, nil)
				return
			}
			deliveryDate = &parsedDate
		}

		err = ordersService.UpdateOrderItemStatus(itemID, wholesaler.Id, req.Status, deliveryDate)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "Order item status updated successfully"}, http.StatusOK, nil)
	}
}

