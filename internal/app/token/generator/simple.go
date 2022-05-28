package generator

import (
	"math/rand"
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

func (st *Simple) Generate() string {
	rs := st.randomString(Length)
	return rs
}

func (st *Simple) randomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}
