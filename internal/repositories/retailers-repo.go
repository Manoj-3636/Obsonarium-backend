package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrRetailerNotFound = errors.New("retailer not found")

type IRetailersRepo interface {
	GetRetailerByID(id int) (*models.Retailer, error)
	UpsertRetailer(retailer *models.Retailer) error
	GetRetailerByEmail(email string) (*models.Retailer, error)
	UpdateRetailer(retailer *models.Retailer) error
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

func (repo *RetailersRepo) UpsertRetailer(retailer *models.Retailer) error {
	query := `
        INSERT INTO retailers (email, name)
        VALUES ($1, $2)
        ON CONFLICT (email) DO UPDATE
        SET 
            name = EXCLUDED.name
        RETURNING id, email, name, phone, address
    `

	// Note: Phone and Address are not updated here as they come from onboarding/profile update
	// and are not provided by Google Auth.

	var phone sql.NullString
	var address sql.NullString

	err := repo.DB.QueryRow(
		query,
		retailer.Email,
		retailer.Name,
	).Scan(
		&retailer.Id,
		&retailer.Email,
		&retailer.Name,
		&phone,
		&address,
	)

	if phone.Valid {
		retailer.Phone = phone.String
	}
	if address.Valid {
		retailer.Address = address.String
	}

	return err
}

func (repo *RetailersRepo) GetRetailerByEmail(email string) (*models.Retailer, error) {
	query := `
		SELECT id, name, email, phone, address
		FROM retailers
		WHERE email = $1`

	var retailer models.Retailer
	var phone sql.NullString
	var address sql.NullString

	row := repo.DB.QueryRow(query, email)

	err := row.Scan(
		&retailer.Id,
		&retailer.Name,
		&retailer.Email,
		&phone,
		&address,
	)

	if phone.Valid {
		retailer.Phone = phone.String
	}
	if address.Valid {
		retailer.Address = address.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.Retailer{}, ErrRetailerNotFound
		}
		return &models.Retailer{}, err
	}

	return &retailer, nil
}

func (repo *RetailersRepo) UpdateRetailer(retailer *models.Retailer) error {
	query := `
		UPDATE retailers
		SET name = $1, phone = $2, address = $3
		WHERE email = $4
		RETURNING id`

	err := repo.DB.QueryRow(
		query,
		retailer.Name,
		retailer.Phone,
		retailer.Address,
		retailer.Email,
	).Scan(&retailer.Id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRetailerNotFound
		}
		return err
	}

	return nil
}

