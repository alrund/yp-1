package storage

import (
	tkn "github.com/alrund/yp-1/internal/app/token"
)

type composite struct {
	Token  *tkn.Token
	Url    string
	UserId string
}
