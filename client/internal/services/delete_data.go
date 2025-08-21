package services

import (
	"context"
)

// DeleteURLs processes a batch of URLs for deletion for a specific user
func (v *Vault) DeleteData(ctx context.Context, jwt, id string) error {
	return v.grpcclient.DeleteData(ctx, jwt, id)
}
