package product_queries

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

type CreateQueryRequest struct {
	QueryText string `json:"query_text"`
}

type ResolveQueryRequest struct {
	ResponseText string `json:"response_text"`
}

// GetQueries gets all queries for a retailer (protected route - requires retailer authentication)
func GetQueries(
	queriesService *services.ProductQueriesService,
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

		queries, err := queriesService.GetQueriesByRetailerID(retailer.Id)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch queries"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"queries": queries}, http.StatusOK, nil)
	}
}

// PostQuery creates a new query (protected route - requires consumer authentication)
func PostQuery(
	queriesService *services.ProductQueriesService,
	usersRepo repositories.IUsersRepo,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user email from context (set by RequireConsumer middleware)
		email, ok := r.Context().Value(auth.UserEmailKey).(string)
		if !ok || email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
			return
		}

		// Get user ID
		user, err := usersRepo.GetUserByEmail(email)
		if err != nil {
			if err == repositories.ErrUserNotFound {
				writeJSON(w, jsonutils.Envelope{"error": "User not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch user"}, http.StatusInternalServerError, nil)
			return
		}

		productIDParam := chi.URLParam(r, "product_id")
		productID, err := strconv.Atoi(productIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		var req CreateQueryRequest
		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		// Validate query text
		req.QueryText = strings.TrimSpace(req.QueryText)
		if req.QueryText == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Query text is required"}, http.StatusBadRequest, nil)
			return
		}

		query := &models.ProductQuery{
			Product_id: productID,
			User_id:    user.Id,
			Query_text: req.QueryText,
		}

		createdQuery, err := queriesService.CreateQuery(query)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to create query"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"query": createdQuery}, http.StatusCreated, nil)
	}
}

// ResolveQuery resolves a query with a response (protected route - requires retailer authentication)
func ResolveQuery(
	queriesService *services.ProductQueriesService,
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

		// Verify retailer exists (we don't need the full retailer object, just verification)
		_, err := retailersService.GetRetailerByEmail(email)
		if err != nil {
			if err == repositories.ErrRetailerNotFound {
				writeJSON(w, jsonutils.Envelope{"error": "Retailer not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch retailer"}, http.StatusInternalServerError, nil)
			return
		}

		queryIDParam := chi.URLParam(r, "query_id")
		queryID, err := strconv.Atoi(queryIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid query ID"}, http.StatusBadRequest, nil)
			return
		}

		var req ResolveQueryRequest
		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		// Validate response text
		req.ResponseText = strings.TrimSpace(req.ResponseText)
		if req.ResponseText == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Response text is required"}, http.StatusBadRequest, nil)
			return
		}

		resolvedQuery, err := queriesService.ResolveQuery(queryID, req.ResponseText)
		if err != nil {
			if err == repositories.ErrProductQueryNotFound {
				writeJSON(w, jsonutils.Envelope{"error": "Query not found"}, http.StatusNotFound, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to resolve query"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"query": resolvedQuery}, http.StatusOK, nil)
	}
}

