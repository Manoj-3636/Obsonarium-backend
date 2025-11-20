package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrWholesalerProductNotFound = errors.New("wholesaler product not found")

type IWholesalerProductRepository interface {
	GetProducts() ([]models.WholesalerProduct, error)
	SearchProducts(keyword string) ([]models.WholesalerProduct, error)
	GetProduct(id int) (*models.WholesalerProduct, error)
	GetProductsByWholesalerID(wholesalerID int) ([]models.WholesalerProduct, error)
	GetProductByIDForWholesaler(productID int, wholesalerID int) (*models.WholesalerProduct, error)
	CreateProduct(product *models.WholesalerProduct) (*models.WholesalerProduct, error)
	UpdateProduct(product *models.WholesalerProduct) (*models.WholesalerProduct, error)
	DeleteProduct(productID int, wholesalerID int) error
}

type WholesalerProductRepository struct {
	DB *sql.DB
}

func NewWholesalerProductRepository(db *sql.DB) *WholesalerProductRepository {
	return &WholesalerProductRepository{DB: db}
}

func (repo *WholesalerProductRepository) GetProducts() ([]models.WholesalerProduct, error) {
	query := `
		SELECT id, wholesaler_id, name, price, stock_qty, image_url, description
		FROM wholesaler_products
		ORDER BY updated_at DESC
		LIMIT 9
	`

	rows, err := repo.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.WholesalerProduct

	for rows.Next() {
		var product models.WholesalerProduct
		err := rows.Scan(
			&product.Id,
			&product.Wholesaler_id,
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

func (repo *WholesalerProductRepository) SearchProducts(keyword string) ([]models.WholesalerProduct, error) {
	query := `
        SELECT id, wholesaler_id, name, price, stock_qty, image_url, description
        FROM wholesaler_products
        WHERE
            to_tsvector('simple', name || ' ' || coalesce(description, ''))
            @@ plainto_tsquery('simple', $1)
        ORDER BY updated_at DESC
        LIMIT 50;
    `

	rows, err := repo.DB.Query(query, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.WholesalerProduct

	for rows.Next() {
		var product models.WholesalerProduct
		err := rows.Scan(
			&product.Id,
			&product.Wholesaler_id,
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

func (repo *WholesalerProductRepository) GetProduct(id int) (*models.WholesalerProduct, error) {
	query := `
		SELECT id, wholesaler_id, name, price, stock_qty, image_url, description
		FROM wholesaler_products
		WHERE id = $1
	`

	var product models.WholesalerProduct
	err := repo.DB.QueryRow(query, id).Scan(
		&product.Id,
		&product.Wholesaler_id,
		&product.Name,
		&product.Price,
		&product.Stock_qty,
		&product.Image_url,
		&product.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.WholesalerProduct{}, ErrWholesalerProductNotFound
		}
		return &models.WholesalerProduct{}, err
	}

	return &product, nil
}

func (repo *WholesalerProductRepository) GetProductsByWholesalerID(wholesalerID int) ([]models.WholesalerProduct, error) {
	query := `
		SELECT id, wholesaler_id, name, price, stock_qty, image_url, description
		FROM wholesaler_products
		WHERE wholesaler_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := repo.DB.Query(query, wholesalerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.WholesalerProduct

	for rows.Next() {
		var product models.WholesalerProduct
		err := rows.Scan(
			&product.Id,
			&product.Wholesaler_id,
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

func (repo *WholesalerProductRepository) GetProductByIDForWholesaler(productID int, wholesalerID int) (*models.WholesalerProduct, error) {
	query := `
		SELECT id, wholesaler_id, name, price, stock_qty, image_url, description
		FROM wholesaler_products
		WHERE id = $1 AND wholesaler_id = $2
	`

	var product models.WholesalerProduct
	err := repo.DB.QueryRow(query, productID, wholesalerID).Scan(
		&product.Id,
		&product.Wholesaler_id,
		&product.Name,
		&product.Price,
		&product.Stock_qty,
		&product.Image_url,
		&product.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.WholesalerProduct{}, ErrWholesalerProductNotFound
		}
		return &models.WholesalerProduct{}, err
	}

	return &product, nil
}

func (repo *WholesalerProductRepository) CreateProduct(product *models.WholesalerProduct) (*models.WholesalerProduct, error) {
	query := `
		INSERT INTO wholesaler_products (wholesaler_id, name, price, stock_qty, image_url, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, wholesaler_id, name, price, stock_qty, image_url, description
	`

	err := repo.DB.QueryRow(
		query,
		product.Wholesaler_id,
		product.Name,
		product.Price,
		product.Stock_qty,
		product.Image_url,
		product.Description,
	).Scan(
		&product.Id,
		&product.Wholesaler_id,
		&product.Name,
		&product.Price,
		&product.Stock_qty,
		&product.Image_url,
		&product.Description,
	)
	if err != nil {
		return &models.WholesalerProduct{}, err
	}

	return product, nil
}

func (repo *WholesalerProductRepository) UpdateProduct(product *models.WholesalerProduct) (*models.WholesalerProduct, error) {
	query := `
		UPDATE wholesaler_products
		SET name = $1,
		    price = $2,
		    stock_qty = $3,
		    image_url = $4,
		    description = $5,
		    updated_at = NOW()
		WHERE id = $6 AND wholesaler_id = $7
		RETURNING id, wholesaler_id, name, price, stock_qty, image_url, description
	`

	err := repo.DB.QueryRow(
		query,
		product.Name,
		product.Price,
		product.Stock_qty,
		product.Image_url,
		product.Description,
		product.Id,
		product.Wholesaler_id,
	).Scan(
		&product.Id,
		&product.Wholesaler_id,
		&product.Name,
		&product.Price,
		&product.Stock_qty,
		&product.Image_url,
		&product.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.WholesalerProduct{}, ErrWholesalerProductNotFound
		}
		return &models.WholesalerProduct{}, err
	}

	return product, nil
}

func (repo *WholesalerProductRepository) DeleteProduct(productID int, wholesalerID int) error {
	result, err := repo.DB.Exec(
		`DELETE FROM wholesaler_products WHERE id = $1 AND wholesaler_id = $2`,
		productID,
		wholesalerID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrWholesalerProductNotFound
	}

	return nil
}
