package healthcheck

import (
	jsonutils "Obsonarium-backend/internal/utils"
	"net/http"
)

func NewHealthCheckHandler(env string,writeJSON jsonutils.JSONwriter) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		response := struct{
			Status string
			Env string
		}{
			Status: "online",
			Env: env,
		}

		writeJSON(w,jsonutils.Envelope{"health":response},http.StatusOK,nil)
	}
}