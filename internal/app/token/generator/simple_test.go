package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	sg := &SimpleGenerator{}
	val := sg.Generate()
	assert.Len(t, val, Length)
	for _, r := range val {
		assert.NotContains(t, charset, r)
	}
}
