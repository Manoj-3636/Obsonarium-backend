package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog"
)


type ProviderKeyType string
const (
	key="8e0f0a0e82854492d6a6b0f229dfd5f8e1ece132a97c122406d515900c8b32c5"
	MaxAge = 60*5
)

const provideKey ProviderKeyType = "provide"

func NewAuth(logger zerolog.Logger, env string){
	err := godotenv.Load()
	if err != nil {
		logger.Error().Err(err)
	}

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = (env == "prod")
	gothic.Store = store

	goth.UseProviders(
		// TOOD don't hardocode
		google.New(googleClientId,googleClientSecret,"http://localhost:8000/auth/google/callback"),
	)
}

func NewAuthCallback(logger zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter,r *http.Request) {
		provider := chi.URLParam(r,"provider")
		
		r = r.WithContext(context.WithValue(r.Context(),provideKey,provider))
		
		user,err := gothic.CompleteUserAuth(w,r)

		if err!= nil {
			fmt.Fprintln(w,r)
			return 
		}

		fmt.Println(user)
		http.Redirect(w,r,"http://localhost:5173",http.StatusFound)
	}
}

func AuthLogout(res http.ResponseWriter, req *http.Request){
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func AuthProvider(w http.ResponseWriter, r *http.Request) {
    provider := chi.URLParam(r, "provider")

    // Add the provider to the request context
    // FIX: Use r.Context() as the parent, not context.Background()
    r = r.WithContext(context.WithValue(r.Context(), provideKey, provider))

    // The 'else' block from your original function is all you need.
    // This handles redirecting the user to Google.
    gothic.BeginAuthHandler(w, r)
}