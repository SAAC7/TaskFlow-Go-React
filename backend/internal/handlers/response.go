// Package handlers is the HTTP layer: routes, JSON in/out and status codes.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxBodyBytes = 1 << 20 // 1 MB

// Every error of the API has this shape, so the frontend handles one format:
//
//	{"error": {"code": "validation_failed", "message": "...", "details": {"title": "is required"}}}
type errorBody struct {
	Error apiError `json:"error"`
}

// apiError is the content of the "error" field of every error response.
type apiError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// respondJSON writes data as a JSON response with the given HTTP status code
// and the standard JSON headers.func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
// respondError writes an error response using the single error format of the
// API: {"error": {"code", "message", "details"}}.
func respondError(w http.ResponseWriter, status int, code, message string, details map[string]string) {
	respondJSON(w, status, errorBody{Error: apiError{Code: code, Message: message, Details: details}})
}

// decodeJSON reads the body into dst, limiting its size (1 MB) and refusing
// trailing garbage after the JSON object. Unknown fields are ignored on purpose,
// so the frontend can send back a whole task (with id, createdAt...) in a PUT.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

// respondDecodeError turns a decodeJSON error into a friendly 4xx response.
func respondDecodeError(w http.ResponseWriter, err error) {
	var (
		maxErr    *http.MaxBytesError
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
	)
	switch {
	case errors.As(err, &maxErr):
		respondError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body is too large", nil)
	case errors.Is(err, io.EOF):
		respondError(w, http.StatusBadRequest, "empty_body", "Request body must not be empty", nil)
	case errors.As(err, &syntaxErr):
		respondError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON", nil)
	case errors.As(err, &typeErr):
		respondError(w, http.StatusBadRequest, "invalid_json", fmt.Sprintf("Field %q has the wrong type", typeErr.Field), nil)
	default:
		respondError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
	}
}
