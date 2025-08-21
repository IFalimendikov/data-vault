package services

import (
	"context"
)

// SaveURL creates a shortened URL from the original URL and stores it with the associated userID
func (v *Vault) PostData(ctx context.Context, jwt, data string) error {
	return v.grpcclient.PostData(ctx, jwt, data)
}
