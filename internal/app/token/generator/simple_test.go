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

func TestNewSimple(t *testing.T) {
	tests := []struct {
		name string
		want *Simple
	}{
		{
			name: "success",
			want: &Simple{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewSimple())
		})
	}
}

func BenchmarkGenerate(b *testing.B) {
	sg := &Simple{}
	for i := 0; i < b.N; i++ {
		_, _ = sg.Generate()
	}
}
