package product_reviews

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

type CreateReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// GetReviews gets all reviews for a specific product (public endpoint, no auth required)
func GetReviews(
	reviewsService *services.ProductReviewsService,
	writeJSON jsonutils.JSONwriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		productIDParam := chi.URLParam(r, "product_id")
		productID, err := strconv.Atoi(productIDParam)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid product ID"}, http.StatusBadRequest, nil)
			return
		}

		reviews, err := reviewsService.GetReviewsByProductID(productID)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch reviews"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"reviews": reviews}, http.StatusOK, nil)
	}
}

// CreateReview creates a new review (protected route - requires consumer authentication)
func CreateReview(
	reviewsService *services.ProductReviewsService,
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

		var req CreateReviewRequest
		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		// Validate rating
		if req.Rating < 1 || req.Rating > 5 {
			writeJSON(w, jsonutils.Envelope{"error": "Rating must be between 1 and 5"}, http.StatusBadRequest, nil)
			return
		}

		// Validate comment
		req.Comment = strings.TrimSpace(req.Comment)
		if req.Comment == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Comment is required"}, http.StatusBadRequest, nil)
			return
		}

		review := &models.ProductReview{
			Product_id: productID,
			User_id:    user.Id,
			Rating:     req.Rating,
			Comment:    req.Comment,
		}

		createdReview, err := reviewsService.CreateReview(review)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to create review"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"review": createdReview}, http.StatusCreated, nil)
	}
}
