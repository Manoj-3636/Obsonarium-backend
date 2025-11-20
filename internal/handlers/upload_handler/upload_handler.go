package upload_handler

import (
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"net/http"
	"strings"
)

func UploadProductImage(uploadService *services.UploadService, writeJSON jsonutils.JSONwriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form with 10MB max memory
		err := r.ParseMultipartForm(10 << 20) // 10MB
		if err != nil {
			writeJSON(w, jsonutils.Envelope{"error": "Failed to parse multipart form"}, http.StatusBadRequest, nil)
			return
		}

		// Retrieve file from form
		file, header, err := r.FormFile("image")
		if err != nil {
			if err == http.ErrMissingFile {
				writeJSON(w, jsonutils.Envelope{"error": "No file provided. Use 'image' as the form field name"}, http.StatusBadRequest, nil)
				return
			}
			writeJSON(w, jsonutils.Envelope{"error": "Failed to retrieve file"}, http.StatusBadRequest, nil)
			return
		}
		defer file.Close()

		// Validate file size (check header size)
		if header.Size > 10*1024*1024 {
			writeJSON(w, jsonutils.Envelope{"error": "File size exceeds maximum allowed size of 10MB"}, http.StatusRequestEntityTooLarge, nil)
			return
		}

		// Save file using upload service
		url, err := uploadService.SaveProductImage(file, header)
		if err != nil {
			// Check error type to return appropriate status code
			errMsg := err.Error()
			if strings.Contains(errMsg, "invalid file extension") {
				writeJSON(w, jsonutils.Envelope{"error": errMsg}, http.StatusBadRequest, nil)
				return
			}
			if strings.Contains(errMsg, "file size exceeds") {
				writeJSON(w, jsonutils.Envelope{"error": errMsg}, http.StatusRequestEntityTooLarge, nil)
				return
			}
			// Default to 500 for save failures
			writeJSON(w, jsonutils.Envelope{"error": "Failed to save file"}, http.StatusInternalServerError, nil)
			return
		}

		// Return success response with URL
		writeJSON(w, jsonutils.Envelope{"url": url}, http.StatusOK, nil)
	}
}
