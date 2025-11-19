package auth

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
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

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const UserEmailKey ContextKey = "user_email"

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

	consumerProvider := google.New(googleClientId, googleClientSecret, "http://localhost:5173/api/auth/google/callback", "email", "profile")
	
	retailerProvider := google.New(googleClientId, googleClientSecret, "http://localhost:5174/api/auth/google-retailer/callback", "email", "profile")
	retailerProvider.SetName("google-retailer")

	goth.UseProviders(
		consumerProvider,
		retailerProvider,
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

		if provider == "google-retailer" {
			retailer := models.Retailer{
				Email: gothUser.Email,
				Name:  gothUser.Name,
			}

			err = authService.UpsertRetailer(gothUser.Email, gothUser.Name)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to find or create retailer")
				http.Error(w, "Failed to process retailer", http.StatusInternalServerError)
				return
			}

			jwtString, err := authService.CreateRetailerJWT(&retailer)
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

			// Redirect to retailer home
			http.Redirect(w, r, "http://localhost:5174", http.StatusFound)
			return
		}

		// Default to consumer logic
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

		// Redirect to home - frontend will check sessionStorage and redirect if needed
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

// RequireAuth is a middleware that checks authentication status and adds user email to context
// If authentication fails, it returns 401 Unauthorized and stops the request
func RequireAuth(authService *services.AuthService, logger zerolog.Logger, writeJSON jsonutils.JSONwriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract JWT token from cookie
			cookie, err := r.Cookie("jwt")
			if err != nil {
				// No cookie found, return 401 Unauthorized
				logger.Debug().Msg("No JWT cookie found")
				writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
				return
			}

			// Verify the token
			claims, err := authService.VerifySelfToken(cookie.Value)
			if err != nil {
				// Token invalid or expired, return 401 Unauthorized
				logger.Debug().Err(err).Msg("JWT verification failed")
				writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
				return
			}

			// Extract email from claims
			email, ok := (*claims)["sub"].(string)
			if !ok || email == "" {
				// Invalid claims, return 401 Unauthorized
				logger.Debug().Msg("Invalid email in JWT claims")
				writeJSON(w, jsonutils.Envelope{"error": "Unauthorized"}, http.StatusUnauthorized, nil)
				return
			}

			// Add email to request context and continue
			ctx := context.WithValue(r.Context(), UserEmailKey, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserEmailFromContext extracts the user email from the request context
// Returns empty string if not found
func GetUserEmailFromContext(r *http.Request) string {
	email, ok := r.Context().Value(UserEmailKey).(string)
	if !ok {
		return ""
	}
	return email
}
