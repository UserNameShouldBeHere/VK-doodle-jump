package errors

import (
	"errors"
	"net/http"
)

// common errors
var (
	ErrInternal   = errors.New("internal or unknown error")
	ErrBadRequest = errors.New("bad request")
)

func ConvertToHttpErr(err error) int {
	switch {
	case errors.Is(err, ErrBadRequest),
		errors.Is(err, ErrRecievedFromApi):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
