package utils

import "errors"

// Sentinel errors for use across repository implementations
var (
	ErrNotFound     = errors.New("record not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("record already exists")
)
