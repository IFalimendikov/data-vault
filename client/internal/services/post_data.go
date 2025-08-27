package services

import (
	"context"
)

// PostData stores encrypted data in the vault
func (v *Vault) PostData(ctx context.Context, jwt, dataType string, data []byte) error {
	return v.grpcclient.PostData(ctx, jwt, dataType, data)
}
