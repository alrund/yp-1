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

// Generator generates the token value.
type Generator interface {
	Generate() (string, error)
}

// Token shortened URL token.
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

// IsExpired checks the token expire date.
func (t *Token) IsExpired() bool {
	return t.Expire.Before(time.Now())
}

// Refresh updates the token expire date.
func (t *Token) Refresh() *Token {
	t.Expire = time.Now().Add(LifeTime)
	return t
}
