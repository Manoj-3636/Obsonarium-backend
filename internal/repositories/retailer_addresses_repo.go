package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
)

type IRetailerAddressesRepo interface {
	CreateAddress(address *models.RetailerAddress) error
	GetAddressesByRetailerID(retailerID int) ([]models.RetailerAddress, error)
	GetAddress(id int) (*models.RetailerAddress, error)
	DeleteAddress(id int) error
}

type RetailerAddressesRepo struct {
	DB *sql.DB
}

func NewRetailerAddressesRepo(db *sql.DB) *RetailerAddressesRepo {
	return &RetailerAddressesRepo{DB: db}
}

func (r *RetailerAddressesRepo) CreateAddress(address *models.RetailerAddress) error {
	query := `
		INSERT INTO retailer_addresses (retailer_id, label, street_address, city, state, postal_code, country)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	return r.DB.QueryRow(
		query,
		address.Retailer_id,
		address.Label,
		address.Street_address,
		address.City,
		address.State,
		address.Postal_code,
		address.Country,
	).Scan(&address.Id)
}

func (r *RetailerAddressesRepo) GetAddressesByRetailerID(retailerID int) ([]models.RetailerAddress, error) {
	query := `
		SELECT id, retailer_id, label, street_address, city, state, postal_code, country
		FROM retailer_addresses
		WHERE retailer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.DB.Query(query, retailerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []models.RetailerAddress
	for rows.Next() {
		var addr models.RetailerAddress
		err := rows.Scan(
			&addr.Id,
			&addr.Retailer_id,
			&addr.Label,
			&addr.Street_address,
			&addr.City,
			&addr.State,
			&addr.Postal_code,
			&addr.Country,
		)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addr)
	}

	return addresses, rows.Err()
}

func (r *RetailerAddressesRepo) GetAddress(id int) (*models.RetailerAddress, error) {
	query := `
		SELECT id, retailer_id, label, street_address, city, state, postal_code, country
		FROM retailer_addresses
		WHERE id = $1
	`

	var addr models.RetailerAddress
	err := r.DB.QueryRow(query, id).Scan(
		&addr.Id,
		&addr.Retailer_id,
		&addr.Label,
		&addr.Street_address,
		&addr.City,
		&addr.State,
		&addr.Postal_code,
		&addr.Country,
	)
	if err != nil {
		return nil, err
	}

	return &addr, nil
}

func (r *RetailerAddressesRepo) DeleteAddress(id int) error {
	query := `DELETE FROM retailer_addresses WHERE id = $1`
	_, err := r.DB.Exec(query, id)
	return err
}
