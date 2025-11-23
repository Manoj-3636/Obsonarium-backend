package repositories

import (
	"database/sql"
	"fmt"
	"time"
)

type IConsumerOTPRepo interface {
	StoreOTP(email string, otp string, expiresAt time.Time) error
	ValidateOTP(email string, otp string) (bool, error)
	MarkOTPAsUsed(email string, otp string) error
	CleanupExpiredOTPs() error
}

type ConsumerOTPRepo struct {
	DB *sql.DB
}

func NewConsumerOTPRepo(db *sql.DB) *ConsumerOTPRepo {
	return &ConsumerOTPRepo{DB: db}
}

// StoreOTP stores a new OTP for the given email
func (r *ConsumerOTPRepo) StoreOTP(email string, otp string, expiresAt time.Time) error {
	// First, mark any existing unused OTPs for this email as used
	_, err := r.DB.Exec(`
		UPDATE consumer_otps 
		SET used = TRUE 
		WHERE email = $1 AND used = FALSE
	`, email)
	if err != nil {
		return fmt.Errorf("failed to invalidate existing OTPs: %w", err)
	}

	// Insert new OTP
	query := `
		INSERT INTO consumer_otps (email, otp, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err = r.DB.Exec(query, email, otp, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	return nil
}

// ValidateOTP checks if the provided OTP is valid for the given email
func (r *ConsumerOTPRepo) ValidateOTP(email string, otp string) (bool, error) {
	query := `
		SELECT id, expires_at, used
		FROM consumer_otps
		WHERE email = $1 AND otp = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var id int
	var expiresAt time.Time
	var used bool

	err := r.DB.QueryRow(query, email, otp).Scan(&id, &expiresAt, &used)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to validate OTP: %w", err)
	}

	// Check if OTP is expired
	if time.Now().After(expiresAt) {
		return false, nil
	}

	// Check if OTP is already used
	if used {
		return false, nil
	}

	return true, nil
}

// MarkOTPAsUsed marks an OTP as used
func (r *ConsumerOTPRepo) MarkOTPAsUsed(email string, otp string) error {
	query := `
		UPDATE consumer_otps
		SET used = TRUE
		WHERE email = $1 AND otp = $2
	`
	result, err := r.DB.Exec(query, email, otp)
	if err != nil {
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("OTP not found")
	}

	return nil
}

// CleanupExpiredOTPs removes expired OTPs (optional cleanup function)
func (r *ConsumerOTPRepo) CleanupExpiredOTPs() error {
	query := `
		DELETE FROM consumer_otps
		WHERE expires_at < NOW() OR used = TRUE
	`
	_, err := r.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired OTPs: %w", err)
	}

	return nil
}

