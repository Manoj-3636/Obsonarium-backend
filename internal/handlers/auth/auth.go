package auth

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/services"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog"
)

const (
	key    = "8e0f0a0e82854492d6a6b0f229dfd5f8e1ece132a97c122406d515900c8b32c5"
	MaxAge = 60 * 60
)

func NewAuth(logger zerolog.Logger, env string) {
	err := godotenv.Load()
	if err != nil {
		logger.Error().Err(err)
	}

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	// fmt.Println(googleClientId)
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	// fmt.Println(googleClientSecret)

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   MaxAge,
		HttpOnly: true,
		Secure:   false,                // because you're on localhost
		SameSite: http.SameSiteLaxMode, // <--- THIS FIXES THE SESSION ISSUE
	}

	//I dont know how the fuck this works please don't change it
	gothic.Store = store

	goth.UseProviders(
		// TOOD don't hardocode
		google.New(googleClientId, googleClientSecret, "http://localhost:5173/api/auth/google/callback", "email", "profile"),
	)
}

func NewAuthCallback(logger zerolog.Logger, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

		gothUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to complete Gothic auth")
			http.Error(w, "Authentication failed", http.StatusInternalServerError)
			return
		}

		receivedUser := models.User{
			Email:   gothUser.Email,
			Name:    gothUser.Name,
			Pfp_url: gothUser.AvatarURL,
		}

		err = authService.UpsertUser(gothUser.Email, gothUser.Name, gothUser.AvatarURL)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to find or create user")
			http.Error(w, "Failed to process user", http.StatusInternalServerError)
			return
		}

		jwtString, err := authService.CreateJWT(&receivedUser)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create JWT")
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		cookie := &http.Cookie{
			Name:     "jwt",
			Value:    jwtString,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "http://localhost:5173", http.StatusFound)
	}
}

func AuthLogout(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func AuthProvider(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	// Add the provider to the request context
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	// The 'else' block from your original function is all you need.
	// This handles redirecting the user to Google.
	gothic.BeginAuthHandler(w, r)
}
