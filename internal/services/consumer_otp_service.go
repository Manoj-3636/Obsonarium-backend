package services

import (
	"Obsonarium-backend/internal/repositories"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type ConsumerOTPService struct {
	OTPRepo     repositories.IConsumerOTPRepo
	EmailService *EmailService
	UsersRepo   repositories.IUsersRepo
}

func NewConsumerOTPService(otpRepo repositories.IConsumerOTPRepo, emailService *EmailService, usersRepo repositories.IUsersRepo) *ConsumerOTPService {
	return &ConsumerOTPService{
		OTPRepo:      otpRepo,
		EmailService: emailService,
		UsersRepo:    usersRepo,
	}
}

// GenerateOTP generates a random 6-digit OTP
func (s *ConsumerOTPService) GenerateOTP() (string, error) {
	otp := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate OTP: %w", err)
		}
		otp += num.String()
	}
	return otp, nil
}

// SendOTP generates and sends an OTP to the given email
func (s *ConsumerOTPService) SendOTP(email string) error {
	// Generate OTP
	otp, err := s.GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Set expiration time (10 minutes from now)
	expiresAt := time.Now().Add(10 * time.Minute)

	// Store OTP in database
	err = s.OTPRepo.StoreOTP(email, otp, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send email with OTP
	subject := "Your Obsonarium Login OTP"
	body := fmt.Sprintf("Your login OTP is: %s\n\nThis OTP will expire in 10 minutes.\n\nIf you didn't request this OTP, please ignore this email.", otp)

	err = s.EmailService.SendEmail(email, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

// VerifyOTP verifies the OTP for the given email
func (s *ConsumerOTPService) VerifyOTP(email string, otp string) (bool, error) {
	// Validate OTP
	valid, err := s.OTPRepo.ValidateOTP(email, otp)
	if err != nil {
		return false, fmt.Errorf("failed to validate OTP: %w", err)
	}

	if !valid {
		return false, nil
	}

	// Mark OTP as used
	err = s.OTPRepo.MarkOTPAsUsed(email, otp)
	if err != nil {
		return false, fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	return true, nil
}

