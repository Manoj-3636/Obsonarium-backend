package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrWholesalerNotFound = errors.New("wholesaler not found")

type IWholesalersRepo interface {
	GetWholesalerByID(id int) (*models.Wholesaler, error)
	UpsertWholesaler(wholesaler *models.Wholesaler) error
	GetWholesalerByEmail(email string) (*models.Wholesaler, error)
	UpdateWholesaler(wholesaler *models.Wholesaler) error
}

type WholesalersRepo struct {
	DB *sql.DB
}

func NewWholesalersRepo(db *sql.DB) *WholesalersRepo {
	return &WholesalersRepo{DB: db}
}

func (repo *WholesalersRepo) GetWholesalerByID(id int) (*models.Wholesaler, error) {
	query := `
		SELECT id, name, business_name, email, phone, address
		FROM wholesalers
		WHERE id = $1`

	var wholesaler models.Wholesaler
	var businessName sql.NullString

	row := repo.DB.QueryRow(query, id)

	err := row.Scan(
		&wholesaler.Id,
		&wholesaler.Name,
		&businessName,
		&wholesaler.Email,
		&wholesaler.Phone,
		&wholesaler.Address,
	)

	if businessName.Valid {
		wholesaler.BusinessName = businessName.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.Wholesaler{}, ErrWholesalerNotFound
		}
		return &models.Wholesaler{}, err
	}

	return &wholesaler, nil
}

func (repo *WholesalersRepo) UpsertWholesaler(wholesaler *models.Wholesaler) error {
	query := `
        INSERT INTO wholesalers (email, name)
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
		wholesaler.Email,
		wholesaler.Name,
	).Scan(
		&wholesaler.Id,
		&wholesaler.Email,
		&wholesaler.Name,
		&businessName,
		&phone,
		&address,
	)

	if businessName.Valid {
		wholesaler.BusinessName = businessName.String
	}
	if phone.Valid {
		wholesaler.Phone = phone.String
	}
	if address.Valid {
		wholesaler.Address = address.String
	}

	return err
}

func (repo *WholesalersRepo) GetWholesalerByEmail(email string) (*models.Wholesaler, error) {
	query := `
		SELECT id, name, business_name, email, phone, address
		FROM wholesalers
		WHERE email = $1`

	var wholesaler models.Wholesaler
	var phone sql.NullString
	var address sql.NullString
	var businessName sql.NullString

	row := repo.DB.QueryRow(query, email)

	err := row.Scan(
		&wholesaler.Id,
		&wholesaler.Name,
		&businessName,
		&wholesaler.Email,
		&phone,
		&address,
	)

	if businessName.Valid {
		wholesaler.BusinessName = businessName.String
	}
	if phone.Valid {
		wholesaler.Phone = phone.String
	}
	if address.Valid {
		wholesaler.Address = address.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.Wholesaler{}, ErrWholesalerNotFound
		}
		return &models.Wholesaler{}, err
	}

	return &wholesaler, nil
}

func (repo *WholesalersRepo) UpdateWholesaler(wholesaler *models.Wholesaler) error {
	query := `
		UPDATE wholesalers
		SET business_name = $1, phone = $2, address = $3
		WHERE email = $4
		RETURNING id`

	// Note: name is not updated here - it only comes from Google OAuth during login

	err := repo.DB.QueryRow(
		query,
		wholesaler.BusinessName,
		wholesaler.Phone,
		wholesaler.Address,
		wholesaler.Email,
	).Scan(&wholesaler.Id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWholesalerNotFound
		}
		return err
	}

	return nil
}

