package orders

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"io"
	"net/http"
	"os"
)

type OrdersHandler struct {
	ordersService *services.OrdersService
	jsonUtils     jsonutils.JSONutils
}

func NewOrdersHandler(ordersService *services.OrdersService, jsonUtils jsonutils.JSONutils) *OrdersHandler {
	return &OrdersHandler{
		ordersService: ordersService,
		jsonUtils:     jsonUtils,
	}
}

func (h *OrdersHandler) CreateConsumerCheckout(w http.ResponseWriter, r *http.Request) {
	email := auth.GetUserEmailFromContext(r)

	var req struct {
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
		AddressID  int    `json:"address_id"`
	}

	if err := h.jsonUtils.Reader(w, r, &req); err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	sessionURL, err := h.ordersService.CreateConsumerCheckoutByEmail(email, req.SuccessURL, req.CancelURL, req.AddressID)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	h.jsonUtils.Writer(w, jsonutils.Envelope{"url": sessionURL}, http.StatusOK, nil)
}

func (h *OrdersHandler) CreateRetailerCheckout(w http.ResponseWriter, r *http.Request) {
	email := r.Context().Value("user_email").(string)

	var req struct {
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}

	if err := h.jsonUtils.Reader(w, r, &req); err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	sessionURL, err := h.ordersService.CreateRetailerCheckoutByEmail(email, req.SuccessURL, req.CancelURL)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	h.jsonUtils.Writer(w, jsonutils.Envelope{"url": sessionURL}, http.StatusOK, nil)
}

func (h *OrdersHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.errorJSON(w, err, http.StatusServiceUnavailable)
		return
	}

	// Verify webhook signature
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	header := r.Header.Get("Stripe-Signature")

	if err := h.ordersService.HandleStripeWebhook(payload, header, webhookSecret); err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *OrdersHandler) errorJSON(w http.ResponseWriter, err error, status int) {
	h.jsonUtils.Writer(w, jsonutils.Envelope{"error": err.Error()}, status, nil)
}
