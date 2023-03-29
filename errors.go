package solauth

import "errors"

// Predefined errors
var (
	ErrUnauthorized = errors.New("Missing or invalid access token")
)
