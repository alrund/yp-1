package handler

import (
	"context"
	"time"

	"github.com/alrund/yp-1/internal/app/storage"
	tkn "github.com/alrund/yp-1/internal/app/token"
)

type TestStorage struct{}

func (st *TestStorage) HasToken(tokenValue string) (bool, error) {
	switch tokenValue {
	case "qwerty":
		return true, nil
	case "expired":
		return true, nil
	case "removed":
		return true, nil
	}
	return false, nil
}

func (st *TestStorage) GetToken(tokenValue string) (*tkn.Token, error) {
	switch tokenValue {
	case "expired":
		return &tkn.Token{Value: "expired", Expire: time.Now().Add(-tkn.LifeTime)}, nil
	case "removed":
		return &tkn.Token{Value: "removed", Expire: time.Now().Add(tkn.LifeTime), Removed: true}, nil
	default:
		return &tkn.Token{Value: "qwerty", Expire: time.Now().Add(tkn.LifeTime)}, nil
	}
}

func (st *TestStorage) GetURLsByUserID(userID string) ([]storage.URLpairs, error) {
	switch userID {
	case "empty":
		return nil, storage.ErrTokenNotFound
	default:
		return []storage.URLpairs{
			{
				OriginalURL: "url",
				ShortURL:    "shorturl",
			},
		}, nil
	}
}

func (st *TestStorage) GetURL(string) (string, error)                                 { return "https://ya.ru", nil }
func (st *TestStorage) GetTokensByUserID(string) ([]*tkn.Token, error)                { return nil, nil }
func (st *TestStorage) GetTokenByURL(string) (*tkn.Token, error)                      { return nil, nil }
func (st *TestStorage) HasURL(string) (bool, error)                                   { return true, nil }
func (st *TestStorage) Set(string, string, *tkn.Token) error                          { return nil }
func (st *TestStorage) SetBatch(userID string, url2token map[string]*tkn.Token) error { return nil }
func (st *TestStorage) Ping(ctx context.Context) error                                { return nil }
func (st *TestStorage) RemoveTokens(tokenValues []string, userID string) error        { return nil }
func (st *TestStorage) GetURLCount() (int, error)                                     { return 2, nil }
func (st *TestStorage) GetUserIDCount() (int, error)                                  { return 3, nil }
