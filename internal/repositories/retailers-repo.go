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
	GetNearbyRetailers(lat, lon, radiusKm float64) ([]models.Retailer, error)
}

type RetailersRepo struct {
	DB *sql.DB
}

func NewRetailersRepo(db *sql.DB) *RetailersRepo {
	return &RetailersRepo{DB: db}
}

func (repo *RetailersRepo) GetRetailerByID(id int) (*models.Retailer, error) {
	query := `
		SELECT id, name, business_name, email, phone, address, latitude, longitude
		FROM retailers
		WHERE id = $1`

	var retailer models.Retailer
	var businessName sql.NullString
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64

	row := repo.DB.QueryRow(query, id)

	err := row.Scan(
		&retailer.Id,
		&retailer.Name,
		&businessName,
		&retailer.Email,
		&retailer.Phone,
		&retailer.Address,
		&latitude,
		&longitude,
	)

	if businessName.Valid {
		retailer.BusinessName = businessName.String
	}
	if latitude.Valid {
		retailer.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		retailer.Longitude = &longitude.Float64
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
		SELECT id, name, business_name, email, phone, address, latitude, longitude
		FROM retailers
		WHERE email = $1`

	var retailer models.Retailer
	var phone sql.NullString
	var address sql.NullString
	var businessName sql.NullString
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64

	row := repo.DB.QueryRow(query, email)

	err := row.Scan(
		&retailer.Id,
		&retailer.Name,
		&businessName,
		&retailer.Email,
		&phone,
		&address,
		&latitude,
		&longitude,
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
	if latitude.Valid {
		retailer.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		retailer.Longitude = &longitude.Float64
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
		SET business_name = $1, phone = $2, address = $3, latitude = $4, longitude = $5
		WHERE email = $6
		RETURNING id`

	// Note: name is not updated here - it only comes from Google OAuth during login

	err := repo.DB.QueryRow(
		query,
		retailer.BusinessName,
		retailer.Phone,
		retailer.Address,
		retailer.Latitude,
		retailer.Longitude,
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

// GetNearbyRetailers gets retailers within a radius using Haversine formula
func (repo *RetailersRepo) GetNearbyRetailers(lat, lon, radiusKm float64) ([]models.Retailer, error) {
	// Haversine formula: distance = 2 * R * asin(sqrt(sin²(Δlat/2) + cos(lat1) * cos(lat2) * sin²(Δlon/2)))
	// Where R = 6371 km (Earth's radius)
	// We'll use a simpler bounding box approach first, then filter by distance
	query := `
		SELECT id, name, business_name, email, phone, address, latitude, longitude
		FROM retailers
		WHERE latitude IS NOT NULL 
			AND longitude IS NOT NULL
			AND (
				6371 * acos(
					cos(radians($1)) * cos(radians(latitude)) *
					cos(radians(longitude) - radians($2)) +
					sin(radians($1)) * sin(radians(latitude))
				)
			) <= $3
		ORDER BY (
			6371 * acos(
				cos(radians($1)) * cos(radians(latitude)) *
				cos(radians(longitude) - radians($2)) +
				sin(radians($1)) * sin(radians(latitude))
			)
		)
		LIMIT 50
	`

	rows, err := repo.DB.Query(query, lat, lon, radiusKm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var retailers []models.Retailer
	for rows.Next() {
		var retailer models.Retailer
		var businessName sql.NullString
		var phone sql.NullString
		var address sql.NullString
		var latitude sql.NullFloat64
		var longitude sql.NullFloat64

		err := rows.Scan(
			&retailer.Id,
			&retailer.Name,
			&businessName,
			&retailer.Email,
			&phone,
			&address,
			&latitude,
			&longitude,
		)
		if err != nil {
			return nil, err
		}

		if businessName.Valid {
			retailer.BusinessName = businessName.String
		}
		if phone.Valid {
			retailer.Phone = phone.String
		}
		if address.Valid {
			retailer.Address = address.String
		}
		if latitude.Valid {
			retailer.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			retailer.Longitude = &longitude.Float64
		}

		retailers = append(retailers, retailer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return retailers, nil
}
