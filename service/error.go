package service

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is returned when an error occurred
//
// swagger:response ErrorResponse
type errorResponseWrapper struct {
	// in:body
	Resp ErrorResponse
}

// ErrorResponse is returned when an error occurred
type ErrorResponse struct {
	// status code
	Code int `json:"code,omitempty"`
	// error message
	Message string `json:"message,omitempty"`
}

// Error replies to the request with JSON formatted ErrorResponse and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w. The error message should be plain text.
func Error(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{Code: code, Message: msg})
}
