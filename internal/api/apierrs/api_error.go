package apierrs

import (
	"net/http"
	"sekai/internal/entities/dto"
)

var (
	CodeNotFound     = "NOT_FOUND"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeBadRequest   = "BAD_REQUEST"
	CodeInternal     = "INTERNAL_SERVER_ERROR"
)

func ErrUnauthorized(message string) dto.APIError {
	return dto.APIError{
		Status:  http.StatusUnauthorized,
		Code:    CodeUnauthorized,
		Message: message,
	}
}

func ErrForbidden(message string) dto.APIError {
	return dto.APIError{
		Status:  http.StatusForbidden,
		Code:    CodeForbidden,
		Message: message,
	}
}

func ErrNotFound(message string) dto.APIError {
	return dto.APIError{
		Status:  http.StatusNotFound,
		Code:    CodeNotFound,
		Message: message,
	}
}

func ErrBadRequest(message string) dto.APIError {
	return dto.APIError{
		Status:  http.StatusBadRequest,
		Code:    CodeBadRequest,
		Message: message,
	}
}

func ErrInternalServerError() dto.APIError {
	return dto.APIError{
		Status: http.StatusInternalServerError,
		Code:   CodeInternal,
	}
}
