package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	sg := &Simple{}
	val, _ := sg.Generate()
	assert.Len(t, val, Length)
	for _, r := range val {
		assert.NotContains(t, charset, r)
	}
}
