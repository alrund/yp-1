package app

import (
	"strings"

	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
)

type Storage interface {
	Set(userID, url string, token *tkn.Token) error
	GetToken(tokenValue string) (*tkn.Token, error)
	GetTokenByURL(url string) (*tkn.Token, error)
	GetTokenByUserID(userID string) (*tkn.Token, error)
	GetURL(tokenValue string) (string, error)
	HasURL(url string) (bool, error)
	HasToken(tokenValue string) (bool, error)
}

type URLShortener struct {
	Config *config.Config
	Storage
	TokenGenerator tkn.Generator
}

func (us *URLShortener) GetServerAddress() string {
	return us.Config.ServerAddress
}

func (us *URLShortener) GetBaseURL() string {
	return strings.TrimRight(us.Config.BaseURL, "/") + "/"
}

func (us *URLShortener) Add(userID, url string) (*tkn.Token, error) {
	ok, err := us.HasURL(url)
	if err != nil {
		return nil, err
	}
	if ok {
		token, err := us.GetTokenByURL(url)
		if err != nil {
			return nil, err
		}
		if token.IsExpired() {
			err = us.Set(userID, url, token.Refresh())
			if err != nil {
				return nil, err
			}
		}
		return token, nil
	}
	token := tkn.NewToken(us.TokenGenerator)
	err = us.Set(userID, url, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (us *URLShortener) Get(tokenValue string) (string, error) {
	ok, err := us.HasToken(tokenValue)
	if err != nil {
		return "", err
	} else if ok {
		token, err := us.GetToken(tokenValue)
		if err != nil {
			return "", err
		}
		if token.IsExpired() {
			return "", tkn.ErrTokenExpiredError
		}
		return us.GetURL(tokenValue)
	}
	return "", storage.ErrTokenNotFound
}
