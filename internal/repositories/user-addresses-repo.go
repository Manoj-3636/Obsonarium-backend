package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrAddressNotFound = errors.New("address not found")

type IUserAddressesRepo interface {
	GetAddressesByUserID(userID int) ([]models.UserAddress, error)
	AddAddress(address *models.UserAddress) error
	RemoveAddress(userID int, addressID int) error
}

type UserAddressesRepo struct {
	DB *sql.DB
}

func NewUserAddressesRepo(db *sql.DB) *UserAddressesRepo {
	return &UserAddressesRepo{DB: db}
}

func (repo *UserAddressesRepo) GetAddressesByUserID(userID int) ([]models.UserAddress, error) {
	query := `
		SELECT id, user_id, label, street_address, city, state, postal_code, country, latitude, longitude
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := repo.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []models.UserAddress

	for rows.Next() {
		var address models.UserAddress
		var latitude sql.NullFloat64
		var longitude sql.NullFloat64
		err := rows.Scan(
			&address.Id,
			&address.User_id,
			&address.Label,
			&address.Street_address,
			&address.City,
			&address.State,
			&address.Postal_code,
			&address.Country,
			&latitude,
			&longitude,
		)
		if err != nil {
			return nil, err
		}
		if latitude.Valid {
			address.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			address.Longitude = &longitude.Float64
		}
		addresses = append(addresses, address)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

func (repo *UserAddressesRepo) AddAddress(address *models.UserAddress) error {
	query := `
		INSERT INTO user_addresses (user_id, label, street_address, city, state, postal_code, country, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	err := repo.DB.QueryRow(
		query,
		address.User_id,
		address.Label,
		address.Street_address,
		address.City,
		address.State,
		address.Postal_code,
		address.Country,
		address.Latitude,
		address.Longitude,
	).Scan(&address.Id)

	return err
}

func (repo *UserAddressesRepo) RemoveAddress(userID int, addressID int) error {
	query := `
		DELETE FROM user_addresses
		WHERE id = $1 AND user_id = $2`

	result, err := repo.DB.Exec(query, addressID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAddressNotFound
	}

	return nil
}
