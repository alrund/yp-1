package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	tkn "github.com/alrund/yp-1/internal/app/token"
)

// File storage.
type File struct {
	FileName string
	state    map[string]composite
	stateMx  sync.RWMutex
	mx       sync.RWMutex
}

func NewFile(fileName string) (*File, error) {
	file := &File{
		FileName: fileName,
		state:    make(map[string]composite),
	}

	if err := file.restoreState(); err != nil {
		return nil, err
	}

	return file, nil
}

func (s *File) Set(userID string, url string, token *tkn.Token) error {
	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	composite := s.state[url]
	composite.Token = token
	composite.URL = url
	composite.UserID = userID
	s.state[url] = composite

	return s.saveState()
}

func (s *File) SetBatch(userID string, url2token map[string]*tkn.Token) error {
	for url, token := range url2token {
		err := s.Set(userID, url, token)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *File) RemoveTokens(tokenValues []string, userID string) error {
	s.stateMx.RLock()
	defer s.stateMx.RUnlock()

	for _, composite := range s.state {
		if composite.Token == nil {
			continue
		}

		if composite.UserID != userID {
			continue
		}

		for _, tokenValue := range tokenValues {
			if tokenValue == composite.Token.Value {
				composite.Token.Removed = true
			}
		}
	}

	return s.saveState()
}

func (s *File) GetToken(tokenValue string) (*tkn.Token, error) {
	s.stateMx.RLock()
	defer s.stateMx.RUnlock()

	for _, composite := range s.state {
		if composite.Token == nil {
			return nil, ErrTokenNotFound
		}
		if tokenValue == composite.Token.Value {
			return composite.Token, nil
		}
	}

	return nil, ErrTokenNotFound
}

func (s *File) GetTokenByURL(url string) (*tkn.Token, error) {
	s.stateMx.RLock()
	defer s.stateMx.RUnlock()

	for u, composite := range s.state {
		if u == url {
			return composite.Token, nil
		}
	}

	return nil, ErrTokenNotFound
}

func (s *File) GetTokensByUserID(userID string) ([]*tkn.Token, error) {
	s.stateMx.RLock()
	defer s.stateMx.RUnlock()

	tokens := make([]*tkn.Token, 0)
	for _, composite := range s.state {
		if userID == composite.UserID {
			tokens = append(tokens, composite.Token)
		}
	}

	if len(tokens) > 0 {
		return tokens, nil
	}

	return nil, ErrTokenNotFound
}

func (s *File) GetURL(tokenValue string) (string, error) {
	s.stateMx.RLock()
	defer s.stateMx.RUnlock()

	for url, composite := range s.state {
		if composite.Token == nil {
			return "", ErrTokenNotFound
		}
		if tokenValue == composite.Token.Value {
			return url, nil
		}
	}

	return "", ErrURLNotFound
}

func (s *File) HasURL(url string) (bool, error) {
	s.stateMx.RLock()
	defer s.stateMx.RUnlock()

	for u := range s.state {
		if u == url {
			return true, nil
		}
	}

	return false, nil
}

func (s *File) HasToken(tokenValue string) (bool, error) {
	s.stateMx.RLock()
	defer s.stateMx.RUnlock()

	for _, composite := range s.state {
		if composite.Token == nil {
			return false, nil
		}
		if tokenValue == composite.Token.Value {
			return true, nil
		}
	}

	return false, nil
}

func (s *File) GetURLsByUserID(userID string) ([]URLpairs, error) {
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

func (s *File) saveState() error {
	s.mx.Lock()
	defer s.mx.Unlock()

	file, err := os.OpenFile(s.FileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	stateJSON, err := json.Marshal(s.state)
	if err != nil {
		return err
	}

	_, err = file.Write(stateJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *File) restoreState() error {
	s.mx.RLock()
	defer s.mx.RUnlock()

	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	state := make(map[string]composite)

	if _, err := os.Stat(s.FileName); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	stateJSON, err := os.ReadFile(s.FileName)
	if err != nil {
		return err
	}

	if len(stateJSON) == 0 {
		return nil
	}

	err = json.Unmarshal(stateJSON, &state)
	if err != nil {
		return err
	}

	s.state = state

	return nil
}

func (s *File) Ping(ctx context.Context) error {
	return nil
}
