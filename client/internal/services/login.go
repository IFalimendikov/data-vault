package services

import (
	"context"
	"data-vault/client/internal/models"
)

// Login authenticates a user and returns a JWT token
func (v *Vault) Login(ctx context.Context, user models.User) (string, error) {
	jwt, err := v.grpcclient.Login(ctx, user)
	if err != nil {
		return "", err
	}

	return jwt, nil
}
