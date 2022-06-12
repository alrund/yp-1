package storage

import (
	"sync"

	tkn "github.com/alrund/yp-1/internal/app/token"
)

type composite struct {
	token *tkn.Token
	url   string
}

type Map struct {
	url2tokenValue       map[string]string
	tokenValue2composite map[string]*composite
	mx                   sync.RWMutex
}

func NewMap() *Map {
	return &Map{
		url2tokenValue:       make(map[string]string),
		tokenValue2composite: make(map[string]*composite),
	}
}

func (s *Map) Set(url string, token *tkn.Token) error {
	s.mx.Lock()
	s.url2tokenValue[url] = token.Value
	s.tokenValue2composite[token.Value] = &composite{token, url}
	s.mx.Unlock()
	return nil
}

func (s *Map) GetToken(tokenValue string) (*tkn.Token, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if value, ok := s.tokenValue2composite[tokenValue]; ok {
		return value.token, nil
	}
	return nil, ErrTokenNotFound
}

func (s *Map) GetTokenByURL(url string) (*tkn.Token, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if tokenValue, ok := s.url2tokenValue[url]; ok {
		return s.GetToken(tokenValue)
	}
	return nil, ErrTokenNotFound
}

func (s *Map) GetURL(tokenValue string) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if value, ok := s.tokenValue2composite[tokenValue]; ok {
		return value.url, nil
	}
	return "", ErrURLNotFound
}

func (s *Map) HasURL(url string) (bool, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if _, ok := s.url2tokenValue[url]; !ok {
		return false, nil
	}
	return true, nil
}

func (s *Map) HasToken(tokenValue string) (bool, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if _, ok := s.tokenValue2composite[tokenValue]; !ok {
		return false, nil
	}
	return true, nil
}
