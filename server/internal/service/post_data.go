package service

import (
	"context"
)

func (s *Vault) PostData(ctx context.Context, login, data string) error {
	if login == "" || data == "" {
		return ErrMalformedRequest
	}

	cipherData, err := s.encrypt(ctx, data)
	if err != nil {
		return err
	}
	data = cipherData

	err = s.Storage.PostData(ctx, login, data)
	if err != nil {
		return err
	}
	return nil
}
