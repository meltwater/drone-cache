package common

import "errors"

var (
	// ErrNotImplemented is returned when a storage backend has not implemented an API yet.
	ErrNotImplemented = errors.New("not implemented")
)
