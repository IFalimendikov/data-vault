package services

import (
	"context"
)

// PostData stores encrypted data in the vault
func (v *Vault) PostData(ctx context.Context, jwt, data string) error {
	return v.grpcclient.PostData(ctx, jwt, data)
}
