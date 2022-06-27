package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			data:    "раз два три",
			want:    "d54f7a94af13b05cf383b5715e8b45d91dbe2ce8588d464304f67eddb237e752b2540951",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ecrypted, err := Encrypt(tt.data)
			assert.NotEqual(t, ecrypted, tt.data)
			assert.Equal(t, tt.want, ecrypted)

			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			data:    "d54f7a94af13b05cf383b5715e8b45d91dbe2ce8588d464304f67eddb237e752b2540951",
			want:    "раз два три",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := Decrypt(tt.data)
			assert.NotEqual(t, decrypted, tt.data)
			assert.Equal(t, tt.want, decrypted)

			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}
