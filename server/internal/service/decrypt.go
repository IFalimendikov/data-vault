package service

import (
	"context"
	"crypto/aes"
    "crypto/cipher"
)

func (s *Vault) decrypt(ctx context.Context, secret string) (string, error) {
    key := []byte(s.cfg.EncryptionKey)
    ciphertext := []byte(secret)

    c, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(c)
    if err != nil {
        return "", err
    }

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return "", err
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
	return string(plaintext), nil
}
