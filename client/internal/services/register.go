package services

import (
	"context"
	"data-vault/client/internal/models"
)

// Register creates a new user account and returns a JWT token
func (v *Vault) Register(ctx context.Context, user models.User) (string, error) {
	jwt, err := v.grpcclient.Register(ctx, user)
	if err != nil {
		return "", err
	}

	return jwt, nil
}
