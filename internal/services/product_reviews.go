package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type ProductReviewsService struct {
	reviewsRepo repositories.IProductReviewsRepo
}

func NewProductReviewsService(reviewsRepo repositories.IProductReviewsRepo) *ProductReviewsService {
	return &ProductReviewsService{
		reviewsRepo: reviewsRepo,
	}
}

func (s *ProductReviewsService) GetReviewsByProductID(productID int) ([]models.ProductReview, error) {
	reviews, err := s.reviewsRepo.GetReviewsByProductID(productID)
	if err != nil {
		return nil, fmt.Errorf("service error fetching reviews: %w", err)
	}
	return reviews, nil
}

func (s *ProductReviewsService) CreateReview(review *models.ProductReview) (*models.ProductReview, error) {
	createdReview, err := s.reviewsRepo.CreateReview(review)
	if err != nil {
		return nil, fmt.Errorf("service error creating review: %w", err)
	}
	return createdReview, nil
}

