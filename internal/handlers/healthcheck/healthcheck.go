package healthcheck

import (
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
)

func NewHealthCheckHandler(env string, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Status string
			Env    string
		}{
			Status: "online",
			Env:    env,
		}

		writeJSON(w, jsonutils.Envelope{"health": response}, http.StatusOK, nil)
	}
}
