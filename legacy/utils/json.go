package utils

import (
	"encoding/json"
	"net/http"
)

/**
 * decodes the request body into a model and validates it. Returns the model
 * and a boolean indicating if the model is valid
 */
func DecodeRequestBody[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var model T

	// check content type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return model, false
	}

	// decode body
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return model, false
	}

	// if model has a Validate method, call it
	if validator, ok := any(model).(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return model, false
		}
	}

	return model, true
}
