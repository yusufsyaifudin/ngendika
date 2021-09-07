package storage

import "errors"

var (
	ErrValidation = errors.New("validation error")

	// String App
	ErrAppWrongClientID = errors.New("app client_id is in wrong format")
	ErrAppWrongName     = errors.New("app name is in wrong format")
)
