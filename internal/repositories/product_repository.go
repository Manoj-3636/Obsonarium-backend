package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

type IProductRepository interface {
	GetProductsByRetailerID(retailerID int) ([]models.RetailerProduct, error)
	GetProductByIDForRetailer(productID int, retailerID int) (*models.RetailerProduct, error)
	CreateProduct(product *models.RetailerProduct) (*models.RetailerProduct, error)
	UpdateProduct(product *models.RetailerProduct) (*models.RetailerProduct, error)
	DeleteProduct(productID int, retailerID int) error
}

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (repo *ProductRepository) GetProductsByRetailerID(retailerID int) ([]models.RetailerProduct, error) {
	query := `
		SELECT id, retailer_id, name, price, stock_qty, image_url, description
		FROM retailer_products
		WHERE retailer_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := repo.DB.Query(query, retailerID)
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

func (repo *ProductRepository) GetProductByIDForRetailer(productID int, retailerID int) (*models.RetailerProduct, error) {
	query := `
		SELECT id, retailer_id, name, price, stock_qty, image_url, description
		FROM retailer_products
		WHERE id = $1 AND retailer_id = $2
	`

	var product models.RetailerProduct
	err := repo.DB.QueryRow(query, productID, retailerID).Scan(
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

func (repo *ProductRepository) CreateProduct(product *models.RetailerProduct) (*models.RetailerProduct, error) {
	query := `
		INSERT INTO retailer_products (retailer_id, name, price, stock_qty, image_url, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, retailer_id, name, price, stock_qty, image_url, description
	`

	err := repo.DB.QueryRow(
		query,
		product.Retailer_id,
		product.Name,
		product.Price,
		product.Stock_qty,
		product.Image_url,
		product.Description,
	).Scan(
		&product.Id,
		&product.Retailer_id,
		&product.Name,
		&product.Price,
		&product.Stock_qty,
		&product.Image_url,
		&product.Description,
	)
	if err != nil {
		return &models.RetailerProduct{}, err
	}

	return product, nil
}

func (repo *ProductRepository) UpdateProduct(product *models.RetailerProduct) (*models.RetailerProduct, error) {
	query := `
		UPDATE retailer_products
		SET name = $1,
		    price = $2,
		    stock_qty = $3,
		    image_url = $4,
		    description = $5,
		    updated_at = NOW()
		WHERE id = $6 AND retailer_id = $7
		RETURNING id, retailer_id, name, price, stock_qty, image_url, description
	`

	err := repo.DB.QueryRow(
		query,
		product.Name,
		product.Price,
		product.Stock_qty,
		product.Image_url,
		product.Description,
		product.Id,
		product.Retailer_id,
	).Scan(
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

	return product, nil
}

func (repo *ProductRepository) DeleteProduct(productID int, retailerID int) error {
	result, err := repo.DB.Exec(
		`DELETE FROM retailer_products WHERE id = $1 AND retailer_id = $2`,
		productID,
		retailerID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

