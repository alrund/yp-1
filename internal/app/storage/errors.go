package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrTokenNotFound = errors.New("token not found")
)
