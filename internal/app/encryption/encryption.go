package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
)

type Encryption struct {
	CipherPass string
}

func NewEncryption(cipherPass string) *Encryption {
	return &Encryption{CipherPass: cipherPass}
}

func (e *Encryption) Encrypt(data string) (string, error) {
	aesgcm, nonce, err := e.getAesgcm()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(data), nil)), nil
}

func (e *Encryption) Decrypt(encrypted string) (string, error) {
	aesgcm, nonce, err := e.getAesgcm()
	if err != nil {
		return "", err
	}

	decoded, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	decrypted, err := aesgcm.Open(nil, nonce, decoded, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func (e *Encryption) getAesgcm() (cipher.AEAD, []byte, error) {
	key := sha256.Sum256([]byte(e.CipherPass))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, nil, err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	return aesgcm, nonce, nil
}
