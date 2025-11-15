package cart

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func GetCart(cartService *services.CartService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)
		if email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		cartItems, err := cartService.GetCartItemsByEmail(email)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch cart"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"cart": cartItems}, http.StatusOK, nil)
	}
}

func AddCartItem(cartService *services.CartService, writeJSON jsonutils.JSONwriter, readJSON jsonutils.JSONreader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)
		if email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		var requestBody struct {
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
		}

		err := readJSON(w, r, &requestBody)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid request body"}, http.StatusBadRequest, nil)
			return
		}

		if requestBody.ProductID <= 0 {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		if requestBody.Quantity <= 0 {
			writeJSON(w, jsonutils.Envelope{"error": "Quantity must be greater than 0"}, http.StatusBadRequest, nil)
			return
		}

		err = cartService.AddCartItem(email, requestBody.ProductID, requestBody.Quantity)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to add item to cart"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "Item added to cart"}, http.StatusOK, nil)
	}
}

func RemoveCartItem(cartService *services.CartService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := auth.GetUserEmailFromContext(r)
		if email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		productIDParam := chi.URLParam(r, "product_id")
		productID, err := strconv.Atoi(productIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		err = cartService.RemoveCartItem(email, productID)
		if err != nil {
			if errors.Is(err, repositories.ErrCartItemNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Cart item not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to remove item from cart"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "Item removed from cart"}, http.StatusOK, nil)
	}
}
