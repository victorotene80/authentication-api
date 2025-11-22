package httputil

import "net/http"


func getErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrCodeBadRequest
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	case http.StatusUnprocessableEntity:
		return ErrCodeUnprocessableEntity
	case http.StatusTooManyRequests:
		return ErrCodeTooManyRequests
	case http.StatusInternalServerError:
		return ErrCodeInternalServerError
	case http.StatusServiceUnavailable:
		return ErrCodeServiceUnavailable
	default:
		return ErrCodeUnknownError
	}
}
