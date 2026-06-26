package errors

import "net/http"

type AppError struct {
	StatusCode int    `json:"_"` // we dont want to expose this to json client
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e AppError) Error() string {
	return e.Message
}

func NewAppError(statusCode int, code string, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// define the error standard that most of the application will use
var (
	ErrInternalServerError = NewAppError(http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred. Please try again later.")
	ErrBadRequest          = NewAppError(http.StatusBadRequest, "bad_request", "The request was invalid or cannot be served. Please check the request and try again.")
	ErrNotFound            = NewAppError(http.StatusNotFound, "not_found", "The requested resource was not found.")
	ErrUnauthorized        = NewAppError(http.StatusUnauthorized, "unauthorized", "You are not authorized to access this resource.")
)
