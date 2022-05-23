package storage

import (
	"errors"
	tkn "github.com/alrund/yp-1/internal/app/token"
)

var ErrUrlNotFound = errors.New("url not found")
var ErrTokenNotFound = errors.New("token not found")

type MapStorage struct {
	tokens         map[string]tkn.Token
	url2tokenValue map[string]string
	tokenValue2url map[string]string
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		make(map[string]tkn.Token),
		make(map[string]string),
		make(map[string]string),
	}
}

func (s *MapStorage) Set(url string, token tkn.Token) error {
	s.tokens[token.Value] = token
	s.url2tokenValue[url] = token.Value
	s.tokenValue2url[token.Value] = url
	return nil
}

func (s *MapStorage) GetToken(tokenValue string) (tkn.Token, error) {
	if value, ok := s.tokens[tokenValue]; ok {
		return value, nil
	}
	var token tkn.Token
	return token, ErrTokenNotFound
}

func (s *MapStorage) GetTokenByUrl(url string) (tkn.Token, error) {
	if tokenValue, ok := s.url2tokenValue[url]; ok {
		return s.GetToken(tokenValue)
	}
	var token tkn.Token
	return token, ErrTokenNotFound
}

func (s *MapStorage) GetUrl(tokenValue string) (string, error) {
	if value, ok := s.tokenValue2url[tokenValue]; ok {
		return value, nil
	}
	return "", ErrUrlNotFound
}

func (s *MapStorage) HasUrl(url string) (bool, error) {
	if _, ok := s.url2tokenValue[url]; !ok {
		return false, nil
	}
	return true, nil
}

func (s *MapStorage) HasToken(tokenValue string) (bool, error) {
	if _, ok := s.tokenValue2url[tokenValue]; !ok {
		return false, nil
	}
	return true, nil
}
