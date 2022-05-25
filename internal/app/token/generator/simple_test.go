package generator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerate(t *testing.T) {
	sg := &SimpleGenerator{}
	val := sg.Generate()
	assert.Len(t, val, Length)
	for _, r := range val {
		assert.NotContains(t, charset, r)
	}
}
