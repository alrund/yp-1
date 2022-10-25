package handler

import "github.com/alrund/yp-1/internal/app"

type Collection struct {
	us *app.URLShortener
}

func NewCollection(us *app.URLShortener) *Collection {
	return &Collection{
		us: us,
	}
}
