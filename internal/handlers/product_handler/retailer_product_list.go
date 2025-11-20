package product_handler

import (
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func ListRetailerProducts(
	productService *services.ProductService,
	retailersService *services.RetailersService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		retailer, err := getAuthenticatedRetailer(r, retailersService)
		if err != nil {
			handleRetailerError(w, err, writeJSON)
			return
		}

		products, err := productService.GetProductsByRetailer(retailer.Id)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch products"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"products": products}, http.StatusOK, nil)
	}
}

func GetRetailerProduct(
	productService *services.ProductService,
	retailersService *services.RetailersService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		retailer, err := getAuthenticatedRetailer(r, retailersService)
		if err != nil {
			handleRetailerError(w, err, writeJSON)
			return
		}

		productIDParam := chi.URLParam(r, "id")
		productID, err := strconv.Atoi(productIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		product, err := productService.GetProductByID(productID, retailer.Id)
		if err != nil {
			if errors.Is(err, repositories.ErrProductNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Product not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch product"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"product": product}, http.StatusOK, nil)
	}
}

