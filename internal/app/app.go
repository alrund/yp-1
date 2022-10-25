package app

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
)

// Storage data storage.
type Storage interface {
	Set(userID, url string, token *tkn.Token) error
	SetBatch(userID string, url2token map[string]*tkn.Token) error
	GetToken(tokenValue string) (*tkn.Token, error)
	GetTokenByURL(url string) (*tkn.Token, error)
	GetTokensByUserID(userID string) ([]*tkn.Token, error)
	GetURL(tokenValue string) (string, error)
	GetURLsByUserID(userID string) ([]storage.URLpairs, error)
	HasURL(url string) (bool, error)
	HasToken(tokenValue string) (bool, error)
	Ping(ctx context.Context) error
	RemoveTokens(tokenValues []string, userID string) error
	GetURLCount() (int, error)
	GetUserIDCount() (int, error)
}

// URLShortener url shortening application.
type URLShortener struct {
	Config *config.Config
	Storage
	TokenGenerator tkn.Generator
	TrustedSubnet  *net.IPNet
}

type Stat struct {
	Urls, Users int
}

// GetConfig returns configuration data.
func (us *URLShortener) GetConfig() *config.Config {
	return us.Config
}

// GetServerAddress returns server address.
func (us *URLShortener) GetServerAddress() string {
	return us.Config.ServerAddress
}

// GetBaseURL returns base url.
func (us *URLShortener) GetBaseURL() string {
	return strings.TrimRight(us.Config.BaseURL, "/") + "/"
}

// Add adds a URL string to shorten.
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
		return token, storage.ErrURLAlreadyExists
	}
	token, err := tkn.NewToken(us.TokenGenerator)
	if err != nil {
		return nil, err
	}
	err = us.Set(userID, url, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// AddBatch adds multiple URLs at once for shortening.
func (us *URLShortener) AddBatch(userID string, urls []string) (map[string]*tkn.Token, error) {
	url2token := map[string]*tkn.Token{}
	url2newtoken := map[string]*tkn.Token{}
	var storageErr error

	for _, url := range urls {
		token, err := us.GetTokenByURL(url)
		if err != nil && !errors.Is(err, storage.ErrTokenNotFound) {
			return nil, err
		}
		if token != nil && token.IsExpired() {
			err = us.Set(userID, url, token.Refresh())
			if err != nil {
				return nil, err
			}
		}
		if token != nil {
			storageErr = storage.ErrURLAlreadyExists
		} else {
			token, err = tkn.NewToken(us.TokenGenerator)
			if err != nil {
				return nil, err
			}
			url2newtoken[url] = token
		}
		url2token[url] = token
	}

	err := us.SetBatch(userID, url2newtoken)
	if err != nil {
		return nil, err
	}

	return url2token, storageErr
}

// Get returns a URL by token.
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
		if token.Removed {
			return "", tkn.ErrTokenRemovedError
		}
		return us.GetURL(tokenValue)
	}
	return "", storage.ErrTokenNotFound
}

// GetUserURLs returns a URL by user ID.
func (us *URLShortener) GetUserURLs(userID string) ([]storage.URLpairs, error) {
	baseURL := us.GetBaseURL()
	URLPairs, err := us.GetURLsByUserID(userID)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(URLPairs); i++ {
		URLPairs[i].ShortURL = baseURL + URLPairs[i].ShortURL
	}

	return URLPairs, nil
}

// GetStats returns statistic.
func (us *URLShortener) GetStats() (*Stat, error) {
	urlCount, err := us.GetURLCount()
	if err != nil {
		return nil, err
	}
	userCount, err := us.GetUserIDCount()
	if err != nil {
		return nil, err
	}

	return &Stat{Urls: urlCount, Users: userCount}, nil
}
