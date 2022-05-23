package app

import (
	"errors"
	stg "github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
)

var ErrTokenExpiredError = errors.New("the token time is up")

type Storage interface {
	Set(url string, token tkn.Token) error
	GetToken(tokenValue string) (tkn.Token, error)
	GetTokenByUrl(url string) (tkn.Token, error)
	GetUrl(tokenValue string) (string, error)
	HasUrl(url string) (bool, error)
	HasToken(tokenValue string) (bool, error)
}

type UrlShortener struct {
	Storage
}

func (us *UrlShortener) Add(url string) (*tkn.Token, error) {
	ok, err := us.HasUrl(url)
	if err != nil {
		return nil, err
	} else if ok {
		token, err := us.GetTokenByUrl(url)
		if err != nil {
			return nil, err
		}
		if token.IsExpired() {
			err = us.Set(url, *token.Refresh())
			if err != nil {
				return nil, err
			}
		}
		return &token, nil
	}
	token := tkn.NewToken(new(tkn.SimpleGenerator))
	err = us.Set(url, *token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (us *UrlShortener) Get(tokenValue string) (string, error) {
	ok, err := us.HasToken(tokenValue)
	if err != nil {
		return "", err
	} else if ok {
		token, err := us.GetToken(tokenValue)
		if err != nil {
			return "", err
		}
		if token.IsExpired() {
			return "", ErrTokenExpiredError
		}
		return us.GetUrl(tokenValue)
	}
	return "", stg.ErrTokenNotFound
}
