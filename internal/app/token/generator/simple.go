package generator

import (
	"math/rand"
	"strings"
)

const Length = 6
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type SimpleGenerator struct{}

func NewSimpleGenerator() *SimpleGenerator {
	return &SimpleGenerator{}
}

func (st *SimpleGenerator) Generate() string {
	rs := st.randomString(Length)
	return rs
}

func (st *SimpleGenerator) randomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}
