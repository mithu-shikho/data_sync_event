package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// encryptor seals/opens connection URIs with AES-256-GCM. A nil encryptor
// means encryption is disabled (plaintext passthrough — dev only).
type encryptor struct {
	gcm cipher.AEAD
}

func newEncryptor(key []byte) (*encryptor, error) {
	if len(key) == 0 {
		return nil, nil // encryption disabled
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return &encryptor{gcm: gcm}, nil
}

// seal encrypts plaintext, returning nonce||ciphertext.
func (e *encryptor) seal(plaintext string) ([]byte, error) {
	if e == nil {
		return []byte(plaintext), nil
	}
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return e.gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

// open reverses seal.
func (e *encryptor) open(data []byte) (string, error) {
	if e == nil {
		return string(data), nil
	}
	ns := e.gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := data[:ns], data[ns:]
	pt, err := e.gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(pt), nil
}
