package generator

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	Length  = 6
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Simple struct{}

func NewSimple() *Simple {
	return &Simple{}
}

func (st *Simple) Generate() (string, error) {
	rs, err := st.randomString(Length)
	if err != nil {
		return "", err
	}
	return rs, nil
}

func (st *Simple) randomString(n int) (string, error) {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(charset[num.Int64()])
	}
	return sb.String(), nil
}
