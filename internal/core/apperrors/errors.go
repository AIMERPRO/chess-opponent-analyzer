package apperrors

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("already exists")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
)
