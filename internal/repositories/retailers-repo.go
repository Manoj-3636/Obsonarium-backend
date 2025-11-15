package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrRetailerNotFound = errors.New("retailer not found")

type IRetailersRepo interface {
	GetRetailerByID(id int) (*models.Retailer, error)
}

type RetailersRepo struct {
	DB *sql.DB
}

func NewRetailersRepo(db *sql.DB) *RetailersRepo {
	return &RetailersRepo{DB: db}
}

func (repo *RetailersRepo) GetRetailerByID(id int) (*models.Retailer, error) {
	query := `
		SELECT id, name, email, phone, address
		FROM retailers
		WHERE id = $1`

	var retailer models.Retailer

	row := repo.DB.QueryRow(query, id)

	err := row.Scan(
		&retailer.Id,
		&retailer.Name,
		&retailer.Email,
		&retailer.Phone,
		&retailer.Address,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.Retailer{}, ErrRetailerNotFound
		}
		return &models.Retailer{}, err
	}

	return &retailer, nil
}

