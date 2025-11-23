package consumer_otp

import (
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
	"time"
)

// RequestOTP handles OTP request
func RequestOTP(
	otpService *services.ConsumerOTPService,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
		}

		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		if req.Email == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Email is required"}, http.StatusBadRequest, nil)
			return
		}

		err := otpService.SendOTP(req.Email)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to send OTP"}, http.StatusInternalServerError, nil)
			return
		}

		writeJSON(w, jsonutils.Envelope{"message": "OTP sent successfully"}, http.StatusOK, nil)
	}
}

// VerifyOTP handles OTP verification and creates JWT if valid
func VerifyOTP(
	otpService *services.ConsumerOTPService,
	authService *services.AuthService,
	writeJSON jsonutils.JSONwriter,
	readJSON jsonutils.JSONreader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
			OTP   string `json:"otp"`
		}

		if err := readJSON(w, r, &req); err != nil {
			writeJSON(w, jsonutils.Envelope{"error": err.Error()}, http.StatusBadRequest, nil)
			return
		}

		if req.Email == "" || req.OTP == "" {
			writeJSON(w, jsonutils.Envelope{"error": "Email and OTP are required"}, http.StatusBadRequest, nil)
			return
		}

		// Verify OTP
		valid, err := otpService.VerifyOTP(req.Email, req.OTP)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to verify OTP"}, http.StatusInternalServerError, nil)
			return
		}

		if !valid {
			writeJSON(w, jsonutils.Envelope{"error": "Invalid or expired OTP"}, http.StatusUnauthorized, nil)
			return
		}

		// Get or create user
		user, err := authService.GetUserByEmail(req.Email)
		if err != nil {
			// User doesn't exist, create one
			err = authService.UpsertUser(req.Email, "", "")
			if err != nil {
				writeJSON(w, jsonutils.Envelope{"error": "Failed to create user"}, http.StatusInternalServerError, nil)
				return
			}
			// Fetch the newly created user
			user, err = authService.GetUserByEmail(req.Email)
			if err != nil {
				writeJSON(w, jsonutils.Envelope{"error": "Failed to fetch user"}, http.StatusInternalServerError, nil)
				return
			}
		}

		// Create JWT
		jwtString, err := authService.CreateJWT(user)
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to create session"}, http.StatusInternalServerError, nil)
			return
		}

		// Set JWT cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    jwtString,
			Expires:  time.Now().Add(7 * 24 * time.Hour), // 7 days
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		writeJSON(w, jsonutils.Envelope{
			"message": "OTP verified successfully",
			"user":    user,
		}, http.StatusOK, nil)
	}
}

