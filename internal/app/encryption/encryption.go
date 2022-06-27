package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
)

const CipherPass = "J53RPX6"

func Encrypt(data string) (string, error) {
	aesgcm, nonce, err := getAesgcm()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(data), nil)), nil
}

func Decrypt(encrypted string) (string, error) {
	aesgcm, nonce, err := getAesgcm()
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

func getAesgcm() (cipher.AEAD, []byte, error) {
	key := sha256.Sum256([]byte(CipherPass))

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
