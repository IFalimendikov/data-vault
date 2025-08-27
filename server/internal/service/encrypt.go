package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// encrypt encrypts a string using AES-GCM encryption
func (s *Vault) encrypt(ctx context.Context, secret string) (string, error) {
	secretByte := []byte(secret)
	key := []byte(s.cfg.EncryptionKey)

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	res := gcm.Seal(nonce, nonce, secretByte, nil)

	return string(res), nil
}

// encryptBytes encrypts byte data using AES-GCM encryption
func (s *Vault) encryptBytes(ctx context.Context, data []byte) ([]byte, error) {
	key := []byte(s.cfg.EncryptionKey)

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	res := gcm.Seal(nonce, nonce, data, nil)

	return res, nil
}
