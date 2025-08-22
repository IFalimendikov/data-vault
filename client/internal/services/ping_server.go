package services

import "context"

// PingServer checks if the server connection is alive and returns true if successful
func (v *Vault) PingServer(ctx context.Context) bool {
	return v.grpcclient.PingServer(ctx)
}
