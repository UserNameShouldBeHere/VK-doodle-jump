package errors

import "errors"

var (
	ErrFailedToCallApi = errors.New("failed to make api call")
	ErrRecievedFromApi = errors.New("error recieved from api")
	ErrStateMismatch   = errors.New("state mismatch")
)
