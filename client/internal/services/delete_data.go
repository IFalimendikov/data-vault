package services

import (
	"context"
)

// DeleteData removes a specific data entry from the vault
func (v *Vault) DeleteData(ctx context.Context, jwt, id string) error {
	return v.grpcclient.DeleteData(ctx, jwt, id)
}
