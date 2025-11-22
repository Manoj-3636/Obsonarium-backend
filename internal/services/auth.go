package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrSelfTokenCreate  error = errors.New("internal error while creating jwt")
	ErrSelfTokenVerify  error = errors.New("internal error occured when trying to verify jwt") //hopefully never shows up
	ErrSelfTokenExpired error = errors.New("self token expired")
	ErrClaimsParse      error = errors.New("error occured when trying to parse jwt claims")
	ErrNoUserFound      error = errors.New("user with id is not found")
	ErrIntDatabase      error = errors.New("internal database error")
)

type AuthService struct {
	selfSigningKey  string
	usersRepo       repositories.IUsersRepo
	retailersRepo   repositories.IRetailersRepo
	wholesalersRepo repositories.IWholesalersRepo
}

func NewAuthService(usersRepo repositories.IUsersRepo, retailersRepo repositories.IRetailersRepo, wholesalersRepo repositories.IWholesalersRepo) *AuthService {
	return &AuthService{
		selfSigningKey:  os.Getenv("LOCOSYNC_SIGNING"),
		usersRepo:       usersRepo,
		retailersRepo:   retailersRepo,
		wholesalersRepo: wholesalersRepo,
	}
}

func (authService *AuthService) CreateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.Email,
		"role": "consumer",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(), // expires in 7 days
		"iat":  time.Now().Unix(),                         // issued at
		"iss":  "Obsonarium",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	self_token, err := token.SignedString([]byte(authService.selfSigningKey))

	if err != nil {
		return "", fmt.Errorf("%w,%w", ErrSelfTokenCreate, err)
	}

	return self_token, nil
}

func (authService *AuthService) VerifySelfToken(selfToken string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(selfToken, func(token *jwt.Token) (interface{}, error) {
		// Validate signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(authService.selfSigningKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrSelfTokenVerify, err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, ErrSelfTokenExpired
			}
		}
		return &claims, nil
	}

	return nil, ErrSelfTokenVerify

}

func (authService *AuthService) ClaimsToUser(claims *jwt.MapClaims) (*models.User, error) {
	email := (*claims)["sub"].(string)

	user, err := authService.usersRepo.GetUserByEmail(email)

	if err != nil {
		return &models.User{}, err
	}

	return user, fmt.Errorf("%w:%w", ErrClaimsParse, err)
}

func (authService *AuthService) GetUserByEmail(email string) (*models.User, error) {
	return authService.usersRepo.GetUserByEmail(email)
}

func (authService *AuthService) UpsertUser(email, name, pfp_url string) error {
	// For now, it just calls the repository.
	// In the future, you could add logic here like:
	// - Validating the email format
	// - Sanitizing the 'name' input
	// - Logging the creation of a new user

	recievedUser := models.User{
		Email:   email,
		Name:    name,
		Pfp_url: pfp_url,
	}
	err := authService.usersRepo.UpsertUser(&recievedUser)

	if err != nil {
		// You can also wrap the error in a service-level error
		return fmt.Errorf("service error finding or creating user: %w", err)
	}

	return nil
}

func (authService *AuthService) UpsertRetailer(email, name string) error {
	retailer := models.Retailer{
		Email: email,
		Name:  name,
	}
	err := authService.retailersRepo.UpsertRetailer(&retailer)

	if err != nil {
		return fmt.Errorf("service error finding or creating retailer: %w", err)
	}

	return nil
}

func (authService *AuthService) CreateRetailerJWT(retailer *models.Retailer) (string, error) {
	claims := jwt.MapClaims{
		"sub":  retailer.Email,
		"role": "retailer",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(), // expires in 7 days
		"iat":  time.Now().Unix(),                         // issued at
		"iss":  "Obsonarium",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	self_token, err := token.SignedString([]byte(authService.selfSigningKey))

	if err != nil {
		return "", fmt.Errorf("%w,%w", ErrSelfTokenCreate, err)
	}

	return self_token, nil
}

func (authService *AuthService) UpsertWholesaler(email, name string) error {
	wholesaler := models.Wholesaler{
		Email: email,
		Name:  name,
	}
	err := authService.wholesalersRepo.UpsertWholesaler(&wholesaler)

	if err != nil {
		return fmt.Errorf("service error finding or creating wholesaler: %w", err)
	}

	return nil
}

func (authService *AuthService) CreateWholesalerJWT(wholesaler *models.Wholesaler) (string, error) {
	claims := jwt.MapClaims{
		"sub":  wholesaler.Email,
		"role": "wholesaler",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(), // expires in 7 days
		"iat":  time.Now().Unix(),                         // issued at
		"iss":  "Obsonarium",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	self_token, err := token.SignedString([]byte(authService.selfSigningKey))

	if err != nil {
		return "", fmt.Errorf("%w,%w", ErrSelfTokenCreate, err)
	}

	return self_token, nil
}
