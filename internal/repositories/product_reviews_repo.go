package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrReviewNotFound = errors.New("review not found")

type IProductReviewsRepo interface {
	GetReviewsByProductID(productID int) ([]models.ProductReview, error)
	CreateReview(review *models.ProductReview) (*models.ProductReview, error)
}

type ProductReviewsRepo struct {
	DB *sql.DB
}

func NewProductReviewsRepo(db *sql.DB) *ProductReviewsRepo {
	return &ProductReviewsRepo{DB: db}
}

func (repo *ProductReviewsRepo) GetReviewsByProductID(productID int) ([]models.ProductReview, error) {
	query := `
		SELECT id, product_id, user_id, rating, comment, created_at, updated_at
		FROM product_reviews
		WHERE product_id = $1
		ORDER BY created_at DESC
	`

	rows, err := repo.DB.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.ProductReview

	for rows.Next() {
		var review models.ProductReview
		err := rows.Scan(
			&review.Id,
			&review.Product_id,
			&review.User_id,
			&review.Rating,
			&review.Comment,
			&review.Created_at,
			&review.Updated_at,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (repo *ProductReviewsRepo) CreateReview(review *models.ProductReview) (*models.ProductReview, error) {
	query := `
		INSERT INTO product_reviews (product_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, product_id, user_id, rating, comment, created_at, updated_at
	`

	var createdReview models.ProductReview
	err := repo.DB.QueryRow(
		query,
		review.Product_id,
		review.User_id,
		review.Rating,
		review.Comment,
	).Scan(
		&createdReview.Id,
		&createdReview.Product_id,
		&createdReview.User_id,
		&createdReview.Rating,
		&createdReview.Comment,
		&createdReview.Created_at,
		&createdReview.Updated_at,
	)
	if err != nil {
		return nil, err
	}

	return &createdReview, nil
}

