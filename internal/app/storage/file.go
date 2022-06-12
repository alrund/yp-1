package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	tkn "github.com/alrund/yp-1/internal/app/token"
)

type File struct {
	FileName string
	mx       sync.RWMutex
}

func NewFile(fileName string) (*File, error) {
	return &File{
		FileName: fileName,
	}, nil
}

func (s *File) Set(url string, token *tkn.Token) error {
	state, err := s.restoreState()
	if err != nil {
		return err
	}
	state[url] = token
	return s.saveState(state)
}

func (s *File) GetToken(tokenValue string) (*tkn.Token, error) {
	state, err := s.restoreState()
	if err != nil {
		return nil, err
	}

	for _, token := range state {
		if tokenValue == token.Value {
			return token, nil
		}
	}

	return nil, ErrTokenNotFound
}

func (s *File) GetTokenByURL(url string) (*tkn.Token, error) {
	state, err := s.restoreState()
	if err != nil {
		return nil, err
	}

	for u, token := range state {
		if u == url {
			return token, nil
		}
	}

	return nil, ErrTokenNotFound
}

func (s *File) GetURL(tokenValue string) (string, error) {
	state, err := s.restoreState()
	if err != nil {
		return "", err
	}

	for url, token := range state {
		if tokenValue == token.Value {
			return url, nil
		}
	}

	return "", ErrURLNotFound
}

func (s *File) HasURL(url string) (bool, error) {
	state, err := s.restoreState()
	if err != nil {
		return false, err
	}

	for u := range state {
		if u == url {
			return true, nil
		}
	}

	return false, nil
}

func (s *File) HasToken(tokenValue string) (bool, error) {
	state, err := s.restoreState()
	if err != nil {
		return false, err
	}

	for _, token := range state {
		if tokenValue == token.Value {
			return true, nil
		}
	}

	return false, nil
}

func (s *File) saveState(state map[string]*tkn.Token) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	file, err := os.OpenFile(s.FileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = file.Write(stateJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *File) restoreState() (map[string]*tkn.Token, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	state := make(map[string]*tkn.Token)

	if _, err := os.Stat(s.FileName); errors.Is(err, os.ErrNotExist) {
		return state, nil
	}

	stateJSON, err := os.ReadFile(s.FileName)
	if err != nil {
		return nil, err
	}

	if len(stateJSON) == 0 {
		return state, nil
	}

	err = json.Unmarshal(stateJSON, &state)
	if err != nil {
		return nil, err
	}

	return state, nil
}
