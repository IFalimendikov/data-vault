package services

import (
	"context"
	"data-vault/client/internal/models"
)

// DeleteURLs processes a batch of URLs for deletion for a specific user
func (v *Vault) Register(ctx context.Context, user models.User) (string, error) {
	jwt, err := v.grpcclient.Register(ctx, user)
	if err != nil {
		return "", err
	}

	return jwt, nil
}
