package wholesaler_product_handler

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

type productRequest struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	StockQty    int     `json:"stock_qty"`
	ImageURL    string  `json:"image_url"`
	Description string  `json:"description"`
}

func CreateProduct(
	productService *services.WholesalerProductService,
	wholesalersService *services.WholesalersService,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wholesaler, err := getAuthenticatedWholesaler(r, wholesalersService)
		if err != nil {
			handleWholesalerError(w, err, writeJSON)
			return
		}

		var req productRequest
		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		sanitizeProductRequest(&req)
		if err := validateProductRequest(req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		product := &models.WholesalerProduct{
			Wholesaler_id: wholesaler.Id,
			Name:          req.Name,
			Price:         req.Price,
			Stock_qty:     req.StockQty,
			Image_url:     req.ImageURL,
			Description:   req.Description,
		}

		created, err := productService.CreateProduct(product)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to create product"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"product": created}, http.StatusCreated, nil)
	}
}

func UpdateProduct(
	productService *services.WholesalerProductService,
	wholesalersService *services.WholesalersService,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wholesaler, err := getAuthenticatedWholesaler(r, wholesalersService)
		if err != nil {
			handleWholesalerError(w, err, writeJSON)
			return
		}

		productIDParam := chi.URLParam(r, "id")
		productID, err := strconv.Atoi(productIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		var req productRequest
		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		sanitizeProductRequest(&req)
		if err := validateProductRequest(req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		product := &models.WholesalerProduct{
			Id:            productID,
			Wholesaler_id: wholesaler.Id,
			Name:          req.Name,
			Price:         req.Price,
			Stock_qty:     req.StockQty,
			Image_url:     req.ImageURL,
			Description:   req.Description,
		}

		updated, err := productService.UpdateProduct(product)
		if err != nil {
			if errors.Is(err, repositories.ErrWholesalerProductNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Product not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to update product"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"product": updated}, http.StatusOK, nil)
	}
}

func GetWholesalerProduct(
	productService *services.WholesalerProductService,
	wholesalersService *services.WholesalersService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wholesaler, err := getAuthenticatedWholesaler(r, wholesalersService)
		if err != nil {
			handleWholesalerError(w, err, writeJSON)
			return
		}

		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		product, err := productService.GetProductByIDForWholesaler(id, wholesaler.Id)
		if err != nil {
			if errors.Is(err, repositories.ErrWholesalerProductNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Product not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch product"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"product": product}, http.StatusOK, nil)
	}
}

func ListWholesalerProducts(
	productService *services.WholesalerProductService,
	wholesalersService *services.WholesalersService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wholesaler, err := getAuthenticatedWholesaler(r, wholesalersService)
		if err != nil {
			handleWholesalerError(w, err, writeJSON)
			return
		}

		products, err := productService.GetProductsByWholesalerID(wholesaler.Id)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch products"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"products": products}, http.StatusOK, nil)
	}
}

func DeleteProduct(
	productService *services.WholesalerProductService,
	wholesalersService *services.WholesalersService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wholesaler, err := getAuthenticatedWholesaler(r, wholesalersService)
		if err != nil {
			handleWholesalerError(w, err, writeJSON)
			return
		}

		productIDParam := chi.URLParam(r, "id")
		productID, err := strconv.Atoi(productIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		err = productService.DeleteProduct(productID, wholesaler.Id)
		if err != nil {
			if errors.Is(err, repositories.ErrWholesalerProductNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Product not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to delete product"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "Product deleted successfully"}, http.StatusOK, nil)
	}
}

func sanitizeProductRequest(req *productRequest) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.ImageURL = strings.TrimSpace(req.ImageURL)
}

func validateProductRequest(req productRequest) error {
	if req.Name == "" {
		return errors.New("Product name is required")
	}
	if req.Price <= 0 {
		return errors.New("Price must be greater than zero")
	}
	if req.StockQty < 0 {
		return errors.New("Stock quantity cannot be negative")
	}
	if req.ImageURL == "" {
		return errors.New("Image URL is required")
	}
	return nil
}

var errUnauthorized = errors.New("unauthorized")

func getAuthenticatedWholesaler(r *http.Request, wholesalersService *services.WholesalersService) (*models.Wholesaler, error) {
	email, ok := r.Context().Value(auth.UserEmailKey).(string)
	if !ok || email == "" {
		return &models.Wholesaler{}, errUnauthorized
	}

	wholesaler, err := wholesalersService.GetWholesalerByEmail(email)
	if err != nil {
		return &models.Wholesaler{}, err
	}

	return wholesaler, nil
}

func handleWholesalerError(w http.ResponseWriter, err error, writeJSON jsonutils.JSONwriter) {
	switch {
	case errors.Is(err, errUnauthorized):
		writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
	case errors.Is(err, repositories.ErrWholesalerNotFound):
		writeJSON(w, jsonutils.Envelope{"error": "Wholesaler not found"}, http.StatusNotFound, nil)
	default:
		writeJSON(w, jsonutils.Envelope{"error": "Failed to resolve wholesaler"}, http.StatusInternalServerError, nil)
	}
}
