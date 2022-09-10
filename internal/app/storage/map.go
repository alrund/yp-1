package storage

import (
	"context"
	"sync"

	tkn "github.com/alrund/yp-1/internal/app/token"
)

// Map hash map storage.
type Map struct {
	userID2tokenValue    map[string][]string
	url2tokenValue       map[string]string
	tokenValue2composite map[string]*composite
	mx                   sync.RWMutex
}

func NewMap() *Map {
	return &Map{
		userID2tokenValue:    make(map[string][]string),
		url2tokenValue:       make(map[string]string),
		tokenValue2composite: make(map[string]*composite),
	}
}

func (s *Map) Set(userID, url string, token *tkn.Token) error {
	s.mx.Lock()
	_, ok := s.userID2tokenValue[userID]
	if !ok {
		s.userID2tokenValue[userID] = []string{}
	}
	s.userID2tokenValue[userID] = append(s.userID2tokenValue[userID], token.Value)
	s.url2tokenValue[url] = token.Value
	s.tokenValue2composite[token.Value] = &composite{token, url, userID}
	s.mx.Unlock()
	return nil
}

func (s *Map) SetBatch(userID string, url2token map[string]*tkn.Token) error {
	for url, token := range url2token {
		err := s.Set(userID, url, token)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Map) RemoveTokens(tokenValues []string, userID string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, tokenValue := range tokenValues {
		userTokenValues, ok := s.userID2tokenValue[userID]
		if !ok {
			continue
		}

		has := false
		for _, userTokenValue := range userTokenValues {
			if userTokenValue == tokenValue {
				has = true
				break
			}
		}
		if !has {
			continue
		}

		composite, ok := s.tokenValue2composite[tokenValue]
		if !ok {
			continue
		}
		composite.Token.Removed = true
	}

	return nil
}

func (s *Map) GetToken(tokenValue string) (*tkn.Token, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if value, ok := s.tokenValue2composite[tokenValue]; ok {
		return value.Token, nil
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

func (s *Map) GetTokensByUserID(userID string) ([]*tkn.Token, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	tokens := make([]*tkn.Token, 0)
	if tokenValues, ok := s.userID2tokenValue[userID]; ok {
		for _, tokenValue := range tokenValues {
			token, err := s.GetToken(tokenValue)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		}
	}

	if len(tokens) > 0 {
		return tokens, nil
	}

	return nil, ErrTokenNotFound
}

func (s *Map) GetURL(tokenValue string) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if value, ok := s.tokenValue2composite[tokenValue]; ok {
		return value.URL, nil
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

func (s *Map) GetURLsByUserID(userID string) ([]URLpairs, error) {
	tokens, err := s.GetTokensByUserID(userID)
	if err != nil {
		return nil, err
	}

	urls := make([]URLpairs, 0)
	for _, token := range tokens {
		originalURL, err := s.GetURL(token.Value)
		if err != nil {
			return nil, err
		}
		urls = append(urls, URLpairs{ShortURL: token.Value, OriginalURL: originalURL})
	}

	if len(urls) > 0 {
		return urls, nil
	}

	return nil, ErrURLNotFound
}

func (s *Map) Ping(ctx context.Context) error {
	return nil
}
