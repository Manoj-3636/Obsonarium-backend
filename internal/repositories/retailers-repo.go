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
		SELECT id, name, business_name, email, phone, address
		FROM retailers
		WHERE id = $1`

	var retailer models.Retailer
	var businessName sql.NullString

	row := repo.DB.QueryRow(query, id)

	err := row.Scan(
		&retailer.Id,
		&retailer.Name,
		&businessName,
		&retailer.Email,
		&retailer.Phone,
		&retailer.Address,
	)

	if businessName.Valid {
		retailer.BusinessName = businessName.String
	}

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
        RETURNING id, email, name, business_name, phone, address
    `

	// Note: Phone, Address, and BusinessName are not updated here as they come from onboarding/profile update
	// and are not provided by Google Auth.

	var phone sql.NullString
	var address sql.NullString
	var businessName sql.NullString

	err := repo.DB.QueryRow(
		query,
		retailer.Email,
		retailer.Name,
	).Scan(
		&retailer.Id,
		&retailer.Email,
		&retailer.Name,
		&businessName,
		&phone,
		&address,
	)

	if businessName.Valid {
		retailer.BusinessName = businessName.String
	}
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
		SELECT id, name, business_name, email, phone, address
		FROM retailers
		WHERE email = $1`

	var retailer models.Retailer
	var phone sql.NullString
	var address sql.NullString
	var businessName sql.NullString

	row := repo.DB.QueryRow(query, email)

	err := row.Scan(
		&retailer.Id,
		&retailer.Name,
		&businessName,
		&retailer.Email,
		&phone,
		&address,
	)

	if businessName.Valid {
		retailer.BusinessName = businessName.String
	}
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
		SET business_name = $1, phone = $2, address = $3
		WHERE email = $4
		RETURNING id`

	// Note: name is not updated here - it only comes from Google OAuth during login

	err := repo.DB.QueryRow(
		query,
		retailer.BusinessName,
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
