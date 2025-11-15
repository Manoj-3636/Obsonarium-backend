package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrProductNotFound = errors.New("product not found")

type IRetailerProductsRepo interface {
	GetProducts() ([]models.RetailerProduct, error)
	GetProduct(id int) (*models.RetailerProduct, error)
}

type RetailerProductsRepo struct {
	DB *sql.DB
}

func NewRetailerProductsRepo(db *sql.DB) *RetailerProductsRepo {
	return &RetailerProductsRepo{DB: db}
}

func (repo *RetailerProductsRepo) GetProducts() ([]models.RetailerProduct, error) {
	query := `
		SELECT id, retailer_id, name, price, stock_qty, image_url, description
		FROM retailer_products
		ORDER BY id`

	rows, err := repo.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.RetailerProduct

	for rows.Next() {
		var product models.RetailerProduct
		err := rows.Scan(
			&product.Id,
			&product.Retailer_id,
			&product.Name,
			&product.Price,
			&product.Stock_qty,
			&product.Image_url,
			&product.Description,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (repo *RetailerProductsRepo) GetProduct(id int) (*models.RetailerProduct, error) {
	query := `
		SELECT id, retailer_id, name, price, stock_qty, image_url, description
		FROM retailer_products
		WHERE id = $1`

	var product models.RetailerProduct

	row := repo.DB.QueryRow(query, id)

	err := row.Scan(
		&product.Id,
		&product.Retailer_id,
		&product.Name,
		&product.Price,
		&product.Stock_qty,
		&product.Image_url,
		&product.Description,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.RetailerProduct{}, ErrProductNotFound
		}
		return &models.RetailerProduct{}, err
	}

	return &product, nil
}
