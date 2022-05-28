package storage

import (
	"errors"
	"sync"

	tkn "github.com/alrund/yp-1/internal/app/token"
)

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrTokenNotFound = errors.New("token not found")
)

type MapStorage struct {
	tokens         map[string]*tkn.Token
	url2tokenValue map[string]string
	tokenValue2url map[string]string
	mx             sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		tokens:         make(map[string]*tkn.Token),
		url2tokenValue: make(map[string]string),
		tokenValue2url: make(map[string]string),
	}
}

func (s *MapStorage) Set(url string, token *tkn.Token) error {
	s.mx.Lock()
	s.tokens[token.Value] = token
	s.url2tokenValue[url] = token.Value
	s.tokenValue2url[token.Value] = url
	s.mx.Unlock()
	return nil
}

func (s *MapStorage) GetToken(tokenValue string) (*tkn.Token, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if value, ok := s.tokens[tokenValue]; ok {
		return value, nil
	}
	return nil, ErrTokenNotFound
}

func (s *MapStorage) GetTokenByURL(url string) (*tkn.Token, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if tokenValue, ok := s.url2tokenValue[url]; ok {
		return s.GetToken(tokenValue)
	}
	return nil, ErrTokenNotFound
}

func (s *MapStorage) GetURL(tokenValue string) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if value, ok := s.tokenValue2url[tokenValue]; ok {
		return value, nil
	}
	return "", ErrURLNotFound
}

func (s *MapStorage) HasURL(url string) (bool, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if _, ok := s.url2tokenValue[url]; !ok {
		return false, nil
	}
	return true, nil
}

func (s *MapStorage) HasToken(tokenValue string) (bool, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if _, ok := s.tokenValue2url[tokenValue]; !ok {
		return false, nil
	}
	return true, nil
}
