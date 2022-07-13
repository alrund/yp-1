package token

import (
	"errors"
	"time"
)

var (
	ErrTokenExpiredError = errors.New("the token time is up")
	ErrTokenRemovedError = errors.New("the token has been removed")
)

const LifeTime = 24 * time.Hour

type Generator interface {
	Generate() (string, error)
}

type Token struct {
	Value   string
	Expire  time.Time
	Removed bool
}

func NewToken(g Generator) (*Token, error) {
	val, err := g.Generate()
	if err != nil {
		return nil, err
	}
	return &Token{
		Value:  val,
		Expire: time.Now().Add(LifeTime),
	}, nil
}

func (t *Token) IsExpired() bool {
	return t.Expire.Before(time.Now())
}

func (t *Token) Refresh() *Token {
	t.Expire = time.Now().Add(LifeTime)
	return t
}
