package service

import (
	"context"
	"data-vault/server/internal/models"
	"database/sql"
)

// Register creates a new user account with encrypted password
func (s *Vault) Register(ctx context.Context, user models.User) error {
	tx, err := s.Storage.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if user.Login == "" || user.Password == "" {
		return ErrMalformedRequest
	}

	cipherPassword, err := s.encrypt(ctx, user.Password)
	if err != nil {
		return err
	}
	user.Password = cipherPassword

	err = s.Storage.Register(ctx, tx, user)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
