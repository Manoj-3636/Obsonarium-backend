package retailer_products

import (
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func GetProducts(productsService *services.RetailerProductsService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := productsService.GetProducts()
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch products"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"products": products}, http.StatusOK, nil)
	}
}

func GetProduct(productsService *services.RetailerProductsService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		product, err := productsService.GetProduct(id)
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
