package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewUsersRepo(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	repo := NewUsersRepo(db)
	if repo == nil {
		t.Fatal("NewUsersRepo returned nil")
	}
	if repo.DB != db {
		t.Error("NewUsersRepo did not set DB correctly")
	}
}

func TestUsersRepo_GetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUsersRepo(db)

	tests := []struct {
		name          string
		email         string
		setupMock     func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:  "successful retrieval",
			email: "test@example.com",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "name", "profile_picture_url"}).
					AddRow(1, "test@example.com", "Test User", "https://example.com/pfp.jpg")
				mock.ExpectQuery("SELECT id, email, name, profile_picture_url").
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedUser: &models.User{
				Id:      1,
				Email:   "test@example.com",
				Name:    "Test User",
				Pfp_url: "https://example.com/pfp.jpg",
			},
			expectedError: nil,
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, email, name, profile_picture_url").
					WithArgs("notfound@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  &models.User{},
			expectedError: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			user, err := repo.GetUserByEmail(tt.email)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if user.Email != tt.expectedUser.Email {
					t.Errorf("Expected email %s, got %s", tt.expectedUser.Email, user.Email)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Mock expectations were not met: %v", err)
			}
		})
	}
}

func TestUsersRepo_UpsertUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUsersRepo(db)

	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		Pfp_url: "https://example.com/pfp.jpg",
	}

	t.Run("successful upsert", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO USERS").
			WithArgs(user.Email, user.Name, user.Pfp_url).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpsertUser(user)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Mock expectations were not met: %v", err)
		}
	})
}
