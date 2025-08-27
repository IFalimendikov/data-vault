package service

import (
	"context"
)

// PostData encrypts and stores user data in the vault
func (s *Vault) PostData(ctx context.Context, login, dataType string, data []byte) error {
	if login == "" || len(data) == 0 || dataType == "" {
		return ErrMalformedRequest
	}

	cipherData, err := s.encryptBytes(ctx, data)
	if err != nil {
		return err
	}

	err = s.Storage.PostData(ctx, login, dataType, cipherData)
	if err != nil {
		return err
	}
	return nil
}
