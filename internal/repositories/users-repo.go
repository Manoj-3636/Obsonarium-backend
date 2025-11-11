package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")

type IUsersRepo interface {
	AddNewUser(*models.User) error
	GetUserByEmail(string) (*models.User, error)
	CreateOrGet(string,string,string) (*models.User, error)
}

type UsersRepo struct {
	DB *sql.DB
}

func NewUsersRepo(db *sql.DB) *UsersRepo {
	return &UsersRepo{DB: db}
}

func (repo *UsersRepo) AddNewUser(user *models.User) error {
	query := `
		INSERT INTO USERS (email,name,profile_picture_url)
		VALUES ($1,$2,$3)
	`

	_, err := repo.DB.Exec(
		query,
		user.Email,
		user.Name,
		user.Pfp_url,
	)

	return err
}

func (repo *UsersRepo) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, name, profile_picture_url 
		FROM users 
		WHERE email = $1`

	var user models.User

	row := repo.DB.QueryRow(query, email)

	err := row.Scan(
		&user.Id,
		&user.Email,
		&user.Name,
		&user.Pfp_url,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.User{}, ErrUserNotFound
		}
		return &models.User{}, err
	}

	return &user, nil
}

func (repo *UsersRepo) CreateOrGet(email, name, pfp_url string) (*models.User, error) {
	// 1. Try to find the user by their email first.
	// (This now uses the .Scan() version from above)
	foundUser, err := repo.GetUserByEmail(email)

	if err == nil {
		// --- User Was Found ---
		return foundUser, nil
	}

	// 2. Check if the error was a *real* database problem.
	if !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("error checking for user: %w", err)
	}

	// --- User Was Not Found ---
	// 3. Create the new user.
	// We MUST explicitly list the columns in RETURNING to match our Scan.
	createQuery := `
		INSERT INTO users (email, name, profile_picture_url)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, profile_picture_url` // No more "RETURNING *"

	var newUser models.User

	// 4. Execute the insert, get the row object, and use Scan.
	// This is the change you requested.
	row := repo.DB.QueryRow(createQuery, email, name, pfp_url)
	err = row.Scan(
		&newUser.Id,
		&newUser.Email,
		&newUser.Name,
		&newUser.Pfp_url,
	)

	if err != nil {
		// The INSERT itself failed (e.g., duplicate email from race condition).
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// 5. Success! Return the newly created user.
	return &newUser, nil
}
