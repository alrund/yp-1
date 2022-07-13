package storage

import "errors"

var (
	ErrURLAlreadyExists = errors.New("url already exists")
	ErrURLNotFound      = errors.New("url not found")
	ErrTokenNotFound    = errors.New("token not found")
)
