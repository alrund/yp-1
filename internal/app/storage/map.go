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

type composite struct {
	token *tkn.Token
	url   string
}

type MapStorage struct {
	url2tokenValue       map[string]string
	tokenValue2composite map[string]*composite
	mx                   sync.RWMutex
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		url2tokenValue:       make(map[string]string),
		tokenValue2composite: make(map[string]*composite),
	}
}

func (s *MapStorage) Set(url string, token *tkn.Token) error {
	s.mx.Lock()
	s.url2tokenValue[url] = token.Value
	s.tokenValue2composite[token.Value] = &composite{token, url}
	s.mx.Unlock()
	return nil
}

func (s *MapStorage) GetToken(tokenValue string) (*tkn.Token, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if value, ok := s.tokenValue2composite[tokenValue]; ok {
		return value.token, nil
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
	if value, ok := s.tokenValue2composite[tokenValue]; ok {
		return value.url, nil
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
	if _, ok := s.tokenValue2composite[tokenValue]; !ok {
		return false, nil
	}
	return true, nil
}
