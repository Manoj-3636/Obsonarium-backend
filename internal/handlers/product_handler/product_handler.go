package product_handler

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
	productService *services.ProductService,
	retailersService *services.RetailersService,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		retailer, err := getAuthenticatedRetailer(r, retailersService)
		if err != nil {
			handleRetailerError(w, err, writeJSON)
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

		product := &models.RetailerProduct{
			Retailer_id: retailer.Id,
			Name:        req.Name,
			Price:       req.Price,
			Stock_qty:   req.StockQty,
			Image_url:   req.ImageURL,
			Description: req.Description,
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
	productService *services.ProductService,
	retailersService *services.RetailersService,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
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

		product := &models.RetailerProduct{
			Id:          productID,
			Retailer_id: retailer.Id,
			Name:        req.Name,
			Price:       req.Price,
			Stock_qty:   req.StockQty,
			Image_url:   req.ImageURL,
			Description: req.Description,
		}

		updated, err := productService.UpdateProduct(product)
		if err != nil {
			if errors.Is(err, repositories.ErrProductNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Product not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to update product"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"product": updated}, http.StatusOK, nil)
	}
}

func DeleteProduct(
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

		err = productService.DeleteProduct(productID, retailer.Id)
		if err != nil {
			if errors.Is(err, repositories.ErrProductNotFound) {
				writeJSON(w, jsonutils.Envelope{"error": "Product not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to delete product"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "Product deleted"}, http.StatusOK, nil)
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

func getAuthenticatedRetailer(r *http.Request, retailersService *services.RetailersService) (*models.Retailer, error) {
	email, ok := r.Context().Value(auth.UserEmailKey).(string)
	if !ok || email == "" {
		return &models.Retailer{}, errUnauthorized
	}

	retailer, err := retailersService.GetRetailerByEmail(email)
	if err != nil {
		return &models.Retailer{}, err
	}

	return retailer, nil
}

func handleRetailerError(w http.ResponseWriter, err error, writeJSON jsonutils.JSONwriter) {
	switch {
	case errors.Is(err, errUnauthorized):
		writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
	case errors.Is(err, repositories.ErrRetailerNotFound):
		writeJSON(w, jsonutils.Envelope{"error": "Retailer not found"}, http.StatusNotFound, nil)
	default:
		writeJSON(w, jsonutils.Envelope{"error": "Failed to resolve retailer"}, http.StatusInternalServerError, nil)
	}
}

