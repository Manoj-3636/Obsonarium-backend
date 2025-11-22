package checkout

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/services"
	"encoding/json"
	"net/http"
)

type CheckoutHandler struct {
	Service *services.CheckoutService
}

func NewCheckoutHandler(service *services.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{Service: service}
}

func (h *CheckoutHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req services.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.Service.ProcessCheckout(r.Context(), userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Order placed successfully",
		"order":   order,
	})
}
