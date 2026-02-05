package common

import "errors"

// Common domain errors used across features
var (
	// ErrNotFound is returned when a requested resource doesn't exist
	ErrNotFound = errors.New("not found")

	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user doesn't have permission
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidInput is returned when request validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicateEntry is returned when a unique constraint is violated
	ErrDuplicateEntry = errors.New("duplicate entry")

	// ErrQuotaExceeded is returned when a tenant exceeds their quota
	ErrQuotaExceeded = errors.New("quota exceeded")

	// ErrRateLimited is returned when rate limit is exceeded
	ErrRateLimited = errors.New("rate limited")

	// ErrInactive is returned when accessing an inactive resource
	ErrInactive = errors.New("inactive")
)
