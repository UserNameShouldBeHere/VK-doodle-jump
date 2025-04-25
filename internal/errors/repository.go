package errors

import "errors"

var (
	ErrTarantoolExec   = errors.New("failed to exec command")
	ErrTarantoolDecode = errors.New("failed to decode response")
)
