package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
)

// decryptBytes decrypts byte data using AES-GCM decryption
func (s *Vault) decryptBytes(ctx context.Context, ciphertext []byte) ([]byte, error) {
	key := []byte(s.cfg.EncryptionKey)

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
