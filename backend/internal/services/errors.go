package services

import "errors"

var (
	ErrForbidden = errors.New("forbidden: insufficient permissions")
	ErrNotFound  = errors.New("not found")
)
