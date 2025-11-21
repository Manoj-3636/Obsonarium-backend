package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrProductQueryNotFound = errors.New("product query not found")

type IProductQueriesRepo interface {
	GetQueriesByRetailerID(retailerID int) ([]models.ProductQuery, error)
	GetQueriesByProductID(productID int) ([]models.ProductQuery, error)
	CreateQuery(query *models.ProductQuery) (*models.ProductQuery, error)
	ResolveQuery(queryID int, responseText string) (*models.ProductQuery, error)
}

type ProductQueriesRepo struct {
	DB *sql.DB
}

func NewProductQueriesRepo(db *sql.DB) *ProductQueriesRepo {
	return &ProductQueriesRepo{DB: db}
}

func (repo *ProductQueriesRepo) GetQueriesByRetailerID(retailerID int) ([]models.ProductQuery, error) {
	query := `
		SELECT q.id, q.product_id, q.user_id, q.query_text, q.response_text, q.is_resolved, 
		       q.created_at, q.updated_at, q.resolved_at
		FROM product_queries q
		JOIN retailer_products p ON p.id = q.product_id
		WHERE p.retailer_id = $1
		ORDER BY q.is_resolved ASC, q.created_at DESC
	`

	rows, err := repo.DB.Query(query, retailerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []models.ProductQuery

	for rows.Next() {
		var q models.ProductQuery
		var responseText sql.NullString
		var resolvedAt sql.NullString

		err := rows.Scan(
			&q.Id,
			&q.Product_id,
			&q.User_id,
			&q.Query_text,
			&responseText,
			&q.Is_resolved,
			&q.Created_at,
			&q.Updated_at,
			&resolvedAt,
		)
		if err != nil {
			return nil, err
		}

		if responseText.Valid {
			q.Response_text = &responseText.String
		}

		if resolvedAt.Valid {
			q.Resolved_at = &resolvedAt.String
		}

		queries = append(queries, q)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return queries, nil
}

func (repo *ProductQueriesRepo) GetQueriesByProductID(productID int) ([]models.ProductQuery, error) {
	query := `
		SELECT id, product_id, user_id, query_text, response_text, is_resolved, 
		       created_at, updated_at, resolved_at
		FROM product_queries
		WHERE product_id = $1
		ORDER BY created_at DESC
	`

	rows, err := repo.DB.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []models.ProductQuery

	for rows.Next() {
		var q models.ProductQuery
		var responseText sql.NullString
		var resolvedAt sql.NullString

		err := rows.Scan(
			&q.Id,
			&q.Product_id,
			&q.User_id,
			&q.Query_text,
			&responseText,
			&q.Is_resolved,
			&q.Created_at,
			&q.Updated_at,
			&resolvedAt,
		)
		if err != nil {
			return nil, err
		}

		if responseText.Valid {
			q.Response_text = &responseText.String
		}

		if resolvedAt.Valid {
			q.Resolved_at = &resolvedAt.String
		}

		queries = append(queries, q)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return queries, nil
}

func (repo *ProductQueriesRepo) CreateQuery(query *models.ProductQuery) (*models.ProductQuery, error) {
	sqlQuery := `
		INSERT INTO product_queries (product_id, user_id, query_text)
		VALUES ($1, $2, $3)
		RETURNING id, product_id, user_id, query_text, response_text, is_resolved, created_at, updated_at, resolved_at
	`

	var createdQuery models.ProductQuery
	var responseText sql.NullString
	var resolvedAt sql.NullString

	err := repo.DB.QueryRow(
		sqlQuery,
		query.Product_id,
		query.User_id,
		query.Query_text,
	).Scan(
		&createdQuery.Id,
		&createdQuery.Product_id,
		&createdQuery.User_id,
		&createdQuery.Query_text,
		&responseText,
		&createdQuery.Is_resolved,
		&createdQuery.Created_at,
		&createdQuery.Updated_at,
		&resolvedAt,
	)
	if err != nil {
		return nil, err
	}

	if responseText.Valid {
		createdQuery.Response_text = &responseText.String
	}

	if resolvedAt.Valid {
		createdQuery.Resolved_at = &resolvedAt.String
	}

	return &createdQuery, nil
}

func (repo *ProductQueriesRepo) ResolveQuery(queryID int, responseText string) (*models.ProductQuery, error) {
	sqlQuery := `
		UPDATE product_queries
		SET response_text = $1,
		    is_resolved = TRUE,
		    resolved_at = NOW(),
		    updated_at = NOW()
		WHERE id = $2
		RETURNING id, product_id, user_id, query_text, response_text, is_resolved, created_at, updated_at, resolved_at
	`

	var resolvedQuery models.ProductQuery
	var responseTextVal sql.NullString
	var resolvedAt sql.NullString

	err := repo.DB.QueryRow(
		sqlQuery,
		responseText,
		queryID,
	).Scan(
		&resolvedQuery.Id,
		&resolvedQuery.Product_id,
		&resolvedQuery.User_id,
		&resolvedQuery.Query_text,
		&responseTextVal,
		&resolvedQuery.Is_resolved,
		&resolvedQuery.Created_at,
		&resolvedQuery.Updated_at,
		&resolvedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductQueryNotFound
		}
		return nil, err
	}

	if responseTextVal.Valid {
		resolvedQuery.Response_text = &responseTextVal.String
	}

	if resolvedAt.Valid {
		resolvedQuery.Resolved_at = &resolvedAt.String
	}

	return &resolvedQuery, nil
}

