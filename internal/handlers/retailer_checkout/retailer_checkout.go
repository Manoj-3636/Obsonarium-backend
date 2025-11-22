package retailer_checkout

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"encoding/json"
	"net/http"
)

type RetailerCheckoutHandler struct {
	Service           *services.RetailerCheckoutService
	RetailersService  *services.RetailersService
}

func NewRetailerCheckoutHandler(
	service *services.RetailerCheckoutService,
	retailersService *services.RetailersService,
) *RetailerCheckoutHandler {
	return &RetailerCheckoutHandler{
		Service:          service,
		RetailersService: retailersService,
	}
}

func (h *RetailerCheckoutHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	// Get retailer email from context (set by RequireRetailer middleware)
	email, ok := r.Context().Value(auth.UserEmailKey).(string)
	if !ok || email == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get retailer ID
	retailer, err := h.RetailersService.GetRetailerByEmail(email)
	if err != nil {
		if err == repositories.ErrRetailerNotFound {
			http.Error(w, "Retailer not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch retailer", http.StatusInternalServerError)
		return
	}

	var req services.RetailerCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.Service.ProcessCheckout(r.Context(), retailer.Id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

