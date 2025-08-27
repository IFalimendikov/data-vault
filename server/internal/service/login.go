package service

import (
	"context"
	"data-vault/server/internal/models"
)

// Login authenticates a user with encrypted password verification
func (s *Vault) Login(ctx context.Context, user models.User) error {
	if user.Login == "" || user.Password == "" {
		return ErrMalformedRequest
	}

	cipherPassword, err := s.encrypt(ctx, user.Password)
	if err != nil {
		return err
	}
	user.Password = cipherPassword

	err = s.Storage.Login(ctx, user)
	if err != nil {
		return err
	}
	return nil
}
