package storage

import (
	tkn "github.com/alrund/yp-1/internal/app/token"
)

type composite struct {
	Token  *tkn.Token
	URL    string
	UserID string
}

type URLpairs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
