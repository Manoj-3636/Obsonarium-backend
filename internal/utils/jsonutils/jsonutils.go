package jsonutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Envelope map[string]interface{}

type JSONwriter func(http.ResponseWriter, Envelope, int, http.Header) error
type JSONreader func(http.ResponseWriter, *http.Request, interface{}) error

type JSONutils struct {
	Writer JSONwriter
	Reader JSONreader
}

func NewJSONutils() JSONutils {
	return JSONutils{
		Writer: WriteJSON,
		Reader: readJSON,
	}
}

// writeJSON writes a json response with configurable status and headers
func WriteJSON(w http.ResponseWriter, data Envelope, status int, headers http.Header) error {
	// http.Header has the type map[string][]string\

	json, err := json.Marshal(data)
	if err != nil {
		return err
	}

	json = append(json, '\n') //for easier viewing with curl

	for key, value := range headers {
		(w).Header()[key] = value
	}

	(w).Header().Set("Content-Type", "application/json")
	(w).WriteHeader(status)
	(w).Write(json)
	return nil
}

// ReadJSON reads a json request, given a destination
func readJSON(_ http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Decode the request body into the target destination.
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	return nil
}
